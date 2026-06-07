package domain

import (
	"errors"
	"testing"
)

func TestTaskEnsureStartable(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr error
	}{
		{
			name:   "startable when pending",
			status: "pending",
		},
		{
			name:    "not startable when done",
			status:  TaskStatusDone,
			wantErr: ErrTaskNotStartable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Status: tt.status}

			err := task.EnsureStartable()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("EnsureStartable() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
