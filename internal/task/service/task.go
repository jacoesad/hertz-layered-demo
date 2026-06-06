package service

import (
	"context"

	"hz-server/internal/task/domain"
)

type Repository interface {
	FindByTenantAndID(ctx context.Context, tenantID string, taskID int64) (*domain.Task, error)
}

type Service interface {
	GetTask(ctx context.Context, tenantID string, taskID int64) (*domain.Task, error)
}

type service struct {
	repo Repository
}

func New(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetTask(ctx context.Context, tenantID string, taskID int64) (*domain.Task, error) {
	return s.repo.FindByTenantAndID(ctx, tenantID, taskID)
}
