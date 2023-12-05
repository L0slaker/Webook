package repository

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository/dao"
	"context"
	"time"
)

type JobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUTime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
	Stop(ctx context.Context, id int64) error
}

type PreemptCronJobRepository struct {
	dao dao.JobDAO
}

func NewPreemptCronJobRepository(dao dao.JobDAO) JobRepository {
	return &PreemptCronJobRepository{
		dao: dao,
	}
}

func (repo *PreemptCronJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := repo.dao.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	return repo.entityToDomain(j), nil
}

func (repo *PreemptCronJobRepository) Release(ctx context.Context, id int64) error {
	return repo.dao.Release(ctx, id)
}

func (repo *PreemptCronJobRepository) UpdateUTime(ctx context.Context, id int64) error {
	return repo.dao.UpdateUTime(ctx, id)
}

func (repo *PreemptCronJobRepository) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return repo.dao.UpdateNextTime(ctx, id, next)
}

func (repo *PreemptCronJobRepository) Stop(ctx context.Context, id int64) error {
	return repo.dao.Stop(ctx, id)
}

func (repo *PreemptCronJobRepository) domainToEntity(job domain.Job) dao.Job {
	return dao.Job{
		Id:       job.Id,
		Cfg:      job.Cfg,
		NextTime: job.NextTime().UnixMilli(),
		Name:     job.Name,
		Executor: job.Executor,
	}
}

func (repo *PreemptCronJobRepository) entityToDomain(job dao.Job) domain.Job {
	return domain.Job{
		Id:       job.Id,
		Cfg:      job.Cfg,
		Name:     job.Name,
		Executor: job.Executor,
	}
}
