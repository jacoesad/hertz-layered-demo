package domain

import "errors"

var ErrTaskNotFound = errors.New("task not found")

type Task struct {
	ID       int64
	TenantID string
	Title    string
	Status   string
	Owner    string
}
