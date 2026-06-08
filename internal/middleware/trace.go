package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"hz-server/internal/trace"
)

func Trace() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		traceID := string(c.GetHeader(trace.HeaderTraceID))
		if traceID == "" {
			traceID = trace.NewTraceID()
		}
		c.Response.Header.Set(trace.HeaderTraceID, traceID)
		c.Next(trace.WithTraceID(ctx, traceID))
	}
}
