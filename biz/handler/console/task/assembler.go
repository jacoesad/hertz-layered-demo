package task

import (
	task "hz-server/biz/model/console/task"
	"hz-server/internal/task/domain"
)

func toConsoleTaskInfo(item *domain.Task) *task.ConsoleTaskInfo {
	return &task.ConsoleTaskInfo{
		ID:       item.ID,
		TenantID: item.TenantID,
		Title:    item.Title,
		Status:   item.Status,
		Owner:    item.Owner,
	}
}

func toStartTaskResult(result *domain.StartTaskResult) *task.StartTaskResult {
	return &task.StartTaskResult{
		Accepted: result.Accepted,
		JobID:    result.JobID,
		Message:  result.Message,
	}
}
