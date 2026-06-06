package service

import (
	"context"

	"hz-server/internal/subtask/domain"
)

type Repository interface {
	FindByTenantAndID(ctx context.Context, tenantID string, subtaskID int64) (*domain.Subtask, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.Subtask, error)
}

type Service interface {
	GetSubtask(ctx context.Context, tenantID string, subtaskID int64) (*domain.Subtask, error)
	ListSubtasks(ctx context.Context, tenantID string) ([]*domain.Subtask, error)
}

type service struct {
	repo          Repository
	starRocksRepo Repository
}

func New(repo Repository) Service {
	return NewWithStarRocks(repo, repo)
}

func NewWithStarRocks(repo Repository, starRocksRepo Repository) Service {
	return &service{
		repo:          repo,
		starRocksRepo: starRocksRepo,
	}
}

func (s *service) GetSubtask(ctx context.Context, tenantID string, subtaskID int64) (*domain.Subtask, error) {
	return s.repo.FindByTenantAndID(ctx, tenantID, subtaskID)
}

func (s *service) ListSubtasks(ctx context.Context, tenantID string) ([]*domain.Subtask, error) {
	return s.starRocksRepo.ListByTenant(ctx, tenantID)
}
