package api

import (
	"net/http"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
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
	c.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.services.UserService.Delete(id); err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}
