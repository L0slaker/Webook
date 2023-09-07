package retryable

//// Service 注意并发问题
//type Service struct {
//	svc sms.Service
//	// 重试值
//	retryCnt int
//}
//
//func (s Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
//	err := s.svc.Send(ctx, tplId, args, numbers...)
//	// 考虑只重试 10 次
//	for err != nil && s.retryCnt < 10 {
//		err = s.svc.Send(ctx, tplId, args, numbers...)
//		s.retryCnt++
//	}
//}
