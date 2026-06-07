package subtask

import (
	internalapp "hz-server/internal/app"
	subtaskservice "hz-server/internal/subtask/service"
)

func service() subtaskservice.Service {
	return internalapp.MustDefault().SubtaskService
}
