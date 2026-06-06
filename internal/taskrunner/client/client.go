package client

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol"
	taskrunnerclient "hz-server/biz/client/task_runner_service"
	taskrunner "hz-server/biz/model/downstream/taskrunner"
	"hz-server/internal/task/domain"
	taskservice "hz-server/internal/task/service"
)

type generatedClient interface {
	StartTask(context.Context, *taskrunner.StartTaskCommand, ...config.RequestOption) (*taskrunner.StartTaskResult, *protocol.Response, error)
}

type Options struct {
	Endpoint string
	Timeout  time.Duration
	Client   generatedClient
}

type Client struct {
	timeout time.Duration
	client  generatedClient
}

func New(opts Options) (*Client, error) {
	generated := opts.Client
	if generated == nil {
		var err error
		generated, err = taskrunnerclient.NewTaskRunnerServiceClient(opts.Endpoint)
		if err != nil {
			return nil, fmt.Errorf("create task runner client: %w", err)
		}
	}

	return &Client{
		timeout: opts.Timeout,
		client:  generated,
	}, nil
}

func (c *Client) StartTask(ctx context.Context, input taskservice.StartTaskInput) (*domain.StartTaskResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, _, err := c.client.StartTask(ctx, &taskrunner.StartTaskCommand{
		TenantID: input.TenantID,
		TaskID:   input.TaskID,
	})
	if err != nil {
		return nil, fmt.Errorf("call task runner: %w", err)
	}

	return &domain.StartTaskResult{
		Accepted: resp.GetAccepted(),
		JobID:    resp.GetJobID(),
		Message:  resp.GetMessage(),
	}, nil
}
