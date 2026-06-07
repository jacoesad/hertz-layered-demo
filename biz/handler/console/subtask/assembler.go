package subtask

import (
	subtask "hz-server/biz/model/console/subtask"
	"hz-server/internal/subtask/domain"
	subtaskservice "hz-server/internal/subtask/service"
)

func toConsoleSubtaskInfo(item *domain.Subtask) *subtask.ConsoleSubtaskInfo {
	return &subtask.ConsoleSubtaskInfo{
		ID:       item.ID,
		TenantID: item.TenantID,
		TaskID:   item.TaskID,
		Title:    item.Title,
		Status:   item.Status,
		Assignee: item.Assignee,
	}
}

func toConsoleSubtaskInfoList(items []*domain.Subtask) []*subtask.ConsoleSubtaskInfo {
	data := make([]*subtask.ConsoleSubtaskInfo, 0, len(items))
	for _, item := range items {
		data = append(data, toConsoleSubtaskInfo(item))
	}
	return data
}

func toListSubtasksInput(req subtask.ListSubtasksRequest) subtaskservice.ListSubtasksInput {
	return subtaskservice.ListSubtasksInput{
		TenantID:    req.TenantID,
		TaskID:      req.TaskID,
		SubtaskType: req.SubtaskType,
	}
}
