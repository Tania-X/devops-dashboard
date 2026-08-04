package api

import (
	"net/http"
	"strings"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.services.AuthService == nil {
			c.Next()
			return
		}
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			ErrorJSON(c, http.StatusUnauthorized, "未提供认证令牌")
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			ErrorJSON(c, http.StatusUnauthorized, "认证格式错误")
			c.Abort()
			return
		}
		user, err := h.services.AuthService.ValidateToken(tokenString)
		if err != nil {
			ErrorJSON(c, http.StatusUnauthorized, "认证令牌无效或已过期")
			c.Abort()
			return
		}
		c.Set("currentUser", user)
		c.Next()
	}
}

func (h *Handler) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.services.AuthService == nil {
			c.Next()
			return
		}
		user, exists := c.Get("currentUser")
		if !exists {
			ErrorJSON(c, http.StatusForbidden, "需要管理员权限")
			c.Abort()
			return
		}
		u, ok := user.(*model.User)
		if !ok || u.Role != "admin" {
			ErrorJSON(c, http.StatusForbidden, "需要管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
