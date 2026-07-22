package monitor

import (
	"sync"
	"time"
)

// HistoryPoint 单个时间点的指标快照
type HistoryPoint struct {
	Timestamp     time.Time
	CPUPercent    float64
	MemoryPercent float64
}

// History 线程安全的历史数据缓存（内存存储）
type History struct {
	mu      sync.RWMutex
	points  []HistoryPoint
	maxSize int
}

// NewHistory 创建历史记录器
//   retain: 保留多长时间的数据（如 24h）
//   interval: 采样间隔（如 10s）
// 根据两者计算出最多存储多少个点
func NewHistory(retain, interval time.Duration) *History {
	maxSize := int(retain / interval)
	if maxSize < 1 {
		maxSize = 1
	}
	return &History{
		points:  make([]HistoryPoint, 0, maxSize),
		maxSize: maxSize,
	}
}

// Record 记录一个数据点（线程安全）
// 供采集器调用，也供测试直接注入假数据
func (h *History) Record(p HistoryPoint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 超出容量时淘汰最旧的数据（队头）
	if len(h.points) >= h.maxSize {
		h.points = h.points[1:]
	}
	h.points = append(h.points, p)
}

// Query 查询最近 hours 小时的数据
// 返回三个切片：时间标签、CPU数组、内存数组，三者一一对应
func (h *History) Query(hours int) (labels []string, cpu []float64, memory []float64) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)

	for _, p := range h.points {
		if p.Timestamp.Before(cutoff) {
			continue
		}
		labels = append(labels, p.Timestamp.Format("15:04"))
		cpu = append(cpu, p.CPUPercent)
		memory = append(memory, p.MemoryPercent)
	}
	return
}

// StartCollector 启动后台 goroutine 定时采样
// 返回 stop channel，调用方通过 close(stopCh) 来停止采集
func (h *History) StartCollector(interval time.Duration) chan struct{} {
	stopCh := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				snapshot, err := Collect()
				if err != nil {
					continue
				}
				h.Record(HistoryPoint{
					Timestamp:     time.Now(),
					CPUPercent:    snapshot.CPUPercent,
					MemoryPercent: snapshot.MemoryPercent,
				})
			case <-stopCh:
				return
			}
		}
	}()
	return stopCh
}
