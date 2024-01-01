package integration

import (
	interv1 "Prove/webook/api/proto/gen/inter/v1"
	"Prove/webook/interactive/integration/startup"
	"Prove/webook/interactive/repository/dao"
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
	"time"
)

type InteractiveTestSuite struct {
	suite.Suite
	db  *gorm.DB
	rdb redis.Cmdable
}

func (s *InteractiveTestSuite) SetupSuite() {
	s.db = startup.InitTestDB()
	s.rdb = startup.InitRedis()
}

func (s *InteractiveTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	// 清空 MySQL
	err := s.db.Exec("TRUNCATE TABLE `interactives`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `user_like_bizs`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `user_collection_bizs`").Error
	assert.NoError(s.T(), err)
	// 清空 Redis
	err = s.rdb.FlushDB(ctx).Err()
	assert.NoError(s.T(), err)
}

func (s *InteractiveTestSuite) TestIncrReadCnt() {
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		biz      string
		bizId    int64
		wantResp *interv1.IncrReadCntResponse
		wantErr  error
	}{
		{
			name: "增加成功，缓存和数据库中都有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         1,
					BizId:      2,
					Biz:        "test",
					ReadCnt:    3,
					LikeCnt:    4,
					CollectCnt: 5,
					CreateTime: 123,
					UpdateTime: 456,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:2", "read_cnt", 3).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 1).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.UpdateTime > 456)
				data.UpdateTime = 0
				assert.Equal(t, dao.Interactive{
					Id:         1,
					BizId:      2,
					Biz:        "test",
					ReadCnt:    4,
					LikeCnt:    4,
					CollectCnt: 5,
					CreateTime: 123,
				}, data)
				cnt, err := s.rdb.HGet(ctx, "interactive:test:2", "read_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 4, cnt)
				err = s.rdb.Del(ctx, "interactive:test:2").Err()
				assert.NoError(t, err)
			},
			biz:      "test",
			bizId:    2,
			wantResp: &interv1.IncrReadCntResponse{},
			wantErr:  nil,
		},
		{
			name: "增加成功，缓存没有数据，数据库中有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         2,
					Biz:        "test",
					BizId:      3,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					CreateTime: 6,
					UpdateTime: 7,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 2).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.UpdateTime > 7)
				data.UpdateTime = 0
				assert.Equal(t, dao.Interactive{
					Id:         2,
					Biz:        "test",
					BizId:      3,
					ReadCnt:    4,
					CollectCnt: 4,
					LikeCnt:    5,
					CreateTime: 6,
				}, data)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:3").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},
			biz:      "test",
			bizId:    3,
			wantResp: &interv1.IncrReadCntResponse{},
		},
		{
			name:   "增加成功，但缓存和数据库中都没有数据",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("biz_id = ? AND biz = ?", 4, "test").First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.CreateTime > 0)
				assert.True(t, data.UpdateTime > 0)
				assert.True(t, data.Id > 0)
				data.CreateTime = 0
				data.UpdateTime = 0
				data.Id = 0
				assert.Equal(t, dao.Interactive{
					Biz:     "test",
					BizId:   4,
					ReadCnt: 1,
				}, data)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:4").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},
			biz:      "test",
			bizId:    4,
			wantResp: &interv1.IncrReadCntResponse{},
		},
	}
	// 不同于 AsyncSms，不需要mock，所以创建一个svc就足够
	svc := startup.InitInteractiveGRPCServer()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			resp, err := svc.IncrReadCnt(context.Background(),
				&interv1.IncrReadCntRequest{
					Biz:   tc.biz,
					BizId: tc.bizId,
				})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
			tc.after(t)
		})
	}
}

