package monitor

import (
	"fmt"
	"math"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type MetricSnapshot struct {
	CPUPercent    float64
	MemoryPercent float64
	DiskPercent   float64
}

func Collect() (*MetricSnapshot, error) {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, fmt.Errorf("采集 CPU 失败: %w", err)
	}
	if len(cpuPercent) == 0 {
		return nil, fmt.Errorf("CPU 数据为空")
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("采集内存失败: %w", err)
	}

	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("采集磁盘失败: %w", err)
	}

	return &MetricSnapshot{
		CPUPercent:    round(cpuPercent[0]),
		MemoryPercent: round(memInfo.UsedPercent),
		DiskPercent:   round(diskInfo.UsedPercent),
	}, nil
}

func Status(percent float64) string {
	if percent < 60 {
		return "normal"
	}
	if percent < 80 {
		return "warning"
	}
	return "critical"
}

func round(v float64) float64 {
	return math.Round(v*10) / 10
}
