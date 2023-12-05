package service

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository"
	"Prove/webook/pkg/logger"
	"context"
	"time"
)

type JobService interface {
	// Preempt 抢占
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
	// Release 释放
	//Release(ctx context.Context,id int64) error
}

type CronJobService struct {
	repo            repository.JobRepository
	refreshInterval time.Duration
	l               logger.LoggerV1
}

func (svc *CronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := svc.repo.Preempt(ctx)

	// 定时续约
	ticker := time.NewTicker(svc.refreshInterval)
	go func() {
		for range ticker.C {
			svc.refresh(j.Id)
		}
	}()

	j.CancelFunc = func() error {
		// 在这里释放
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		return svc.repo.Release(ctx, j.Id)
	}
	return j, err
}

func (svc *CronJobService) refresh(id int64) {
	// status = jobStatusRunning，但是update_time在三分钟之前
	// 说明此时没有续约，否则我们会一直更新update_time
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := svc.repo.UpdateUTime(ctx, id)
	if err != nil {
		svc.l.Error("续约失败", logger.Error(err), logger.Int64("jid", id))
	}
}

func (svc *CronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	next := j.NextTime()
	if next.IsZero() {
		// 没有下一次的任务了，不需要更新了
		return svc.repo.Stop(ctx, j.Id)
	}
	return svc.repo.UpdateNextTime(ctx, j.Id, next)
}
