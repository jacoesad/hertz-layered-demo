package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"hz-server/internal/logger"
	"hz-server/internal/signature"
)

func Signature(verifier signature.Verifier, log logger.Logger) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if verifier == nil || !verifier.Enabled() {
			c.Next(ctx)
			return
		}

		method := string(c.Method())
		path := string(c.Path())
		header := verifier.Header()
		value := string(c.GetHeader(header))

		if err := verifier.Verify(method, path, value); err != nil {
			if log != nil {
				log.Errorf(ctx, "signature rejected method=%s path=%s err=%v", method, path, err)
			}
			c.AbortWithStatusJSON(consts.StatusUnauthorized, map[string]any{
				"code":    consts.StatusUnauthorized,
				"message": err.Error(),
			})
			return
		}

		if log != nil {
			log.Infof(ctx, "signature accepted method=%s path=%s", method, path)
		}
		c.Next(ctx)
	}
}
