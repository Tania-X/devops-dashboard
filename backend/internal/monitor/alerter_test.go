package monitor

import (
	"strings"
	"sync"
	"testing"
)

// ──────────────────────────────────────────────
// 本文件所有测试基于以下阈值：
//   CPU:    warning=60, critical=80
//   Memory: warning=70, critical=85
//   Disk:   warning=75, critical=90
//
// GetAlerts 返回的切片「最新的在前」，tests 依赖此排序约定。
// ──────────────────────────────────────────────

// ── 单指标场景（表驱动）───────────────────────

func TestAlerter_SingleMetric(t *testing.T) {
	cases := []struct {
		name      string
		snapshot  MetricSnapshot
		wantLevel string
		wantAlert bool
	}{
		// ── 正常值 ──
		{"cpu_normal", MetricSnapshot{CPUPercent: 30, MemoryPercent: 40, DiskPercent: 50}, "", false},
		// ── CPU ──
		{"cpu_warning", MetricSnapshot{CPUPercent: 65, MemoryPercent: 40, DiskPercent: 50}, "warning", true},
		{"cpu_critical", MetricSnapshot{CPUPercent: 85, MemoryPercent: 40, DiskPercent: 50}, "critical", true},
		{"cpu_boundary_60", MetricSnapshot{CPUPercent: 60, MemoryPercent: 40, DiskPercent: 50}, "warning", true}, // 恰好 60 → warning
		{"cpu_boundary_80", MetricSnapshot{CPUPercent: 80, MemoryPercent: 40, DiskPercent: 50}, "critical", true}, // 恰好 80
		{"cpu_boundary_81", MetricSnapshot{CPUPercent: 81, MemoryPercent: 40, DiskPercent: 50}, "critical", true}, // 超过 80
		// ── Memory ──
		{"mem_warning", MetricSnapshot{CPUPercent: 30, MemoryPercent: 75, DiskPercent: 50}, "warning", true},
		{"mem_critical", MetricSnapshot{CPUPercent: 30, MemoryPercent: 86, DiskPercent: 50}, "critical", true},
		{"mem_boundary_85", MetricSnapshot{CPUPercent: 30, MemoryPercent: 85, DiskPercent: 50}, "critical", true}, // 恰好 85
		// ── Disk ──
		{"disk_warning", MetricSnapshot{CPUPercent: 30, MemoryPercent: 40, DiskPercent: 76}, "warning", true},
		{"disk_critical", MetricSnapshot{CPUPercent: 30, MemoryPercent: 40, DiskPercent: 91}, "critical", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewAlerter()
			a.Evaluate(&tc.snapshot)
			alerts := a.GetAlerts(10)

			if !tc.wantAlert {
				if len(alerts) != 0 {
					t.Fatalf("期望无告警，得到 %d 条", len(alerts))
				}
				return
			}

			if len(alerts) == 0 {
				t.Fatal("期望有告警，但结果为空")
			}
			gotLevel := alerts[0].Level
			if gotLevel != tc.wantLevel {
				t.Errorf("期望 level=%s, 得到 level=%s", tc.wantLevel, gotLevel)
			}
			// 消息中应包含指标名和当前值
			metricLabels := map[string]string{"cpu": "CPU", "mem": "内存", "dis": "磁盘"}
			keyword := metricLabels[tc.name[0:3]]
			if !strings.Contains(alerts[0].Message, keyword) {
				t.Errorf("告警消息应包含指标名 %s，得到: %s", keyword, alerts[0].Message)
			}
		})
	}
}

// ── 有状态场景（独立测试）─────────────────────

func TestAlerter_Deduplication(t *testing.T) {
	a := NewAlerter()

	// 第一次：CPU 85 → 触发 critical
	a.Evaluate(&MetricSnapshot{CPUPercent: 85, MemoryPercent: 40, DiskPercent: 50})
	if n := len(a.GetAlerts(10)); n != 1 {
		t.Fatalf("第一次应生成 1 条告警，得到 %d 条", n)
	}

	// 第二次：CPU 还是 85 → 不重复
	a.Evaluate(&MetricSnapshot{CPUPercent: 85, MemoryPercent: 40, DiskPercent: 50})
	if n := len(a.GetAlerts(10)); n != 1 {
		t.Fatalf("状态未变不应重复告警，却有 %d 条", n)
	}
}

