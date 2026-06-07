package task

import (
	internalapp "hz-server/internal/app"
	taskservice "hz-server/internal/task/service"
)

func service() taskservice.Service {
	return internalapp.MustDefault().TaskService
}
