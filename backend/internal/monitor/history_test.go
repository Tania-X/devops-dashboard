package monitor

import (
	"testing"
	"time"
)

// TestHistory_RecordAndQuery 验证基本的记录和查询逻辑
func TestHistory_RecordAndQuery(t *testing.T) {
	// 保留1小时，每1分钟采一个点，最多60个点
	h := NewHistory(time.Hour, time.Minute)

	now := time.Now()
	h.Record(HistoryPoint{Timestamp: now.Add(-30 * time.Minute), CPUPercent: 10.0, MemoryPercent: 20.0})
	h.Record(HistoryPoint{Timestamp: now.Add(-20 * time.Minute), CPUPercent: 20.0, MemoryPercent: 30.0})
	h.Record(HistoryPoint{Timestamp: now.Add(-10 * time.Minute), CPUPercent: 30.0, MemoryPercent: 40.0})

	labels, cpu, memory := h.Query(1) // 查最近1小时

	if len(labels) != 3 {
		t.Fatalf("期望 3 个点，got %d", len(labels))
	}

	cases := []struct {
		idx        int
		wantCPU    float64
		wantMemory float64
		wantLabel  string
	}{
		{0, 10.0, 20.0, now.Add(-30 * time.Minute).Format("15:04")},
		{1, 20.0, 30.0, now.Add(-20 * time.Minute).Format("15:04")},
		{2, 30.0, 40.0, now.Add(-10 * time.Minute).Format("15:04")},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			if cpu[c.idx] != c.wantCPU {
				t.Errorf("cpu[%d] = %.1f, want %.1f", c.idx, cpu[c.idx], c.wantCPU)
			}
			if memory[c.idx] != c.wantMemory {
				t.Errorf("memory[%d] = %.1f, want %.1f", c.idx, memory[c.idx], c.wantMemory)
			}
			if labels[c.idx] != c.wantLabel {
				t.Errorf("labels[%d] = %s, want %s", c.idx, labels[c.idx], c.wantLabel)
			}
		})
	}
}

// TestHistory_QueryCutoff 验证 Query 会正确过滤超时的数据
func TestHistory_QueryCutoff(t *testing.T) {
	// 保留1小时，每1分钟一个点
	h := NewHistory(time.Hour, time.Minute)

	now := time.Now()
	// 插入一个 3 小时前的点（超出查询范围）
	h.Record(HistoryPoint{Timestamp: now.Add(-3 * time.Hour), CPUPercent: 99.0, MemoryPercent: 99.0})
	// 插入一个 30 分钟前的点（在范围内）
	h.Record(HistoryPoint{Timestamp: now.Add(-30 * time.Minute), CPUPercent: 50.0, MemoryPercent: 60.0})

	labels, cpu, _ := h.Query(1) // 只查最近1小时

	if len(labels) != 1 {
		t.Fatalf("期望过滤后剩 1 个点，got %d", len(labels))
	}
	if cpu[0] != 50.0 {
		t.Errorf("过滤逻辑错误，cpu[0] = %.1f, want 50.0", cpu[0])
	}
}

// TestHistory_Eviction 验证容量满了之后淘汰最旧的数据
func TestHistory_Eviction(t *testing.T) {
	// 保留5秒，每1秒一个点，最多5个点
	h := NewHistory(5*time.Second, time.Second)

	base := time.Now()
	for i := 0; i < 10; i++ {
		h.Record(HistoryPoint{
			Timestamp:  base.Add(time.Duration(i) * time.Second),
			CPUPercent: float64(i),
		})
	}

	_, cpu, _ := h.Query(1)

	if len(cpu) != 5 {
		t.Fatalf("期望淘汰后剩 5 个点，got %d", len(cpu))
	}
	// 0,1,2,3,4 被淘汰，第一个应该是 5
	if cpu[0] != 5.0 {
		t.Errorf("淘汰逻辑错误，首个值: %.1f, want 5.0", cpu[0])
	}
	// 最后一个是 9
	if cpu[4] != 9.0 {
		t.Errorf("淘汰逻辑错误，末个值: %.1f, want 9.0", cpu[4])
	}
}

// TestHistory_Concurrent 验证并发读写安全
func TestHistory_Concurrent(t *testing.T) {
	h := NewHistory(time.Hour, time.Second)

	// 启动 10 个 goroutine 同时写入
	for i := 0; i < 10; i++ {
		go func(val float64) {
			for j := 0; j < 100; j++ {
				h.Record(HistoryPoint{
					Timestamp:  time.Now(),
					CPUPercent: val,
				})
			}
		}(float64(i))
	}

	// 同时启动查询
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				h.Query(1)
			}
		}()
	}

	time.Sleep(200 * time.Millisecond)
	// 如果能跑完不 panic，说明锁保护有效
	labels, _, _ := h.Query(1)
	t.Logf("并发测试完成，当前点数: %d", len(labels))
}

// TestHistory_StartCollector 验证后台采集 goroutine 能正常工作
func TestHistory_StartCollector(t *testing.T) {
	// 用 100ms 间隔快速验证，保留 1 秒，最多 10 个点
	h := NewHistory(time.Second, 100*time.Millisecond)

	stopCh := h.StartCollector(100 * time.Millisecond)
	defer close(stopCh)

	// 等 350ms，应该至少采到 2~3 个点
	time.Sleep(350 * time.Millisecond)

	labels, cpu, memory := h.Query(1)
	if len(labels) == 0 {
		t.Fatal("后台采集没有写入任何数据")
	}

	t.Logf("✓ 后台采集正常，采到 %d 个点: CPU=%.1f%% Memory=%.1f%%",
		len(labels), cpu[len(cpu)-1], memory[len(memory)-1])
}
