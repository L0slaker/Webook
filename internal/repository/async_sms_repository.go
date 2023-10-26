package repository

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository/dao"
	"context"
	"github.com/ecodeclub/ekit/sqlx"
)

var ErrWaitingSMSNotFound = dao.ErrWaitingSMSNotFound

type AsyncSMSRepository interface {
	Add(ctx context.Context, sms domain.AsyncSms) error
	PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error)
	ReportScheduleResult(ctx context.Context, id int64, success bool) error
}

type asyncSMSRepository struct {
	dao dao.AsyncSmsDAO
}

func NewAsyncSMSRepository(dao dao.AsyncSmsDAO) AsyncSMSRepository {
	return &asyncSMSRepository{
		dao: dao,
	}
}

func (a *asyncSMSRepository) Add(ctx context.Context, s domain.AsyncSms) error {
	return a.dao.Insert(ctx, dao.AsyncSms{
		Config: sqlx.JsonColumn[dao.SmsConfig]{
			Val: dao.SmsConfig{
				TplId:   s.TplId,
				Args:    s.Args,
				Numbers: s.Numbers,
			},
			Valid: true,
		},
		RetryMax: s.RetryMax,
	})
}

func (a *asyncSMSRepository) PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error) {
	as, err := a.dao.GetWaitingSMS(ctx)
	if err != nil {
		return domain.AsyncSms{}, nil
	}
	return domain.AsyncSms{
		Id:       as.Id,
		TplId:    as.Config.Val.TplId,
		Args:     as.Config.Val.Args,
		Numbers:  as.Config.Val.Numbers,
		RetryMax: as.RetryMax,
	}, nil
}

func (a *asyncSMSRepository) ReportScheduleResult(ctx context.Context, id int64, success bool) error {
	if success {
		return a.dao.MarkSuccess(ctx, id)
	}
	return a.dao.MarkFailed(ctx, id)
}
