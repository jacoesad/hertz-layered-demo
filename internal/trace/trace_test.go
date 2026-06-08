package trace

import "testing"

func TestNewTraceID(t *testing.T) {
	first := NewTraceID()
	second := NewTraceID()

	if first == "" {
		t.Fatalf("NewTraceID() is empty")
	}
	if first == second {
		t.Fatalf("NewTraceID() generated duplicate id %q", first)
	}
}
