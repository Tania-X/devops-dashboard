package monitor

import (
	"log/slog"
	"sync"
	"time"
)

// HistoryPoint 单个时间点的指标快照
type HistoryPoint struct {
	Timestamp     time.Time
	CPUPercent    float64
	MemoryPercent float64
}

// History 线程安全的历史数据缓存（固定大小环形缓冲）
type History struct {
	mu       sync.RWMutex
	data     []HistoryPoint // 固定长度 maxSize，预分配
	writePos int            // 下一个写入位置
	count    int            // 当前有效元素数（0..maxSize）
	maxSize  int
	alerter  *Alerter
}

// NewHistory 创建历史记录器
//
//	retain: 保留多长时间的数据（如 24h）
//	interval: 采样间隔（如 10s）
//
// 根据两者计算出最多存储多少个点
func NewHistory(retain, interval time.Duration, alerter *Alerter) *History {
	maxSize := int(retain / interval)
	if maxSize < 1 {
		maxSize = 1
	}
	return &History{
		data:    make([]HistoryPoint, maxSize),
		maxSize: maxSize,
		alerter: alerter,
	}
}

// Record 记录一个数据点（线程安全）
// 供采集器调用，也供测试直接注入假数据
func (h *History) Record(p HistoryPoint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.data[h.writePos] = p
	h.writePos = (h.writePos + 1) % h.maxSize
	if h.count < h.maxSize {
		h.count++
	}
}

// Query 查询最近 hours 小时的数据
// 返回三个切片：时间标签、CPU数组、内存数组，三者一一对应
func (h *History) Query(hours int) (labels []string, cpu []float64, memory []float64) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.count == 0 {
		return
	}

	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	// 最旧有效数据的起始下标，+maxSize 是为了避免 writePos-count 为负数
	start := (h.writePos - h.count + h.maxSize) % h.maxSize

	for i := 0; i < h.count; i++ {
		idx := (start + i) % h.maxSize
		p := h.data[idx]
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
// 返回 (stopCh, done):调用方 close(stopCh) 停止采集,<-done 等待采集 goroutine 退出
// (确保退出前不再产生新告警,供关闭流程先停采集再关消费端,避免最后一批告警丢失)
func (h *History) StartCollector(interval time.Duration) (chan struct{}, chan struct{}) {
	stopCh := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				snapshot, err := Collect()
				if err != nil {
					slog.Warn("定时采集系统指标失败", "err", err)
					continue
				}
				h.Record(HistoryPoint{
					Timestamp:     time.Now(),
					CPUPercent:    snapshot.CPUPercent,
					MemoryPercent: snapshot.MemoryPercent,
				})
				if h.alerter != nil {
					h.alerter.Evaluate(snapshot)
				}
			case <-stopCh:
				return
			}
		}
	}()
	return stopCh, done
}
