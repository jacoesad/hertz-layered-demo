package service

import (
	"context"
	"errors"

	"hz-server/internal/apperror"
	"hz-server/internal/task/domain"
)

const (
	CodeTaskNotFound     int32 = 10001
	CodeTaskNotStartable int32 = 10002
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
	task, err := s.repo.FindByTenantAndID(ctx, tenantID, taskID)
	if err != nil {
		return nil, toAppError(err)
	}
	return task, nil
}

func (s *service) StartTask(ctx context.Context, input StartTaskInput) (*domain.StartTaskResult, error) {
	task, err := s.repo.FindByTenantAndID(ctx, input.TenantID, input.TaskID)
	if err != nil {
		return nil, toAppError(err)
	}
	if err := task.EnsureStartable(); err != nil {
		return nil, toAppError(err)
	}
	return s.taskRunner.StartTask(ctx, input)
}

func toAppError(err error) error {
	if errors.Is(err, domain.ErrTaskNotFound) {
		return apperror.New(CodeTaskNotFound, "task not found", err)
	}
	if errors.Is(err, domain.ErrTaskNotStartable) {
		return apperror.New(CodeTaskNotStartable, "task is not startable", err)
	}
	return err
}
