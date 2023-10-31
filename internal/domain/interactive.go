package domain

// Interactive 总体交互的计数
type Interactive struct {
	ReadCnt    int64 `json:"read_cnt"`
	LikeCnt    int64 `json:"like_cnt"`
	CollectCnt int64 `json:"collect_cnt"`
	Liked      bool  `json:"liked"`     // 是否有点赞或收集
	Collected  bool  `json:"collected"` //是否有收集
}
