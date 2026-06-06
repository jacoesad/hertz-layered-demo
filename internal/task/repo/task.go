package repo

import (
	"context"

	"hz-server/internal/task/domain"
)

type Row struct {
	ID       int64
	TenantID string
	Title    string
	Status   string
	Owner    string
}

type SQL interface {
	SelectByTenantAndID(ctx context.Context, tenantID string, taskID int64) (*Row, error)
}

type Repository struct {
	sql SQL
}

func New(sql SQL) *Repository {
	return &Repository{sql: sql}
}

func (r *Repository) FindByTenantAndID(ctx context.Context, tenantID string, taskID int64) (*domain.Task, error) {
	row, err := r.sql.SelectByTenantAndID(ctx, tenantID, taskID)
	if err != nil {
		return nil, err
	}

	return &domain.Task{
		ID:       row.ID,
		TenantID: row.TenantID,
		Title:    row.Title,
		Status:   row.Status,
		Owner:    row.Owner,
	}, nil
}
