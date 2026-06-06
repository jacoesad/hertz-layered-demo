package service

import (
	"context"

	"hz-server/internal/task/domain"
)

type Repository interface {
	FindByTenantAndID(ctx context.Context, tenantID string, taskID int64) (*domain.Task, error)
}

type TaskRunner interface {
	StartTask(ctx context.Context, input StartTaskInput) (*domain.StartTaskResult, error)
}

type Service interface {
	GetTask(ctx context.Context, tenantID string, taskID int64) (*domain.Task, error)
	StartTask(ctx context.Context, input StartTaskInput) (*domain.StartTaskResult, error)
}

type StartTaskInput struct {
	TenantID string
	TaskID   int64
}

type service struct {
	repo       Repository
	taskRunner TaskRunner
}

func New(repo Repository, taskRunner TaskRunner) Service {
	return &service{
		repo:       repo,
		taskRunner: taskRunner,
	}
}

func (s *service) GetTask(ctx context.Context, tenantID string, taskID int64) (*domain.Task, error) {
	return s.repo.FindByTenantAndID(ctx, tenantID, taskID)
}

func (s *service) StartTask(ctx context.Context, input StartTaskInput) (*domain.StartTaskResult, error) {
	task, err := s.repo.FindByTenantAndID(ctx, input.TenantID, input.TaskID)
	if err != nil {
		return nil, err
	}
	if task.Status == "done" {
		return nil, domain.ErrTaskNotStartable
	}
	return s.taskRunner.StartTask(ctx, input)
}
