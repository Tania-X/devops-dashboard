package api

import (
	"net/http"
	"strconv"

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
	var req model.UpdateWebhookConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	if req.URL == "" {
		ErrorJSON(c, http.StatusBadRequest, "Webhook 地址不能为空")
		return
	}
	if req.Kind != "dingtalk" && req.Kind != "wecom" {
		ErrorJSON(c, http.StatusBadRequest, "渠道类型仅支持 dingtalk 或 wecom")
		return
	}

	cfg, err := h.services.WebhookManager.Update(model.WebhookConfig{
		Enabled: req.Enabled,
		Kind:    req.Kind,
		URL:     req.URL,
		Secret:  req.Secret,
	})
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

// GetAlertThresholds 获取告警阈值配置(DB 有则返回,无则默认值)
func (h *Handler) GetAlertThresholds(c *gin.Context) {
	c.JSON(http.StatusOK, h.services.AlertThresholdManager.List())
}

// UpdateAlertThreshold 更新告警阈值(校验 + 热生效,下次采集即用新阈值)
func (h *Handler) UpdateAlertThreshold(c *gin.Context) {
	var req model.UpdateAlertThresholdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	if err := h.services.AlertThresholdManager.Update(req.Metric, req.WarnThreshold, req.CritThreshold); err != nil {
		ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, h.services.AlertThresholdManager.List())
}

// GetAlertHistory 告警历史分页查询(落库记录,支持级别筛选)
func (h *Handler) GetAlertHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	level := c.Query("level")

	// List 返回钳制后的 page/pageSize,响应字段与实际查询语义一致
	list, total, page, pageSize, err := h.services.AlertRecorder.List(page, pageSize, level)
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, "查询告警历史失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}
