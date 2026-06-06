package task

import (
	task "hz-server/biz/model/openapi/task"
	"hz-server/internal/task/domain"
)

func toOpenAPITaskInfo(item *domain.Task) *task.OpenAPITaskInfo {
	return &task.OpenAPITaskInfo{
		ID:     item.ID,
		Title:  item.Title,
		Status: item.Status,
	}
}
