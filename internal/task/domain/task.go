package domain

import "errors"

var ErrTaskNotFound = errors.New("task not found")
var ErrTaskNotStartable = errors.New("task is not startable")

type Task struct {
	ID       int64
	TenantID string
	Title    string
	Status   string
	Owner    string
}

type StartTaskResult struct {
	Accepted bool
	JobID    string
	Message  string
}
