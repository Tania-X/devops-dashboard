package api

import (
	"net/http"

	"github.com/Tania-X/devops-dashboard/backend/internal/authz"
	"github.com/gin-gonic/gin"
)

// ListRoles 返回全部角色及其权限点（供权限配置页矩阵渲染）。
// 仅管理员（路由标注 settings:manage）。
func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := authz.ListRoles()
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// ListPermissionGroups 返回权限点分组清单（obj 分组 + 中文标签），供权限矩阵按组渲染。
func (h *Handler) ListPermissionGroups(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"groups": authz.PermissionGroups()})
}

// UpdateRolePermissions 更新指定角色的权限点集合并热生效。
// admin 为通配策略锁定，返回 400。
type updateRolePermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

func (h *Handler) UpdateRolePermissions(c *gin.Context) {
	role := c.Param("role")
	var req updateRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	if err := authz.UpdateRolePermissions(role, req.Permissions); err != nil {
		ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}
	// 返回更新后的角色权限，供前端刷新
	perms, err := authz.PermissionsOf(role)
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": role, "permissions": perms})
}
