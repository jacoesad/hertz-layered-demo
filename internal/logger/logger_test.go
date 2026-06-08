package logger

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hz-server/internal/config"
	"hz-server/internal/trace"
)

func TestFormatLineIncludesTraceLevelCallerAndMessage(t *testing.T) {
	ctx := trace.WithTraceID(context.Background(), "trace-demo")

	line := formatLine(ctx, "INFO", "hello %s", "world")

	for _, want := range []string{
		"trace_id=trace-demo",
		"level=INFO",
		"caller=",
		"msg=hello world",
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("formatLine() = %q, want contains %q", line, want)
		}
	}
}

func TestNewWithFileOutput(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "app.log")

	log, err := New(config.LoggerConfig{
		Level:     "info",
		FilePath:  filePath,
		MaxSizeMB: 1,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	log.Infof(context.Background(), "hello file")

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "msg=hello file") {
		t.Fatalf("log file = %q, want message", string(data))
	}
}
