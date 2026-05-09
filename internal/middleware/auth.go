package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/gopress/internal/config"
	"github.com/yourorg/gopress/pkg/jwt"
	"github.com/yourorg/gopress/pkg/response"
)

// JWTAuth JWT 鉴权中间件
func JWTAuth(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Abort(c, http.StatusUnauthorized, "missing or invalid token")
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwt.ParseAccessToken(token, cfg.AccessSecret)
		if err != nil {
			response.Abort(c, http.StatusUnauthorized, "token expired or invalid")
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}
