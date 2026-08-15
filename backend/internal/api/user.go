package api

import (
	"net/http"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/Tania-X/devops-dashboard/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetUserList(c *gin.Context) {
	users, err := h.services.UserService.List()
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	result, err := h.services.UserService.Create(model.User{
		Username: req.Username,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.services.AuditService.Record(currentUsername(c), service.ActionUserCreate, "用户 "+req.Username, "role="+req.Role)
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	result, err := h.services.UserService.Update(id, model.User{Role: req.Role, Password: req.Password})
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.services.AuditService.Record(currentUsername(c), service.ActionUserUpdate, "用户 "+result.Username, "role="+req.Role)
	c.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	// 删除前取用户名（审计用）；删除失败不影响主流程
	var victim *model.User
	if users, err := h.services.UserService.List(); err == nil {
		for i := range users {
			if users[i].ID == id {
				victim = &users[i]
				break
			}
		}
	}
	if err := h.services.UserService.Delete(id); err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	name := "unknown"
	if victim != nil {
		name = victim.Username
	}
	h.services.AuditService.Record(currentUsername(c), service.ActionUserDelete, "用户 "+name, "")
	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}
