package repo

import (
	"context"

	"hz-server/internal/subtask/domain"
)

type Row struct {
	ID          int64
	TenantID    string
	TaskID      int64
	SubtaskType string
	Title       string
	Status      string
	Assignee    string
}

type SQL interface {
	SelectByTenantAndID(ctx context.Context, tenantID string, subtaskID int64) (*Row, error)
	SelectByCriteria(ctx context.Context, criteria domain.ListCriteria) ([]Row, error)
}

type Repository struct {
	sql SQL
}

func New(sql SQL) *Repository {
	return &Repository{sql: sql}
}

func (r *Repository) FindByTenantAndID(ctx context.Context, tenantID string, subtaskID int64) (*domain.Subtask, error) {
	row, err := r.sql.SelectByTenantAndID(ctx, tenantID, subtaskID)
	if err != nil {
		return nil, err
	}

	return &domain.Subtask{
		ID:       row.ID,
		TenantID: row.TenantID,
		TaskID:   row.TaskID,
		Type:     row.SubtaskType,
		Title:    row.Title,
		Status:   row.Status,
		Assignee: row.Assignee,
	}, nil
}

func (r *Repository) ListByCriteria(ctx context.Context, criteria domain.ListCriteria) ([]*domain.Subtask, error) {
	rows, err := r.sql.SelectByCriteria(ctx, criteria)
	if err != nil {
		return nil, err
	}

	items := make([]*domain.Subtask, 0, len(rows))
	for _, row := range rows {
		items = append(items, &domain.Subtask{
			ID:       row.ID,
			TenantID: row.TenantID,
			TaskID:   row.TaskID,
			Type:     row.SubtaskType,
			Title:    row.Title,
			Status:   row.Status,
			Assignee: row.Assignee,
		})
	}
	return items, nil
}
