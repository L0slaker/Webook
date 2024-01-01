package web

import (
	"Prove/webook/api/proto/gen/inter/v1"
	"Prove/webook/internal/domain"
	"Prove/webook/internal/errs"
	"Prove/webook/internal/service"
	ijwt "Prove/webook/internal/web/jwt"
	"Prove/webook/pkg/ginx"
	"Prove/webook/pkg/logger"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc service.ArticleService
	l   logger.LoggerV1
	//interSvc service2.InteractiveService
	interSvc interv1.InteractiveServiceClient
	biz      string
}

func NewArticleHandler(svc service.ArticleService, l logger.LoggerV1,
	interSvc interv1.InteractiveServiceClient) *ArticleHandler {
	return &ArticleHandler{
		svc:      svc,
		l:        l,
		biz:      "article",
		interSvc: interSvc,
	}
}

func (a *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 登录校验,拿到author id
	c := ctx.MustGet("claims")
	claim, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: errs.UserUnauthorizedError,
			Msg:  "未授权!",
		})
		a.l.Error("未发现用户的 session 信息！")
		return
	}

	// 调用 svc 的代码
	id, err := a.svc.Save(ctx, req.toDomain(claim.UserId))
	fmt.Println(err)
	fmt.Println(err)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误！",
		})
		a.l.Error("保存帖子失败！", logger.Error(err))
		return
	}

	// 返回响应
	ctx.JSON(http.StatusOK, Result{
		Msg:  "保存成功！",
		Data: id,
	})
}

func (a *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 登录校验,拿到author id
	c := ctx.MustGet("claims")
	claim, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: errs.UserUnauthorizedError,
			Msg:  "未授权!",
		})
		a.l.Error("未发现用户的 session 信息！")
		return
	}

	id, err := a.svc.Publish(ctx, req.toDomain(claim.UserId))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误！",
		})
		a.l.Error("发表帖子失败！", logger.Error(err))
		return
	}

	// 返回响应
	ctx.JSON(http.StatusOK, Result{
		Msg:  "发布成功！",
		Data: id,
	})
}

// Withdraw 将制作库和线上库的文章都设置为仅自己可见
func (a *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 登录校验,拿到author id
	c := ctx.MustGet("claims")
	claim, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: errs.UserUnauthorizedError,
			Msg:  "未授权!",
		})
		a.l.Error("未发现用户的 session 信息！")
		return
	}

	err := a.svc.Withdraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: claim.UserId,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误！",
		})
		a.l.Error("撤回帖子失败！", logger.Error(err))
		return
	}

	// 返回响应
	ctx.JSON(http.StatusOK, Result{
		Msg: "撤回帖子成功！",
	})
}

func (a *ArticleHandler) List(ctx *gin.Context, req ListReq, uc ijwt.UserClaims) (ginx.Result, error) {
	res, err := a.svc.List(ctx, uc.UserId, req.Offset, req.Limit)
	if err != nil {
		return Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误！",
		}, fmt.Errorf("获取文章列表出错 %w！", err)
	}
	// 列表页不展示全文，而是显示一个摘要
	// 简单摘要可以是文章的前几句话；强大的摘要是 AI 生成的
	return ginx.Result{
		Data: slice.Map[domain.Article, ArticleVO](res, func(idx int, src domain.Article) ArticleVO {
			return ArticleVO{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),
				// 列表请求不需要返回内容
				//Content:    src.Content,
				// 创作者自己的文章列表，也无需展示作者字段
				//Author:     src.Author,
				Status:     src.Status.ToUint8(),
				CreateTime: src.CreateTime.Format(time.RFC1123),
				UpdateTime: src.UpdateTime.Format(time.RFC1123),
			}
		}),
	}, nil
}

