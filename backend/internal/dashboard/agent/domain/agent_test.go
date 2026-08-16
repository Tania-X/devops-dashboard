package domain

import "testing"

func TestNewAgent(t *testing.T) {
	a := NewAgent("web-1", "10.0.0.1", 22, "root", "password", "/opt/agent", 9100)
	if a.ID == "" {
		t.Error("ID 不应为空(工厂应生成)")
	}
	if a.Status != AgentStatusUnknown {
		t.Errorf("初始状态应为 unknown, got %q", a.Status)
	}
	if a.Name != "web-1" || a.Host != "10.0.0.1" || a.AgentPort != 9100 {
		t.Errorf("字段赋值不正确: %+v", a)
	}
}

func TestAgentTarget_CheckStoppable(t *testing.T) {
	tests := []struct {
		name   string
		status string
		wantOK bool
	}{
		{"unknown 可停止", AgentStatusUnknown, true},
		{"online 可停止", AgentStatusOnline, true},
		{"offline 拒绝重复停止", AgentStatusOffline, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAgent("x", "h", 22, "u", "p", "/opt", 9100)
			a.Status = tt.status
			err := a.CheckStoppable()
			if (err == nil) != tt.wantOK {
				t.Errorf("CheckStoppable(%q) err=%v, wantOK=%v", tt.status, err, tt.wantOK)
			}
		})
	}
}

func TestAgentTarget_StateTransitions(t *testing.T) {
	a := NewAgent("x", "h", 22, "u", "p", "/opt", 9100)

	a.MarkDeployed()
	if a.Status != AgentStatusUnknown {
		t.Errorf("MarkDeployed 后应为 unknown, got %q", a.Status)
	}

	a.MarkOnline()
	if a.Status != AgentStatusOnline {
		t.Errorf("MarkOnline 后应为 online, got %q", a.Status)
	}

	a.MarkOffline()
	if a.Status != AgentStatusOffline {
		t.Errorf("MarkOffline 后应为 offline, got %q", a.Status)
	}

	// offline 后不可重复停止(状态机闭环校验)
	if err := a.CheckStoppable(); err == nil {
		t.Error("offline 后重复停止应报错")
	}
}
