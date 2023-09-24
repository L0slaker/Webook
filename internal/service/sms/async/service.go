package async

//type SMSService struct {
//	svc sms.Service
//	//repo repository.SMSAsyncReqRepository
//}
//
//func (s *SMSService)StartAsync(){
//	go func() {
//		// 查找没发出去的请求
//		reqs := s.repo.Find()
//		for _, req := range reqs {
//			// 在这里发送，并且控制重试
//			s.svc.Send(,req.biz,req.args,req.numbers...)
//		}
//
//	}()
//}
//
//func (s *SMSService)Send(ctx context.Context,biz string,args []string,numbers...string) error {
//	// 首先是正常路径
//	err := s.svc.Send(ctx,biz,args,numbers...)
//	if err != nil {
//		// 判定是否崩溃
//
//		// 如果崩溃了...
//		if isDown {
//			s.repo.Store()
//		}
//	}
//	return
//}
