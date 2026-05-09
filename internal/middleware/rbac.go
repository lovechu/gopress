package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/gopress/internal/user"
	"github.com/yourorg/gopress/pkg/response"
)

// RequireRole 要求请求者角色 >= 任意一个指定角色（层级校验）
// 用法：admin.Use(RequireRole(user.RoleAdmin))
func RequireRole(required ...user.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("user_role")
		if !exists {
			response.Abort(c, http.StatusUnauthorized, "not authenticated")
			return
		}

		currentRole := user.Role(roleVal.(string))

		for _, r := range required {
			if currentRole.HasPermission(r) {
				c.Next()
				return
			}
		}

		response.Abort(c, http.StatusForbidden, "insufficient permissions")
	}
}

// RequireOwnerOrRole 允许资源所有者 OR 指定角色访问
func RequireOwnerOrRole(ownerIDFunc func(*gin.Context) uint, roles ...user.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUserID, _ := c.Get("user_id")
		currentRole := user.Role(c.GetString("user_role"))

		for _, r := range roles {
			if currentRole.HasPermission(r) {
				c.Next()
				return
			}
		}

		ownerID := ownerIDFunc(c)
		if currentUserID.(uint) == ownerID {
			c.Next()
			return
		}

		response.Abort(c, http.StatusForbidden, "access denied")
	}
}
