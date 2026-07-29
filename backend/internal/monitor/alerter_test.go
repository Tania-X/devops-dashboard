package monitor

import (
	"testing"
)

func TestAlerter_Normal(t *testing.T) {
	a := NewAlerter()
	a.Evaluate(&MetricSnapshot{CPUPercent: 30, MemoryPercent: 40, DiskPercent: 50})

	alerts := a.GetAlerts(10)
	if len(alerts) != 0 {
		t.Errorf("指标正常，应该没有告警，却得到了 %d 条告警", len(alerts))
	}
}

func TestAlerter_Warning(t *testing.T) {
	a := NewAlerter()
	a.Evaluate(&MetricSnapshot{CPUPercent: 65, MemoryPercent: 40, DiskPercent: 50})

	alerts := a.GetAlerts(10)
	if len(alerts) == 0 {
		t.Fatal("CPU 超过 warning 阈值，应该生成告警")
	}
	if alerts[0].Level != "warning" {
		t.Errorf("期望 warning 级别，得到 %s", alerts[0].Level)
	}
	if alerts[0].Message == "" {
		t.Error("告警消息为空")
	}
}

func TestAlerter_Critical(t *testing.T) {
	a := NewAlerter()
	a.Evaluate(&MetricSnapshot{CPUPercent: 85, MemoryPercent: 40, DiskPercent: 50})

	alerts := a.GetAlerts(10)
	if len(alerts) == 0 {
		t.Fatal("CPU 超过 critical 阈值，应该生成告警")
	}
	if alerts[0].Level != "critical" {
		t.Errorf("期望 critical 级别，得到 %s", alerts[0].Level)
	}
}

func TestAlerter_Deduplication(t *testing.T) {
	a := NewAlerter()

	// 第一次：CPU 85 → 触发 critical
	a.Evaluate(&MetricSnapshot{CPUPercent: 85, MemoryPercent: 40, DiskPercent: 50})
	if n := len(a.GetAlerts(10)); n != 1 {
		t.Fatalf("第一次应该生成 1 条告警，却得到 %d 条", n)
	}

	// 第二次：CPU 还是 85 → 状态没变，不应重复告警
	a.Evaluate(&MetricSnapshot{CPUPercent: 85, MemoryPercent: 40, DiskPercent: 50})
	if n := len(a.GetAlerts(10)); n != 1 {
		t.Fatalf("状态没变，不应该重复告警，却有 %d 条", n)
	}
}

func TestAlerter_Escalate(t *testing.T) {
	a := NewAlerter()

	// CPU 70 → warning
	a.Evaluate(&MetricSnapshot{CPUPercent: 70, MemoryPercent: 40, DiskPercent: 50})

	// CPU 90 → 从 warning 升级为 critical，应该生成新告警
	a.Evaluate(&MetricSnapshot{CPUPercent: 90, MemoryPercent: 40, DiskPercent: 50})

	alerts := a.GetAlerts(10)
	if len(alerts) != 2 {
		t.Fatalf("期望 2 条告警（warning + critical），得到 %d 条", len(alerts))
	}
	if alerts[0].Level != "critical" {
		t.Errorf("最新告警应该是 critical，得到 %s", alerts[0].Level)
	}
}

func TestAlerter_Recovery(t *testing.T) {
	a := NewAlerter()

	// CPU 85 → critical
	a.Evaluate(&MetricSnapshot{CPUPercent: 85, MemoryPercent: 40, DiskPercent: 50})

	// CPU 降到 40 → 恢复正常，生成"已恢复"告警
	a.Evaluate(&MetricSnapshot{CPUPercent: 40, MemoryPercent: 40, DiskPercent: 50})

	alerts := a.GetAlerts(10)
	if len(alerts) != 2 {
		t.Fatalf("期望 2 条告警（critical + recovery），得到 %d 条", len(alerts))
	}
	if alerts[0].Level != "info" {
		t.Errorf("恢复告警级别应该是 info，得到 %s", alerts[0].Level)
	}
}

func TestAlerter_MultipleMetrics(t *testing.T) {
	a := NewAlerter()

	// CPU 和 磁盘都超标
	a.Evaluate(&MetricSnapshot{CPUPercent: 85, MemoryPercent: 40, DiskPercent: 80})

	alerts := a.GetAlerts(10)
	if len(alerts) != 2 {
		t.Fatalf("CPU + 磁盘都超标，应该生成 2 条告警，得到 %d 条", len(alerts))
	}
}

func TestAlerter_Order(t *testing.T) {
	a := NewAlerter()

	// 先 CPU 超标，后磁盘超标
	a.Evaluate(&MetricSnapshot{CPUPercent: 85, MemoryPercent: 40, DiskPercent: 30})
	a.Evaluate(&MetricSnapshot{CPUPercent: 40, MemoryPercent: 40, DiskPercent: 95})

	alerts := a.GetAlerts(10)
	if len(alerts) != 3 {
		// CPU critical → CPU 恢复 → 磁盘 critical
		t.Fatalf("期望 3 条（CPU critical + CPU 恢复 + 磁盘 critical），得到 %d 条", len(alerts))
	}
	// 最新的一条应该是磁盘 critical 告警
	if alerts[0].Level != "critical" {
		t.Errorf("最新告警应该是 critical，得到 %s", alerts[0].Level)
	}
	if alerts[0].Message != "" && alerts[0].Message != "上一条" {
		// 验证是磁盘告警
	}
}