func (s *InteractiveTestSuite) TestLike() {
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		uid      int64
		biz      string
		bizId    int64
		wantResp *interv1.LikeResponse
		wantErr  error
	}{
		{
			name: "点赞，数据库和缓存中都有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				err := s.db.Create(dao.Interactive{
					Id:         4,
					BizId:      5,
					Biz:        "test",
					ReadCnt:    3,
					LikeCnt:    4,
					CollectCnt: 5,
					CreateTime: 6,
					UpdateTime: 7,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:5", "like_cnt", 4).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 4).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.UpdateTime > 7)
				data.UpdateTime = 0
				assert.Equal(t, dao.Interactive{
					Id:         4,
					BizId:      5,
					Biz:        "test",
					ReadCnt:    3,
					LikeCnt:    5,
					CollectCnt: 5,
					CreateTime: 6,
				}, data)

				var likeBiz dao.UserLikeBiz
				err = s.db.Where("biz = ? AND biz_id = ? AND uid = ?",
					"test", 5, 1).First(&likeBiz).Error
				assert.NoError(t, err)
				assert.True(t, likeBiz.Id > 0)
				assert.True(t, likeBiz.CreateTime > 0)
				assert.True(t, likeBiz.UpdateTime > 0)
				likeBiz.Id = 0
				likeBiz.CreateTime = 0
				likeBiz.UpdateTime = 0
				assert.Equal(t, dao.UserLikeBiz{
					Uid:    1,
					BizId:  5,
					Biz:    "test",
					Status: 1,
				}, likeBiz)

				cnt, err := s.rdb.HGet(ctx, "interactive:test:5", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 5, cnt)
				err = s.rdb.Del(ctx, "interactive:test:5").Err()
				assert.NoError(t, err)
			},
			uid:      1,
			biz:      "test",
			bizId:    5,
			wantResp: &interv1.LikeResponse{},
		},
		{
			name:   "点赞，数据库和缓存中都没数据",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("biz = ? AND biz_id = ?", "test", 6).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.UpdateTime > 0)
				assert.True(t, data.CreateTime > 0)
				assert.True(t, data.Id > 0)
				data.UpdateTime = 0
				data.CreateTime = 0
				data.Id = 0
				assert.Equal(t, dao.Interactive{
					BizId:   6,
					Biz:     "test",
					LikeCnt: 1,
				}, data)

				var likeBiz dao.UserLikeBiz
				err = s.db.Where("biz = ? AND biz_id = ? AND uid = ?",
					"test", 6, 2).First(&likeBiz).Error
				assert.NoError(t, err)
				assert.True(t, likeBiz.Id > 0)
				assert.True(t, likeBiz.CreateTime > 0)
				assert.True(t, likeBiz.UpdateTime > 0)
				likeBiz.Id = 0
				likeBiz.CreateTime = 0
				likeBiz.UpdateTime = 0
				assert.Equal(t, dao.UserLikeBiz{
					Uid:    2,
					BizId:  6,
					Biz:    "test",
					Status: 1,
				}, likeBiz)

				cnt, err := s.rdb.Exists(ctx, "interactive:test:6").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},
			uid:      2,
			biz:      "test",
			bizId:    6,
			wantResp: &interv1.LikeResponse{},
		},
	}
	svc := startup.InitInteractiveGRPCServer()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			resp, err := svc.Like(context.Background(), &interv1.LikeRequest{
				Biz:   tc.biz,
				BizId: tc.bizId,
				Uid:   tc.uid,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
			tc.after(t)
		})
	}
}

func (s *InteractiveTestSuite) TestDislike() {
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		biz      string
		bizId    int64
		uid      int64
		wantResp *interv1.CancelLikeResponse
		wantErr  error
	}{
		{
			name: "取消点赞，缓存和数据库都有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				err := s.db.Create(dao.Interactive{
					Id:         6,
					BizId:      7,
					Biz:        "test",
					ReadCnt:    3,
					LikeCnt:    4,
					CollectCnt: 5,
					CreateTime: 6,
					UpdateTime: 7,
				}).Error
				assert.NoError(t, err)
				err = s.db.Create(dao.UserLikeBiz{
					Id:         6,
					Uid:        3,
					BizId:      7,
					Biz:        "test",
					CreateTime: 6,
					UpdateTime: 7,
					Status:     1,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:7", "like_cnt", 3).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 6).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.UpdateTime > 7)
				data.UpdateTime = 0
				assert.Equal(t, dao.Interactive{
					Id:         6,
					BizId:      7,
					Biz:        "test",
					ReadCnt:    3,
					LikeCnt:    3,
					CollectCnt: 5,
					CreateTime: 6,
				}, data)

				var likeBiz dao.UserLikeBiz
				err = s.db.Where("id = ?", 6).First(&likeBiz).Error
				assert.NoError(t, err)
				assert.True(t, likeBiz.UpdateTime > 7)
				likeBiz.UpdateTime = 0
				assert.Equal(t, dao.UserLikeBiz{
					Id:         6,
					Uid:        3,
					BizId:      7,
					Biz:        "test",
					CreateTime: 6,
				}, likeBiz)
				cnt, err := s.rdb.HGet(ctx, "interactive:test:7", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 2, cnt)
				err = s.rdb.Del(ctx, "interactive:test:7").Err()
				assert.NoError(t, err)
			},
			biz:      "test",
			bizId:    7,
			uid:      3,
			wantResp: &interv1.CancelLikeResponse{},
		},
	}

	svc := startup.InitInteractiveGRPCServer()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			resp, err := svc.CancelLike(context.Background(), &interv1.CancelLikeRequest{
				Biz:   tc.biz,
				BizId: tc.bizId,
				Uid:   tc.uid,
			})
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResp, resp)
			tc.after(t)
		})
	}
}

