package domain

import (
	"testing"
	"time"
)

func TestNewDeployment(t *testing.T) {
	d, err := NewDeployment("api-gateway", "v2.1.0", "prod")
	if err != nil {
		t.Fatalf("NewDeployment() error = %v", err)
	}
	if d.ID == "" {
		t.Error("ID 应为非空（工厂生成 uuid）")
	}
	if d.Status != DeploymentStatusPending {
		t.Errorf("初始状态 = %q, want %q", d.Status, DeploymentStatusPending)
	}
	if d.HistoryCount() != 0 {
		t.Errorf("初始历史数 = %d, want 0", d.HistoryCount())
	}
}

func TestNewDeployment_Validation(t *testing.T) {
	cases := []struct {
		name, appName, version, env string
	}{
		{"空应用名", "", "v1.0.0", "prod"},
		{"空版本号", "api-gateway", "", "prod"},
		{"空环境", "api-gateway", "v1.0.0", ""},
	}
	for _, c := range cases {
		if _, err := NewDeployment(c.appName, c.version, c.env); err == nil {
			t.Errorf("%s: 应返回错误", c.name)
		}
	}
}

func TestAddHistory_SyncsRootState(t *testing.T) {
	d, _ := NewDeployment("api-gateway", "v2.1.0", "prod")
	at := time.Now()

	if err := d.AddHistory("v2.1.0", "operator-1", 120, HistoryStatusSuccess, at); err != nil {
		t.Fatalf("AddHistory() error = %v", err)
	}
	// 不变式：根状态/版本/时间与最新历史同步
	if d.Status != HistoryStatusSuccess {
		t.Errorf("根状态 = %q, want %q（与最新历史一致）", d.Status, HistoryStatusSuccess)
	}
	if d.Version != "v2.1.0" {
		t.Errorf("根版本 = %q, want v2.1.0", d.Version)
	}
	if !d.LastDeployedAt.Equal(at) {
		t.Errorf("LastDeployedAt 未同步, got %v want %v", d.LastDeployedAt, at)
	}

	// 追加失败历史后，根状态跟随变为 failed
	if err := d.AddHistory("v2.1.1", "operator-2", 30, HistoryStatusFailed, at.Add(time.Hour)); err != nil {
		t.Fatalf("AddHistory() error = %v", err)
	}
	if d.Status != HistoryStatusFailed {
		t.Errorf("根状态 = %q, want %q", d.Status, HistoryStatusFailed)
	}
	latest := d.LatestHistory()
	if latest == nil || latest.Status != HistoryStatusFailed {
		t.Error("LatestHistory() 应返回最新一条 failed 历史")
	}
	if d.HistoryCount() != 2 {
		t.Errorf("历史数 = %d, want 2", d.HistoryCount())
	}
}

func TestAddHistory_Validation(t *testing.T) {
	d, _ := NewDeployment("api-gateway", "v2.1.0", "prod")
	at := time.Now()

	if err := d.AddHistory("", "op", 1, HistoryStatusSuccess, at); err == nil {
		t.Error("空版本号应报错")
	}
	if err := d.AddHistory("v2.1.0", "", 1, HistoryStatusSuccess, at); err == nil {
		t.Error("空操作人应报错")
	}
	if err := d.AddHistory("v2.1.0", "op", 1, "unknown", at); err == nil {
		t.Error("非法历史状态应报错")
	}
	if d.HistoryCount() != 0 {
		t.Errorf("校验失败不应产生历史, got %d", d.HistoryCount())
	}
}

func TestChangeStatus(t *testing.T) {
	d, _ := NewDeployment("api-gateway", "v2.1.0", "prod")

	if err := d.ChangeStatus(DeploymentStatusDeploying); err != nil {
		t.Fatalf("ChangeStatus(deploying) error = %v", err)
	}
	if d.Status != DeploymentStatusDeploying {
		t.Errorf("状态 = %q, want deploying", d.Status)
	}
	if err := d.ChangeStatus("broken"); err == nil {
		t.Error("非法状态应报错")
	}
	// 终态必须经 AddHistory 写入（内部同步根状态与最新历史），直接 ChangeStatus 应被拒绝
	for _, final := range []string{DeploymentStatusSuccess, DeploymentStatusFailed} {
		if err := d.ChangeStatus(final); err == nil {
			t.Errorf("ChangeStatus(%s) 应报错：终态必须经 AddHistory 写入", final)
		}
	}
}

// TestChangeStatus_NewRound:终态后进入 deploying 表示"新一轮发布开始"(合法中间态)。
// 根状态进入 deploying、最新历史保留上一轮结果——deploying 下与最新历史不同步是合法语义。
func TestChangeStatus_NewRound(t *testing.T) {
	d, _ := NewDeployment("api-gateway", "v2.1.0", "prod")
	at := time.Now()
	if err := d.AddHistory("v2.1.0", "operator-1", 120, HistoryStatusSuccess, at); err != nil {
		t.Fatalf("AddHistory() error = %v", err)
	}

	// 新一轮发布开始：终态 → deploying（合法）
	if err := d.ChangeStatus(DeploymentStatusDeploying); err != nil {
		t.Fatalf("终态后 ChangeStatus(deploying) 应合法, got %v", err)
	}
	if d.Status != DeploymentStatusDeploying {
		t.Errorf("根状态 = %q, want deploying", d.Status)
	}
	// 历史保留上一轮结果，LatestHistory 仍为 success
	if latest := d.LatestHistory(); latest == nil || latest.Status != HistoryStatusSuccess {
		t.Errorf("deploying 中间态下最新历史应保留上一轮 success, got %+v", latest)
	}
}

func TestLatestHistory_Empty(t *testing.T) {
	d, _ := NewDeployment("api-gateway", "v2.1.0", "prod")
	if got := d.LatestHistory(); got != nil {
		t.Errorf("无历史时 LatestHistory() = %v, want nil", got)
	}
}
