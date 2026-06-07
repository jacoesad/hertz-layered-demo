package service

import (
	"context"
	"errors"

	"hz-server/internal/apperror"
	"hz-server/internal/subtask/domain"
)

const (
	CodeSubtaskNotFound           int32 = 20001
	CodeSubtaskListFilterRequired int32 = 20002
)

type Repository interface {
	FindByTenantAndID(ctx context.Context, tenantID string, subtaskID int64) (*domain.Subtask, error)
	ListByCriteria(ctx context.Context, criteria domain.ListCriteria) ([]*domain.Subtask, error)
}

type Service interface {
	GetSubtask(ctx context.Context, tenantID string, subtaskID int64) (*domain.Subtask, error)
	ListSubtasks(ctx context.Context, input ListSubtasksInput) ([]*domain.Subtask, error)
}

type ListSubtasksInput struct {
	TenantID    string
	TaskID      *int64
	SubtaskType string
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
	subtask, err := s.repo.FindByTenantAndID(ctx, tenantID, subtaskID)
	if err != nil {
		return nil, toAppError(err)
	}
	return subtask, nil
}

func (s *service) ListSubtasks(ctx context.Context, input ListSubtasksInput) ([]*domain.Subtask, error) {
	criteria := domain.NewListCriteria(input.TenantID, input.TaskID, input.SubtaskType)
	if err := criteria.Validate(); err != nil {
		return nil, toAppError(err)
	}
	items, err := s.starRocksRepo.ListByCriteria(ctx, criteria)
	if err != nil {
		return nil, toAppError(err)
	}
	return items, nil
}

func toAppError(err error) error {
	if errors.Is(err, domain.ErrSubtaskNotFound) {
		return apperror.New(CodeSubtaskNotFound, "subtask not found", err)
	}
	if errors.Is(err, domain.ErrSubtaskListFilterRequired) {
		return apperror.New(CodeSubtaskListFilterRequired, "task_id or subtask_type is required", err)
	}
	return err
}
