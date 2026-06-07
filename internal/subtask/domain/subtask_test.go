package domain

import (
	"errors"
	"testing"
)

func TestListCriteriaValidate(t *testing.T) {
	taskID := int64(1001)

	tests := []struct {
		name    string
		input   ListCriteria
		wantErr error
	}{
		{
			name:    "requires task id or subtask type",
			input:   NewListCriteria("tenant-a", nil, ""),
			wantErr: ErrSubtaskListFilterRequired,
		},
		{
			name:  "accepts task id",
			input: NewListCriteria("tenant-a", &taskID, ""),
		},
		{
			name:  "accepts subtask type",
			input: NewListCriteria("tenant-a", nil, "manual"),
		},
		{
			name:  "trims subtask type before validation",
			input: NewListCriteria("tenant-a", nil, "  manual  "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
