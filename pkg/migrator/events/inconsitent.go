package events

import "context"

type Producer interface {
	ProduceInconsistentEvent(ctx context.Context, evt InconsistentEvent) error
}

type InconsistentEvent struct {
	ID int64
	// 取值 SRC 表示以源表为准；取值 DST 表示以目标表为准
	Direction string // 以 xx 为准来修复数据
	Type      string // 什么引起不一致
}

const (
	InconsistentEventTypeTM  = "target_missing" // 校验的目标数据缺失
	InconsistentEventTypeBM  = "base_missing"   // 校验的基础数据缺失
	InconsistentEventTypeNEQ = "not_equal"      // 不相等
)
