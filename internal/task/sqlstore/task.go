package sqlstore

import (
	"context"
	"database/sql"
	"errors"

	"hz-server/internal/task/domain"
	"hz-server/internal/task/repo"
)

type Store struct {
	db *sql.DB
}

var _ repo.SQL = (*Store)(nil)

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) SelectByTenantAndID(ctx context.Context, tenantID string, taskID int64) (*repo.Row, error) {
	const query = `
		SELECT id, tenant_id, title, status, owner
		FROM tasks
		WHERE tenant_id = ? AND id = ?
	`

	var row repo.Row
	err := s.db.QueryRowContext(ctx, query, tenantID, taskID).Scan(
		&row.ID,
		&row.TenantID,
		&row.Title,
		&row.Status,
		&row.Owner,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, err
	}
	return &row, nil
}
