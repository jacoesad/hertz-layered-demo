package subtask

import (
	subtask "hz-server/biz/model/openapi/subtask"
	"hz-server/internal/subtask/domain"
)

func toOpenAPISubtaskInfo(item *domain.Subtask) *subtask.OpenAPISubtaskInfo {
	return &subtask.OpenAPISubtaskInfo{
		ID:     item.ID,
		TaskID: item.TaskID,
		Title:  item.Title,
		Status: item.Status,
	}
}
