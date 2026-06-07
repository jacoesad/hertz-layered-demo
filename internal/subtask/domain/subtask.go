package domain

import (
	"errors"
	"strings"
)

var ErrSubtaskNotFound = errors.New("subtask not found")
var ErrSubtaskListFilterRequired = errors.New("task_id or subtask_type is required")

type Subtask struct {
	ID       int64
	TenantID string
	TaskID   int64
	Type     string
	Title    string
	Status   string
	Assignee string
}

type ListCriteria struct {
	TenantID    string
	TaskID      *int64
	SubtaskType string
}

func NewListCriteria(tenantID string, taskID *int64, subtaskType string) ListCriteria {
	return ListCriteria{
		TenantID:    strings.TrimSpace(tenantID),
		TaskID:      taskID,
		SubtaskType: strings.TrimSpace(subtaskType),
	}
}

func (c ListCriteria) Validate() error {
	if c.TaskID == nil && c.SubtaskType == "" {
		return ErrSubtaskListFilterRequired
	}
	return nil
}
