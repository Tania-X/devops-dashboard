package api

import (
	"net/http"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/gin-gonic/gin"
)

// GetWebhookConfig 获取告警 Webhook 配置（secret 不回传）
func (h *Handler) GetWebhookConfig(c *gin.Context) {
	cfg := h.services.WebhookManager.Get()
	c.JSON(http.StatusOK, cfg)
}

// UpdateWebhookConfig 更新告警 Webhook 配置（仅管理员），保存后热生效
func (h *Handler) UpdateWebhookConfig(c *gin.Context) {
	var input model.WebhookConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	if input.URL == "" {
		ErrorJSON(c, http.StatusBadRequest, "Webhook 地址不能为空")
		return
	}
	if input.Kind != "dingtalk" && input.Kind != "wecom" {
		ErrorJSON(c, http.StatusBadRequest, "渠道类型仅支持 dingtalk 或 wecom")
		return
	}

	cfg, err := h.services.WebhookManager.Update(input)
	if err != nil {
		ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// TestWebhookConfig 发送测试告警到当前配置的 Webhook（仅管理员）
func (h *Handler) TestWebhookConfig(c *gin.Context) {
	detail, err := h.services.WebhookManager.Test()
	if err != nil {
		ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "detail": detail})
}
