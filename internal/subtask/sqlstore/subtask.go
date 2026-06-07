package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"hz-server/internal/subtask/domain"
	"hz-server/internal/subtask/repo"
)

type Store struct {
	db *sql.DB
}

var _ repo.SQL = (*Store)(nil)

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) SelectByTenantAndID(ctx context.Context, tenantID string, subtaskID int64) (*repo.Row, error) {
	const query = `
		SELECT id, tenant_id, task_id, subtask_type, title, status, assignee
		FROM subtasks
		WHERE tenant_id = ? AND id = ?
	`

	var row repo.Row
	err := s.db.QueryRowContext(ctx, query, tenantID, subtaskID).Scan(
		&row.ID,
		&row.TenantID,
		&row.TaskID,
		&row.SubtaskType,
		&row.Title,
		&row.Status,
		&row.Assignee,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSubtaskNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (s *Store) SelectByCriteria(ctx context.Context, criteria domain.ListCriteria) ([]repo.Row, error) {
	var query strings.Builder
	query.WriteString(`
		SELECT id, tenant_id, task_id, subtask_type, title, status, assignee
		FROM subtasks
		WHERE tenant_id = ?
	`)

	args := []any{criteria.TenantID}
	if criteria.TaskID != nil {
		query.WriteString(" AND task_id = ?")
		args = append(args, *criteria.TaskID)
	}
	if criteria.SubtaskType != "" {
		query.WriteString(" AND subtask_type = ?")
		args = append(args, criteria.SubtaskType)
	}
	query.WriteString(" ORDER BY id ASC")

	rows, err := s.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]repo.Row, 0)
	for rows.Next() {
		var item repo.Row
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.TaskID,
			&item.SubtaskType,
			&item.Title,
			&item.Status,
			&item.Assignee,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
