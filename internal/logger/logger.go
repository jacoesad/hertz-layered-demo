package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gopkg.in/natefinch/lumberjack.v2"
	"hz-server/internal/config"
	"hz-server/internal/trace"
)

type Logger interface {
	Infof(ctx context.Context, format string, args ...any)
	Errorf(ctx context.Context, format string, args ...any)
}

type stdLogger struct {
	info  *log.Logger
	error *log.Logger
}

func New(cfg config.LoggerConfig) (Logger, error) {
	level := strings.ToLower(cfg.Level)
	if level == "" {
		level = "info"
	}

	writer, err := newWriter(cfg)
	if err != nil {
		return nil, err
	}

	infoWriter := writer
	if level == "error" {
		infoWriter = io.Discard
	}

	configureHLog(writer, level)

	return &stdLogger{
		info:  log.New(infoWriter, "", 0),
		error: log.New(writer, "", 0),
	}, nil
}

func (l *stdLogger) Infof(ctx context.Context, format string, args ...any) {
	if l == nil || l.info == nil {
		return
	}
	l.info.Print(formatLine(ctx, "INFO", format, args...))
}

func (l *stdLogger) Errorf(ctx context.Context, format string, args ...any) {
	if l == nil || l.error == nil {
		return
	}
	l.error.Print(formatLine(ctx, "ERROR", format, args...))
}

func formatLine(ctx context.Context, level string, format string, args ...any) string {
	traceID := trace.TraceID(ctx)
	if traceID == "" {
		traceID = "-"
	}
	return fmt.Sprintf(
		"%s trace_id=%s level=%s caller=%s msg=%s",
		time.Now().Format("2006-01-02 15:04:05.000"),
		traceID,
		level,
		caller(3),
		fmt.Sprintf(format, args...),
	)
}

func caller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "-"
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

func newWriter(cfg config.LoggerConfig) (io.Writer, error) {
	if cfg.FilePath == "" {
		return nil, fmt.Errorf("logger.file_path is required when file output is enabled")
	}
	if cfg.MaxSizeMB <= 0 {
		return nil, fmt.Errorf("logger.max_size_mb must be positive when file output is enabled")
	}

	if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	writer := io.MultiWriter(os.Stdout, &lumberjack.Logger{
		Filename: cfg.FilePath,
		MaxSize:  cfg.MaxSizeMB,
	})
	return writer, nil
}

func configureHLog(writer io.Writer, level string) {
	hlog.SetOutput(writer)
	switch level {
	case "debug":
		hlog.SetLevel(hlog.LevelDebug)
	case "error":
		hlog.SetLevel(hlog.LevelError)
	default:
		hlog.SetLevel(hlog.LevelInfo)
	}
}
