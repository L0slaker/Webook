package domain

import "time"

const (
	ArticleStatusUnknown     ArticleStatus = iota // 未知状态
	ArticleStatusUnPublished                      // 未发表
	ArticleStatusPublished                        // 已发表
	ArticleStatusPrivate                          // 仅自己可见
)

type Article struct {
	Id         int64
	Title      string
	Content    string
	Author     Author
	Status     ArticleStatus
	CreateTime time.Time
	UpdateTime time.Time
}

type Author struct {
	Id   int64
	Name string
}

type ArticleStatus uint8

func (status ArticleStatus) ToUint8() uint8 {
	return uint8(status)
}

// Valid 检查状态是否合法
func (status ArticleStatus) Valid() bool {
	return status.ToUint8() > 0
}

func (status ArticleStatus) NonPublished() bool {
	return status != ArticleStatusPrivate
}

func (a Article) Abstract() string {
	// 取文章内容的前几句作为摘要，小心中文问题
	cs := []rune(a.Content)
	if len(cs) < 100 {
		return a.Content
	}
	return string(cs[:100])
}

func (status ArticleStatus) String() string {
	switch status {
	case ArticleStatusUnPublished:
		return "unpublished"
	case ArticleStatusPublished:
		return "published"
	case ArticleStatusPrivate:
		return "private"
	default:
		return "unknown"
	}
}
