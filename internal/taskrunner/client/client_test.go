package client

import (
	"context"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol"
	taskrunner "hz-server/biz/model/downstream/taskrunner"
	taskservice "hz-server/internal/task/service"
)

func TestStartTaskUsesGeneratedClient(t *testing.T) {
	generated := &fakeGeneratedClient{
		response: &taskrunner.StartTaskResult{
			Accepted: true,
			JobID:    stringPtr("job-generated-1001"),
			Message:  stringPtr("accepted by generated client"),
		},
	}

	client, err := New(Options{
		Timeout: time.Second,
		Client:  generated,
	})
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	result, err := client.StartTask(t.Context(), taskservice.StartTaskInput{
		TenantID: "tenant-a",
		TaskID:   1001,
	})
	if err != nil {
		t.Fatalf("StartTask error: %v", err)
	}
	if generated.request == nil {
		t.Fatal("generated client was not called")
	}
	if generated.request.TenantID != "tenant-a" || generated.request.TaskID != 1001 {
		t.Fatalf("request = %+v, want tenant-a/1001", generated.request)
	}
	if !result.Accepted || result.JobID != "job-generated-1001" || result.Message != "accepted by generated client" {
		t.Fatalf("result = %+v", result)
	}
}

type fakeGeneratedClient struct {
	request  *taskrunner.StartTaskCommand
	response *taskrunner.StartTaskResult
}

func (f *fakeGeneratedClient) StartTask(ctx context.Context, req *taskrunner.StartTaskCommand, reqOpt ...config.RequestOption) (*taskrunner.StartTaskResult, *protocol.Response, error) {
	f.request = req
	return f.response, nil, nil
}

func stringPtr(value string) *string {
	return &value
}
