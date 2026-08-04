package api

import (
	"net/http"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	resp, err := h.services.AuthService.Login(req.Username, req.Password)
	if err != nil {
		ErrorJSON(c, http.StatusUnauthorized, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetMe(c *gin.Context) {
	user, exists := c.Get("currentUser")
	if !exists {
		ErrorJSON(c, http.StatusUnauthorized, "未登录")
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "已退出登录"})
}
