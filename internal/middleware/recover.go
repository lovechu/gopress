package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/gopress/pkg/response"
	"go.uber.org/zap"
)

func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				)
				response.Abort(c, http.StatusInternalServerError, "internal server error")
			}
		}()
		c.Next()
	}
}
