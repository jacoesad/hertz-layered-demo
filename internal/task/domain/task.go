package domain

import "errors"

var ErrTaskNotFound = errors.New("task not found")
var ErrTaskNotStartable = errors.New("task is not startable")

const TaskStatusDone = "done"

type Task struct {
	ID       int64
	TenantID string
	Title    string
	Status   string
	Owner    string
}

func (t *Task) EnsureStartable() error {
	if t.Status == TaskStatusDone {
		return ErrTaskNotStartable
	}
	return nil
}

type StartTaskResult struct {
	Accepted bool
	JobID    string
	Message  string
}