func (s *InteractiveTestSuite) TestCollect() {
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		biz      string
		bizId    int64
		uid      int64
		cid      int64
		wantResp *interv1.CollectResponse
	}{
		{
			name:   "收藏成功，数据库和缓存都没数据",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				var inter dao.Interactive
				err := s.db.Where("biz = ? AND biz_id = ?", "test", 8).First(&inter).Error
				assert.NoError(t, err)
				assert.True(t, inter.CreateTime > 0)
				assert.True(t, inter.UpdateTime > 0)
				assert.True(t, inter.Id > 0)
				inter.CreateTime = 0
				inter.UpdateTime = 0
				inter.Id = 0
				assert.Equal(t, dao.Interactive{
					BizId:      8,
					Biz:        "test",
					CollectCnt: 1,
				}, inter)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:8").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
				// 收藏记录
				var ucb dao.UserCollectionBiz
				err = s.db.Where("uid = ? AND biz = ? AND biz_id = ?", 4, "test", 8).First(&ucb).Error
				assert.NoError(t, err)
				assert.True(t, ucb.CreateTime > 0)
				assert.True(t, ucb.UpdateTime > 0)
				assert.True(t, ucb.Id > 0)
				ucb.CreateTime = 0
				ucb.UpdateTime = 0
				ucb.Id = 0
				assert.Equal(t, dao.UserCollectionBiz{
					Uid:   4,
					Cid:   1,
					BizId: 8,
					Biz:   "test",
				}, ucb)
			},
			biz:      "test",
			bizId:    8,
			uid:      4,
			cid:      1,
			wantResp: &interv1.CollectResponse{},
		},
		{
			name: "收藏成功，数据库和缓存都有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         8,
					BizId:      9,
					Biz:        "test",
					CollectCnt: 6,
					CreateTime: 7,
					UpdateTime: 8,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:9", "collect_cnt", 6).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var intr dao.Interactive
				err := s.db.WithContext(ctx).
					Where("biz = ? AND biz_id = ?", "test", 9).First(&intr).Error
				assert.NoError(t, err)
				assert.True(t, intr.CreateTime > 0)
				intr.CreateTime = 0
				assert.True(t, intr.UpdateTime > 0)
				intr.UpdateTime = 0
				assert.True(t, intr.Id > 0)
				intr.Id = 0
				assert.Equal(t, dao.Interactive{
					Biz:        "test",
					BizId:      9,
					CollectCnt: 7,
				}, intr)
				cnt, err := s.rdb.HGet(ctx, "interactive:test:9", "collect_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 7, cnt)

				var cbiz dao.UserCollectionBiz
				err = s.db.WithContext(ctx).
					Where("uid = ? AND biz = ? AND biz_id = ?", 5, "test", 9).
					First(&cbiz).Error
				assert.NoError(t, err)
				assert.True(t, cbiz.CreateTime > 0)
				cbiz.CreateTime = 0
				assert.True(t, cbiz.UpdateTime > 0)
				cbiz.UpdateTime = 0
				assert.True(t, cbiz.Id > 0)
				cbiz.Id = 0
				assert.Equal(t, dao.UserCollectionBiz{
					Biz:   "test",
					BizId: 9,
					Cid:   2,
					Uid:   5,
				}, cbiz)
			},
			bizId:    9,
			biz:      "test",
			cid:      2,
			uid:      5,
			wantResp: &interv1.CollectResponse{},
		},
		{
			name: "收藏成功，数据库有数据,缓存没数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         9,
					BizId:      10,
					Biz:        "test",
					CollectCnt: 7,
					CreateTime: 8,
					UpdateTime: 9,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				var inter dao.Interactive
				err := s.db.Where("biz = ? AND biz_id = ?", "test", 10).First(&inter).Error
				assert.NoError(t, err)
				assert.True(t, inter.CreateTime > 0)
				assert.True(t, inter.UpdateTime > 0)
				assert.True(t, inter.Id > 0)
				inter.CreateTime = 0
				inter.UpdateTime = 0
				inter.Id = 0
				assert.Equal(t, dao.Interactive{
					BizId:      10,
					Biz:        "test",
					CollectCnt: 8,
				}, inter)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:10").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
				// 收藏记录
				var ucb dao.UserCollectionBiz
				err = s.db.Where("uid = ? AND biz = ? AND biz_id = ?", 6, "test", 10).First(&ucb).Error
				assert.NoError(t, err)
				assert.True(t, ucb.CreateTime > 0)
				assert.True(t, ucb.UpdateTime > 0)
				assert.True(t, ucb.Id > 0)
				ucb.CreateTime = 0
				ucb.UpdateTime = 0
				ucb.Id = 0
				assert.Equal(t, dao.UserCollectionBiz{
					Uid:   6,
					Cid:   3,
					BizId: 10,
					Biz:   "test",
				}, ucb)
			},
			biz:      "test",
			bizId:    10,
			uid:      6,
			cid:      3,
			wantResp: &interv1.CollectResponse{},
		},
	}

	svc := startup.InitInteractiveGRPCServer()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			resp, err := svc.Collect(context.Background(), &interv1.CollectRequest{
				Biz:   tc.biz,
				BizId: tc.bizId,
				Uid:   tc.uid,
				Cid:   tc.cid,
			})
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResp, resp)
			tc.after(t)
		})
	}
}