func TestAlerter_Escalate(t *testing.T) {
	a := NewAlerter()

	// CPU 70 → warning
	a.Evaluate(&MetricSnapshot{CPUPercent: 70, MemoryPercent: 40, DiskPercent: 50})
	// CPU 90 → 升级为 critical
	a.Evaluate(&MetricSnapshot{CPUPercent: 90, MemoryPercent: 40, DiskPercent: 50})

	alerts := a.GetAlerts(10)
	if len(alerts) != 2 {
		t.Fatalf("期望 2 条（warning + critical），得到 %d 条", len(alerts))
	}
	if alerts[0].Level != "critical" {
		t.Errorf("最新告警应是 critical，得到 %s", alerts[0].Level)
	}
}

func TestAlerter_Recovery(t *testing.T) {
	a := NewAlerter()

	// CPU 85 → critical
	a.Evaluate(&MetricSnapshot{CPUPercent: 85, MemoryPercent: 40, DiskPercent: 50})
	// CPU 降到 40 → 恢复
	a.Evaluate(&MetricSnapshot{CPUPercent: 40, MemoryPercent: 40, DiskPercent: 50})

	alerts := a.GetAlerts(10)
	if len(alerts) != 2 {
		t.Fatalf("期望 2 条（critical + recovery），得到 %d 条", len(alerts))
	}
	if alerts[0].Level != "info" {
		t.Errorf("恢复告警应是 info，得到 %s", alerts[0].Level)
	}
}

func TestAlerter_RecoveryThenRetrigger(t *testing.T) {
	a := NewAlerter()

	// CPU 85 → critical
	a.Evaluate(&MetricSnapshot{CPUPercent: 85, MemoryPercent: 40, DiskPercent: 50})
	// CPU 降到 40 → 恢复
	a.Evaluate(&MetricSnapshot{CPUPercent: 40, MemoryPercent: 40, DiskPercent: 50})
	// CPU 再次升到 90 → 应该再次告警（之前恢复过，状态已重置）
	a.Evaluate(&MetricSnapshot{CPUPercent: 90, MemoryPercent: 40, DiskPercent: 50})

	alerts := a.GetAlerts(10)
	if len(alerts) != 3 {
		t.Fatalf("期望 3 条（critical→recovery→critical），得到 %d 条", len(alerts))
	}
	if alerts[0].Level != "critical" {
		t.Errorf("最新告警应是 critical（第二次触发），得到 %s", alerts[0].Level)
	}
}

func TestAlerter_MultipleMetrics(t *testing.T) {
	a := NewAlerter()

	// CPU 85 + Disk 80 → 2 条告警
	a.Evaluate(&MetricSnapshot{CPUPercent: 85, MemoryPercent: 40, DiskPercent: 80})

	alerts := a.GetAlerts(10)
	if len(alerts) != 2 {
		t.Fatalf("期望 2 条（CPU + Disk），得到 %d 条", len(alerts))
	}
}

func TestAlerter_GetAlertsLimit(t *testing.T) {
	a := NewAlerter()

	// 制造 5 条告警：CPU critical → CPU recovery → Disk critical → Disk recovery → CPU warning
	a.Evaluate(&MetricSnapshot{CPUPercent: 85, MemoryPercent: 40, DiskPercent: 50})
	a.Evaluate(&MetricSnapshot{CPUPercent: 40, MemoryPercent: 40, DiskPercent: 50})
	a.Evaluate(&MetricSnapshot{CPUPercent: 40, MemoryPercent: 40, DiskPercent: 95})
	a.Evaluate(&MetricSnapshot{CPUPercent: 40, MemoryPercent: 40, DiskPercent: 30})
	a.Evaluate(&MetricSnapshot{CPUPercent: 70, MemoryPercent: 40, DiskPercent: 50})

	if n := len(a.GetAlerts(2)); n != 2 {
		t.Errorf("GetAlerts(2) 应返回 2 条，得到 %d 条", n)
	}
	if n := len(a.GetAlerts(10)); n != 5 {
		t.Errorf("GetAlerts(10) 应返回全部 5 条，得到 %d 条", n)
	}
	if n := len(a.GetAlerts(0)); n != 0 {
		t.Errorf("GetAlerts(0) 应返回 0 条，得到 %d 条", n)
	}
}

// ── 并发安全 ──

func TestAlerter_Concurrent(t *testing.T) {
	a := NewAlerter()
	var wg sync.WaitGroup

	// 并发写入 Evaluate
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(v float64) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				a.Evaluate(&MetricSnapshot{CPUPercent: v, MemoryPercent: 40, DiskPercent: 50})
			}
		}(float64(30 + i))
	}

	// 并发读取 GetAlerts
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				a.GetAlerts(10)
			}
		}()
	}

	wg.Wait()
	// 不 panic / 不 race 即通过
}
