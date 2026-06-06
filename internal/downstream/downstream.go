package downstream

import (
	"fmt"
	"time"

	"hz-server/internal/config"
	taskservice "hz-server/internal/task/service"
	taskrunnerclient "hz-server/internal/taskrunner/client"
)

type Clients struct {
	TaskRunner taskservice.TaskRunner
}

func Init(cfg config.DownstreamConfig) (*Clients, error) {
	if cfg.TaskRunner.Endpoint == "" {
		return nil, fmt.Errorf("task runner endpoint is required")
	}
	if cfg.TaskRunner.TimeoutMS <= 0 {
		return nil, fmt.Errorf("task runner timeout must be positive")
	}

	taskRunner, err := taskrunnerclient.New(taskrunnerclient.Options{
		Endpoint: cfg.TaskRunner.Endpoint,
		Timeout:  time.Duration(cfg.TaskRunner.TimeoutMS) * time.Millisecond,
	})
	if err != nil {
		return nil, err
	}

	return &Clients{
		TaskRunner: taskRunner,
	}, nil
}
