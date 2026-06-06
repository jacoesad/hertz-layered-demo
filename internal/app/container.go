package app

import (
	"fmt"

	"hz-server/internal/database"
	"hz-server/internal/downstream"
	subtaskrepo "hz-server/internal/subtask/repo"
	subtaskservice "hz-server/internal/subtask/service"
	subtasksqlstore "hz-server/internal/subtask/sqlstore"
	taskrepo "hz-server/internal/task/repo"
	taskservice "hz-server/internal/task/service"
	tasksqlstore "hz-server/internal/task/sqlstore"
)

type Container struct {
	TaskService    taskservice.Service
	SubtaskService subtaskservice.Service
}

var Default *Container

func Init(container *Container) {
	Default = container
}

func MustDefault() *Container {
	if Default == nil {
		panic("app container is not initialized")
	}
	return Default
}

func NewContainer(ds *database.DataSources, clients *downstream.Clients) (*Container, error) {
	if ds == nil {
		return nil, fmt.Errorf("data sources are required")
	}
	if ds.DB == nil {
		return nil, fmt.Errorf("main database is required")
	}
	if ds.StarRocksDB == nil {
		return nil, fmt.Errorf("starrocks database is required")
	}
	if clients == nil {
		return nil, fmt.Errorf("downstream clients are required")
	}
	if clients.TaskRunner == nil {
		return nil, fmt.Errorf("task runner client is required")
	}

	taskSQL := tasksqlstore.New(ds.DB)
	taskRepo := taskrepo.New(taskSQL)

	subtaskSQL := subtasksqlstore.New(ds.DB)
	subtaskRepo := subtaskrepo.New(subtaskSQL)

	starRocksSubtaskSQL := subtasksqlstore.New(ds.StarRocksDB)
	starRocksSubtaskRepo := subtaskrepo.New(starRocksSubtaskSQL)

	return &Container{
		TaskService:    taskservice.New(taskRepo, clients.TaskRunner),
		SubtaskService: subtaskservice.NewWithStarRocks(subtaskRepo, starRocksSubtaskRepo),
	}, nil
}
