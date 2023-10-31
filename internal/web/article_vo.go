package web

import (
	"Prove/webook/internal/domain"
)

// VO -> view object，也就是前端传入的参数

type ArticleVO struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
	// 摘要
	Abstract   string `json:"abstract"`
	Content    string `json:"content"`
	Author     string `json:"author"`
	ReadCnt    int64  `json:"read_cnt"`
	LikeCnt    int64  `json:"like_cnt"`
	CollectCnt int64  `json:"collect_cnt"`
	// 个人是否点赞或收藏
	Liked      bool   `json:"liked"`
	Collected  bool   `json:"collected"`
	Status     uint8  `json:"status"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

type ListReq struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	}
}

type LikeReq struct {
	Id   int64 `json:"id"`
	Like bool  `json:"like"`
}

type CollectReq struct {
	Id  int64 `json:"id"`
	Cid int64 `json:"cid"`
}