func (s *InteractiveTestSuite) TestGet() {
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		biz      string
		bizId    int64
		uid      int64
		wantResp *interv1.GetResponse
	}{
		{
			name: "没命中缓存，取出了所有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
				defer cancel()
				err := s.db.WithContext(ctx).Create(&dao.Interactive{
					Id:         10,
					BizId:      11,
					Biz:        "test",
					ReadCnt:    200,
					LikeCnt:    300,
					CollectCnt: 400,
					CreateTime: 8,
					UpdateTime: 9,
				}).Error
				assert.NoError(t, err)
			},
			biz:   "test",
			bizId: 11,
			uid:   7,
			wantResp: &interv1.GetResponse{
				Inter: &interv1.Interactive{
					Biz:        "test",
					BizId:      11,
					ReadCnt:    200,
					LikeCnt:    300,
					CollectCnt: 400,
				},
			},
		},
		{
			name:  "全部取出来了-命中缓存-用户已点赞收藏",
			biz:   "test",
			bizId: 12,
			uid:   8,
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				err := s.db.WithContext(ctx).
					Create(&dao.UserCollectionBiz{
						Cid:        5,
						Biz:        "test",
						BizId:      12,
						Uid:        8,
						CreateTime: 123,
						UpdateTime: 124,
					}).Error
				assert.NoError(t, err)
				err = s.db.WithContext(ctx).
					Create(&dao.UserLikeBiz{
						Biz:        "test",
						BizId:      12,
						Uid:        8,
						CreateTime: 123,
						UpdateTime: 124,
						Status:     1,
					}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:12",
					"like_cnt", 0, "collect_cnt", 1).Err()
				assert.NoError(t, err)
			},
			wantResp: &interv1.GetResponse{
				Inter: &interv1.Interactive{
					Biz:        "test",
					BizId:      12,
					LikeCnt:    0,
					CollectCnt: 1,
					Liked:      true,
					Collected:  true,
				},
			},
		},
	}
	svc := startup.InitInteractiveGRPCServer()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			resp, err := svc.Get(context.Background(), &interv1.GetRequest{
				Biz:   tc.biz,
				BizId: tc.bizId,
				Uid:   tc.uid,
			})
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func (s *InteractiveTestSuite) TestGetByIds() {
	preCtx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// 准备数据
	for i := 1; i < 5; i++ {
		i := int64(i)
		err := s.db.WithContext(preCtx).
			Create(&dao.Interactive{
				Id:         i,
				Biz:        "test",
				BizId:      i,
				ReadCnt:    i,
				CollectCnt: i + 1,
				LikeCnt:    i + 2,
			}).Error
		assert.NoError(s.T(), err)
	}

	testCases := []struct {
		name string

		before func(t *testing.T)
		biz    string
		ids    []int64

		wantErr  error
		wantResp *interv1.GetByIdsResponse
	}{
		{
			name: "查找成功",
			biz:  "test",
			ids:  []int64{1, 2},
			wantResp: &interv1.GetByIdsResponse{
				Inters: map[int64]*interv1.Interactive{
					1: {
						Biz:        "test",
						BizId:      1,
						ReadCnt:    1,
						CollectCnt: 2,
						LikeCnt:    3,
					},
					2: {
						Biz:        "test",
						BizId:      2,
						ReadCnt:    2,
						CollectCnt: 3,
						LikeCnt:    4,
					},
				},
			},
		},
		{
			name: "没有对应的数据",
			biz:  "test",
			ids:  []int64{100, 200},
			wantResp: &interv1.GetByIdsResponse{
				Inters: map[int64]*interv1.Interactive{},
			},
		},
	}

	svc := startup.InitInteractiveGRPCServer()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetByIds(context.Background(), &interv1.GetByIdsRequest{
				Biz:    tc.biz,
				BizIds: tc.ids,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestInteractiveService(t *testing.T) {
	suite.Run(t, &InteractiveTestSuite{})
}
