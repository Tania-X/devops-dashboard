package api

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/Tania-X/devops-dashboard/backend/internal/monitor"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// trendHistory 是全局的历史数据缓存，由 main.go 初始化并启动采集
var trendHistory *monitor.History

// InitMonitorHistory 初始化历史缓存并启动后台采集
func InitMonitorHistory(h *monitor.History) {
	trendHistory = h
}

// GetProcessList 获取本机进程列表
func GetProcessList(c *gin.Context) {
	sortBy := c.DefaultQuery("sortBy", "cpu")
	order := c.DefaultQuery("order", "desc")
	keyword := strings.ToLower(c.DefaultQuery("keyword", ""))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 200 {
		limit = 50
	}

	pids, err := process.Pids()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取进程列表失败"})
		return
	}

	var items []model.ProcessItem
	for _, pid := range pids {
		p, err := process.NewProcess(int32(pid))
		if err != nil {
			continue
		}

		name, _ := p.Name()
		if keyword != "" && !strings.Contains(strings.ToLower(name), keyword) {
			continue
		}

		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()
		memInfo, _ := p.MemoryInfo()
		status, _ := p.Status()
		cmdline, _ := p.Cmdline()

		memoryMb := 0.0
		if memInfo != nil {
			memoryMb = float64(memInfo.RSS) / 1024 / 1024
		}

		items = append(items, model.ProcessItem{
			PID:           int(pid),
			Name:          name,
			Cmdline:       cmdline,
			CPUPercent:    monitor.Round(cpuPercent),
			MemoryPercent: monitor.Round(float64(memPercent)),
			MemoryMb:      monitor.Round(memoryMb),
			Status:        strings.Join(status, ","),
		})
	}

	// 排序
	sort.Slice(items, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "pid":
			less = items[i].PID < items[j].PID
		case "name":
			less = items[i].Name < items[j].Name
		case "memory":
			less = items[i].MemoryPercent < items[j].MemoryPercent
		default: // cpu
			less = items[i].CPUPercent < items[j].CPUPercent
		}
		if order == "asc" {
			return less
		}
		return !less
	})

	if len(items) > limit {
		items = items[:limit]
	}

	c.JSON(http.StatusOK, items)
}

// GetProcessDetail 获取单个进程详情
func GetProcessDetail(c *gin.Context) {
	pid, err := strconv.Atoi(c.Param("pid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 PID"})
		return
	}

	p, err := process.NewProcess(int32(pid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "进程不存在或已退出"})
		return
	}

	name, _ := p.Name()
	cpuPercent, _ := p.CPUPercent()
	memPercent, _ := p.MemoryPercent()
	memInfo, _ := p.MemoryInfo()
	status, _ := p.Status()
	cmdline, _ := p.Cmdline()
	ppid, _ := p.Ppid()
	numThreads, _ := p.NumThreads()
	numOpenFiles, _ := p.OpenFiles()
	numConnections, _ := p.Connections()
	createTime, _ := p.CreateTime()
	username, _ := p.Username()
	cwd, _ := p.Cwd()
	env, _ := p.Environ()

	memoryMb := 0.0
	if memInfo != nil {
		memoryMb = float64(memInfo.RSS) / 1024 / 1024
	}

	createTimeStr := ""
	if createTime > 0 {
		createTimeStr = time.Unix(0, createTime*int64(time.Millisecond)).Format("2006-01-02 15:04:05")
	}

	c.JSON(http.StatusOK, model.ProcessDetail{
		ProcessItem: model.ProcessItem{
			PID:           pid,
			Name:          name,
			Cmdline:       cmdline,
			CPUPercent:    monitor.Round(cpuPercent),
			MemoryPercent: monitor.Round(float64(memPercent)),
			MemoryMb:      monitor.Round(memoryMb),
			Status:        strings.Join(status, ","),
		},
		PPid:           int(ppid),
		NumThreads:     int(numThreads),
		NumOpenFiles:   len(numOpenFiles),
		NumConnections: len(numConnections),
		CreateTime:     createTimeStr,
		Username:       username,
		WorkingDir:     cwd,
		Env:            env,
	})
}

// GetHostInfo 获取主机信息
func GetHostInfo(c *gin.Context) {
	info, err := host.Info()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取主机信息失败"})
		return
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取 CPU 信息失败"})
		return
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取内存信息失败"})
		return
	}

	cpuCountsPhysical, _ := cpu.Counts(false)
	cpuCountsLogical, _ := cpu.Counts(true)

	bootTime := time.Unix(int64(info.BootTime), 0).Format("2006-01-02 15:04:05")
	uptime := time.Duration(info.Uptime) * time.Second

	cpuModel := "Unknown"
	if len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	cpuCores := cpuCountsPhysical
	cpuLogical := cpuCountsLogical
	if cpuCores == 0 && len(cpuInfo) > 0 {
		cpuCores = int(cpuInfo[0].Cores)
	}
	if cpuLogical == 0 && len(cpuInfo) > 0 {
		cpuLogical = int(cpuInfo[0].Cores)
	}

	c.JSON(http.StatusOK, model.HostInfo{
		Hostname:         info.Hostname,
		OS:               info.OS,
		Platform:         info.Platform,
		PlatformVersion:  info.PlatformVersion,
		KernelVersion:    info.KernelVersion,
		Arch:             info.KernelArch,
		BootTime:         bootTime,
		Uptime:           uptime.String(),
		CPUModel:         cpuModel,
		CPUCores:         cpuCores,
		CPULogicalCores:  cpuLogical,
		TotalMemoryGb:    monitor.Round(float64(memInfo.Total) / 1024 / 1024 / 1024),
		VirtualMemoryGb:  monitor.Round(float64(memInfo.SwapTotal) / 1024 / 1024 / 1024),
	})
}