func (a *ArticleHandler) Detail(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 4)
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInvalidInput,
			Msg:  "参数错误！",
		}, fmt.Errorf("参数错误 %w！", err)
	}
	art, err := a.svc.GetById(ctx, id)
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误！",
		}, fmt.Errorf("查看文章详情失败 %w！", err)
	}
	// 判定
	if art.Author.Id != uc.UserId {
		return ginx.Result{
			Code: errs.ArticleInvalidInput,
			Msg:  "输入有误！",
		}, fmt.Errorf("非法访问文章，作者 ID 不匹配 %d！", uc.UserId)
	}
	return Result{
		Data: ArticleVO{
			Id:    art.Id,
			Title: art.Title,
			// 文章详情不用显示摘要
			//Abstract:   art.Abstract(),
			Content: art.Content,
			// 创作者自己的文章列表，也无需展示作者字段
			//Author:     art.Author,
			Status:     art.Status.ToUint8(),
			CreateTime: art.CreateTime.Format(time.RFC1123),
			UpdateTime: art.UpdateTime.Format(time.RFC1123),
		},
	}, nil
}

func (a *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 4)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: errs.ArticleInvalidInput,
			Msg:  "输入有误！",
		})
		a.l.Error("前端输入的 Id 不对", logger.Error(err))
		return
	}

	var (
		//inter domain2.Interactive
		resp *interv1.GetResponse
		eg   errgroup.Group
		art  domain.Article
	)
	uc := ctx.MustGet("users").(ijwt.UserClaims)

	// 读取文章
	eg.Go(func() error {
		art, err = a.svc.GetPublishedById(ctx, id, uc.UserId)
		return err
	})

	// 获取文章的计数
	eg.Go(func() error {
		resp, err = a.interSvc.Get(ctx, &interv1.GetRequest{
			Biz:   a.biz,
			BizId: id,
			Uid:   uc.UserId,
		})
		return err
	})

	if err = eg.Wait(); err != nil {
		// 查询出错
		ctx.JSON(http.StatusOK, Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误！",
		})
		return
	}

	// 增加阅读计数，异步执行即可
	go func() {
		_, cntErr := a.interSvc.IncrReadCnt(ctx, &interv1.IncrReadCntRequest{
			Biz:   a.biz,
			BizId: art.Id,
		})
		if cntErr != nil {
			a.l.Error("增加阅读计数失败！", logger.Int64("author_id", art.Id), logger.Error(cntErr))
		}
	}()

	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVO{
			Id:         art.Id,
			Title:      art.Title,
			Content:    art.Content,
			Author:     art.Author.Name,
			ReadCnt:    resp.Inter.ReadCnt,
			LikeCnt:    resp.Inter.LikeCnt,
			CollectCnt: resp.Inter.CollectCnt,
			Liked:      resp.Inter.Liked,
			Collected:  resp.Inter.Collected,
			Status:     art.Status.ToUint8(),
			CreateTime: art.CreateTime.Format(time.RFC1123),
			UpdateTime: art.UpdateTime.Format(time.RFC1123),
		},
	})
}

func (a *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc ijwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		_, err = a.interSvc.Like(ctx, &interv1.LikeRequest{
			Biz:   a.biz,
			BizId: req.Id,
			Uid:   uc.UserId,
		})
	} else {
		_, err = a.interSvc.CancelLike(ctx, &interv1.CancelLikeRequest{
			Biz:   a.biz,
			BizId: req.Id,
			Uid:   uc.UserId,
		})
	}
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误！",
		}, err
	}
	return ginx.Result{Msg: "OK！"}, nil
}

func (a *ArticleHandler) Collect(ctx *gin.Context, req CollectReq, uc ijwt.UserClaims) (ginx.Result, error) {
	_, err := a.interSvc.Collect(ctx, &interv1.CollectRequest{
		Biz:   a.biz,
		BizId: req.Id,
		Cid:   req.Cid,
		Uid:   uc.UserId,
	})
	if err != nil {
		return Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误！",
		}, err
	}
	return Result{Msg: "OK"}, nil
}
