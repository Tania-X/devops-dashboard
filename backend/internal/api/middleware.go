package api

import (
	"net/http"
	"strings"

	"github.com/Tania-X/devops-dashboard/backend/internal/authz"
	userdomain "github.com/Tania-X/devops-dashboard/backend/internal/dashboard/user/domain"
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

// RequirePermission 权限校验中间件：校验当前登录用户角色是否拥有指定权限点（如 "user:read"）。
// 取代原来的 AdminMiddleware——权限点驱动，支持任意角色（admin/viewer/operator...）。
func RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("currentUser")
		if !exists {
			ErrorJSON(c, http.StatusUnauthorized, "未认证")
			c.Abort()
			return
		}
		u, ok := user.(*userdomain.User)
		if !ok {
			ErrorJSON(c, http.StatusForbidden, "权限不足")
			c.Abort()
			return
		}
		obj, act := splitPermissionPoint(perm)
		allowed, err := authz.HasPermission(u.Role, obj, act)
		if err != nil || !allowed {
			ErrorJSON(c, http.StatusForbidden, "权限不足")
			c.Abort()
			return
		}
		c.Next()
	}
}

// splitPermissionPoint 拆分 "obj:act" 为 (obj, act)。
func splitPermissionPoint(perm string) (string, string) {
	for i := 0; i < len(perm); i++ {
		if perm[i] == ':' {
			return perm[:i], perm[i+1:]
		}
	}
	return perm, ""
}
