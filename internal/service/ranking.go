package service

import (
	interv1 "Prove/webook/api/proto/gen/inter/v1"
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository"
	"context"
	"errors"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
)

type RankingService interface {
	TopN(ctx context.Context) error
}

// BatchRankingService 既操作文章，也操作计数，作为聚合服务来使用
type BatchRankingService struct {
	artSvc ArticleService
	//interSvc  service.InteractiveService
	interSvc  interv1.InteractiveServiceClient
	repo      repository.RankingRepository
	batchSize int // 批量大小
	n         int //排行榜数量
	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(artSvc ArticleService, repo repository.RankingRepository, interSvc interv1.InteractiveServiceClient) RankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		interSvc:  interSvc,
		batchSize: 100,
		n:         100,
		repo:      repo,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			sec := time.Since(t).Seconds()
			return float64(likeCnt-1) / math.Pow(sec+2, 1.5)
		},
	}
}

func (svc *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := svc.topN(ctx)
	if err != nil {
		return err
	}
	return svc.repo.ReplaceTopN(ctx, arts)
}

// 测试专用，由于 TopN 方法没有返回值，不利于测试，所以用此方法进行测试
func (svc *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	now := time.Now()
	offset := 0
	type Score struct {
		art   domain.Article
		score float64
	}
	// 为什么需要一个优先级队列？用来存放排行榜吗？
	topN := queue.NewConcurrentPriorityQueue[Score](svc.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})
	for {
		// 拉取一批数据
		arts, err := svc.artSvc.ListPub(ctx, now, offset, svc.batchSize)
		if err != nil {
			return nil, err
		}
		// 找到对应的点赞数据
		ids := slice.Map[domain.Article, int64](arts,
			func(idx int, src domain.Article) int64 {
				return src.Id
			})
		inters, err := svc.interSvc.GetByIds(ctx, &interv1.GetByIdsRequest{
			Biz:    "article",
			BizIds: ids,
		})

		if err != nil {
			return nil, err
		}
		if len(inters.Inters) == 0 {
			return nil, errors.New("没有数据")
		}

		// 计算结果
		for _, art := range arts {
			inter := inters.Inters[art.Id]
			score := svc.scoreFunc(art.UpdateTime, inter.LikeCnt)
			// 这里需要注意，如果榜单没有满，那么替换操作毫无意义
			err = topN.Enqueue(Score{
				art:   art,
				score: score,
			})
			// 此时榜单已满，考虑替换数据。按照热度来排序的小顶堆，最顶部的数据一定是热度最低的
			if err == errors.New("ekit: 超出最大容量限制") {
				val, _ := topN.Dequeue()
				// 判断 score 是否在前n名中
				if val.score < score {
					err = topN.Enqueue(Score{
						art:   art,
						score: score,
					})
				} else {
					// 如果榜单没满，直接插入
					_ = topN.Enqueue(val)
				}
			}
		}
		// 处理完一批之后，怎么判断是否还要继续处理下一批？或者说知道还有没有？
		// 1.一批都没有取够，肯定余量不足够下一批了
		// 2.已经取到七天之前的数据
		if len(arts) < svc.batchSize ||
			now.Sub(arts[len(arts)-1].UpdateTime).Hours() > 7*24 {
			break
		}
		// 更新 offset
		offset += len(arts)
	}
	// 得到结果
	res := make([]domain.Article, svc.n)
	// 最先出队的优先级是最小的，所以返回的结果需要倒序输出
	for i := svc.n - 1; i >= 0; i-- {
		val, err := topN.Dequeue()
		if err != nil {
			break // 取完了
		}
		res[i] = val.art
	}
	return res, nil
}

// 1.拉取一批文章
// 2.找出每个文章的对应的点赞数
// 3.计算结果
// 4.是否加入优先级队列
//		优先级队列未满：直接入队
//		优先级队列已满：将顶部出队
// 5.继续处理批次
//		如果处理完一批没有下一批的余量了，也就是这一批次没有取够，就认为是取完了
//		已经取到七天之前的数据
// 6.将优先级队列的结果倒序输出到返回的结果集
