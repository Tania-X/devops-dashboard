package monitor

import (
	"testing"
)

// TestStatus 演示 Go 的表驱动测试（Table-Driven Test）
// 这是 Go 社区最推荐的测试写法：一个函数覆盖多组场景
func TestStatus(t *testing.T) {
	cases := []struct {
		name  string
		input float64
		want  string
	}{
		{"normal 边界", 0.0, "normal"},
		{"normal 中间", 45.5, "normal"},
		{"normal 临界", 59.9, "normal"},
		{"warning 边界", 60.0, "warning"},
		{"warning 中间", 75.0, "warning"},
		{"warning 临界", 79.9, "warning"},
		{"critical 边界", 80.0, "critical"},
		{"critical 超高", 99.9, "critical"},
	}

	for _, c := range cases {
		// t.Run 创建子测试，失败时输出具体场景名，方便定位
		t.Run(c.name, func(t *testing.T) {
			got := Status(c.input)
			if got != c.want {
				// t.Errorf 不会中断后续测试，只是标记失败
				t.Errorf("Status(%.1f) = %q, want %q", c.input, got, c.want)
			}
		})
	}
}

// TestCollect 验证 gopsutil 在本机（macOS）能正确采集
// 这是白盒测试（package monitor），可以直接调用未导出的辅助函数
func TestCollect(t *testing.T) {
	snapshot, err := Collect()
	if err != nil {
		// 如果采集失败，直接终止此测试
		t.Fatalf("Collect() 失败: %v", err)
	}

	// 验证数值在合理范围内
	if snapshot.CPUPercent < 0 || snapshot.CPUPercent > 100 {
		t.Errorf("CPU 超出范围: %.2f%%", snapshot.CPUPercent)
	}
	if snapshot.MemoryPercent < 0 || snapshot.MemoryPercent > 100 {
		t.Errorf("Memory 超出范围: %.2f%%", snapshot.MemoryPercent)
	}
	if snapshot.DiskPercent < 0 || snapshot.DiskPercent > 100 {
		t.Errorf("Disk 超出范围: %.2f%%", snapshot.DiskPercent)
	}

	// 打印采集结果，方便肉眼确认（go test -v 时可见）
	t.Logf("✓ 采集成功: CPU=%.1f%% Memory=%.1f%% Disk=%.1f%%",
		snapshot.CPUPercent, snapshot.MemoryPercent, snapshot.DiskPercent)
}

// TestRound 测试辅助函数 Round
// 白盒测试的优势：可以验证内部实现的边界行为
func TestRound(t *testing.T) {
	cases := []struct {
		input float64
		want  float64
	}{
		{12.34, 12.3},
		{12.35, 12.4},
		{12.0, 12.0},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			got := Round(c.input)
			if got != c.want {
				t.Errorf("Round(%.2f) = %.1f, want %.1f", c.input, got, c.want)
			}
		})
	}
}
