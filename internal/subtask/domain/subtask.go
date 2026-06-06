package domain

import "errors"

var ErrSubtaskNotFound = errors.New("subtask not found")

type Subtask struct {
	ID       int64
	TenantID string
	TaskID   int64
	Title    string
	Status   string
	Assignee string
}
