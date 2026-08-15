package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Tania-X/devops-dashboard/backend/internal/authz"
	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/Tania-X/devops-dashboard/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// currentUsername 从上下文中取当前操作人用户名（审计用）。
func currentUsername(c *gin.Context) string {
	if u, ok := c.Get("currentUser"); ok {
		if user, ok := u.(*model.User); ok {
			return user.Username
		}
	}
	return "unknown"
}

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

// CreateRole 创建自定义角色（默认空权限）。
func (h *Handler) CreateRole(c *gin.Context) {
	var req model.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	role := model.Role{
		Name:        req.Name,
		Label:       req.Label,
		Description: req.Description,
	}
	if err := authz.CreateRole(role); err != nil {
		ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}
	h.services.AuditService.Record(currentUsername(c), service.ActionRoleCreate, "角色 "+req.Name, req.Label)
	c.JSON(http.StatusCreated, gin.H{"role": req.Name, "label": req.Label})
}

// UpdateRole 更新角色显示名/描述（名称不可修改；内置角色显示名锁定）。
func (h *Handler) UpdateRole(c *gin.Context) {
	name := c.Param("role")
	var req model.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	before, err := authz.ListRoles()
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	if err := authz.UpdateRole(name, req); err != nil {
		ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}
	after, err := authz.ListRoles()
	if err != nil {
		// 更新已成功，仅审计的 after 部分获取失败：记录失败信息，不阻塞接口返回
		h.services.AuditService.Record(currentUsername(c), service.ActionRoleUpdate, "角色 "+name,
			"before="+rolesSummary(before, name)+" after=<获取失败: "+err.Error()+">")
		c.JSON(http.StatusOK, gin.H{"role": name})
		return
	}
	h.services.AuditService.Record(currentUsername(c), service.ActionRoleUpdate, "角色 "+name,
		"before="+rolesSummary(before, name)+" after="+rolesSummary(after, name))
	c.JSON(http.StatusOK, gin.H{"role": name})
}

// DeleteRole 删除自定义角色（内置角色/有用户绑定不可删）。
func (h *Handler) DeleteRole(c *gin.Context) {
	name := c.Param("role")
	if err := authz.DeleteRole(name); err != nil {
		ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}
	h.services.AuditService.Record(currentUsername(c), service.ActionRoleDelete, "角色 "+name, "")
	c.JSON(http.StatusOK, gin.H{"role": name, "deleted": true})
}

// ListAuditLogs 分页查询审计日志（仅管理员）。
func (h *Handler) ListAuditLogs(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 20)
	result, err := h.services.AuditService.List(page, size)
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": result.Items, "total": result.Total})
}

// rolesSummary 提取指定角色的 Label 用于审计详情。
func rolesSummary(roles []authz.RoleInfo, name string) string {
	for _, r := range roles {
		if r.Name == name {
			return r.Label
		}
	}
	return ""
}

// parseIntDefault 解析查询参数整数，非法/缺省时返回默认值。
func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return def
	}
	return n
}

// stringify 序列化任意值为 JSON 字符串（审计详情用；失败返回空串）。
func stringify(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
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
	// 审计：记录权限变更（PermissionsOf 失败时记录错误信息，不静默忽略）
	perms, err := authz.PermissionsOf(role)
	if err != nil {
		h.services.AuditService.Record(currentUsername(c), service.ActionPermissionUpdate, "角色 "+role,
			"<获取权限失败: "+err.Error()+">")
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.services.AuditService.Record(currentUsername(c), service.ActionPermissionUpdate, "角色 "+role, stringify(perms))
	// 返回更新后的角色权限，供前端刷新
	c.JSON(http.StatusOK, gin.H{"role": role, "permissions": perms})
}
