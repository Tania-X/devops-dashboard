package service

import "testing"

func TestValidDeployDir(t *testing.T) {
	tests := []struct {
		dir    string
		wantOK bool
	}{
		{"/opt/agent", true},
		{"/data/agent-1", true},
		{"./agent", true},
		{"", false},                        // 空
		{"/opt/agent; rm -rf /", false},    // 命令注入
		{"$(whoami)/x", false},             // 命令替换
		{"`reboot`", false},                // 反引号
		{"/opt/my agent", false},           // 空格
		{"/opt/'x'", false},                // 引号
		{"-opt/agent", false},              // 以 - 开头
	}
	for _, tt := range tests {
		err := validDeployDir(tt.dir)
		if (err == nil) != tt.wantOK {
			t.Errorf("validDeployDir(%q) err=%v, wantOK=%v", tt.dir, err, tt.wantOK)
		}
	}
}
