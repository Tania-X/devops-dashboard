package service

import (
	"sort"
	"strings"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/Tania-X/devops-dashboard/backend/internal/monitor"
	"gorm.io/gorm"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type MonitorService struct {
	db *gorm.DB
}

func NewMonitorService(db *gorm.DB) *MonitorService {
	return &MonitorService{db: db}
}

func (m *MonitorService) GetProcessList(sortBy string, order string, keyword string, limit int) ([]model.ProcessItem, error) {
	pids, err := process.Pids()
	if err != nil {
		return nil, err
	}

	var items []model.ProcessItem
	for _, pid := range pids {
		p, err := process.NewProcess(pid)
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

	return items, nil
}

func (m *MonitorService) GetProcessDetail(pid int) (*model.ProcessDetail, error) {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, nil // 进程不存在，不是系统错误
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

	return &model.ProcessDetail{
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
	}, nil
}

func (m *MonitorService) GetHostInfo() (*model.HostInfo, error) {
	info, err := host.Info()
	if err != nil {
		return nil, err
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
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

	return &model.HostInfo{
		Hostname:        info.Hostname,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		Arch:            info.KernelArch,
		BootTime:        bootTime,
		Uptime:          uptime.String(),
		CPUModel:        cpuModel,
		CPUCores:        cpuCores,
		CPULogicalCores: cpuLogical,
		TotalMemoryGb:   monitor.Round(float64(memInfo.Total) / 1024 / 1024 / 1024),
		VirtualMemoryGb: monitor.Round(float64(memInfo.SwapTotal) / 1024 / 1024 / 1024),
	}, nil
}
