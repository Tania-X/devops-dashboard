package api

import (
	"net/http"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetAgentList(c *gin.Context) {
	agents, err := h.services.AgentService.List()
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, agents)
}

func (h *Handler) CreateAgent(c *gin.Context) {
	var req model.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	result, err := h.services.AgentService.Create(model.AgentTarget{
		Name:      req.Name,
		Host:      req.Host,
		Port:      req.Port,
		Username:  req.Username,
		AuthType:  req.AuthType,
		Password:  req.Password,
		DeployDir: req.DeployDir,
		AgentPort: req.AgentPort,
	})
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) UpdateAgent(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	result, err := h.services.AgentService.Update(id, model.AgentTarget{
		Name:      req.Name,
		Host:      req.Host,
		Port:      req.Port,
		Username:  req.Username,
		AuthType:  req.AuthType,
		Password:  req.Password,
		DeployDir: req.DeployDir,
		AgentPort: req.AgentPort,
	})
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteAgent(c *gin.Context) {
	id := c.Param("id")
	if err := h.services.AgentService.Delete(id); err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

func (h *Handler) DeployAgent(c *gin.Context) {
	id := c.Param("id")
	if err := h.services.AgentService.Deploy(id); err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "部署成功"})
}

func (h *Handler) StopAgent(c *gin.Context) {
	id := c.Param("id")
	if err := h.services.AgentService.Stop(id); err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已停止"})
}

func (h *Handler) CheckAgentStatus(c *gin.Context) {
	id := c.Param("id")
	status, err := h.services.AgentService.StatusCheck(id)
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": status})
}
