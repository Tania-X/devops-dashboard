package monitor

import (
	"testing"
)

func TestNewRemoteCollector(t *testing.T) {
	cases := []struct {
		name      string
		agentAddr string
		wantURL   string
	}{
		{"ip_with_port", "192.168.1.100:9100", "http://192.168.1.100:9100"},
		{"localhost", "localhost:9100", "http://localhost:9100"},
		{"ip_only", "10.0.0.1:8080", "http://10.0.0.1:8080"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rc := NewRemoteCollector(tc.agentAddr)
			if rc.baseURL != tc.wantURL {
				t.Errorf("baseURL = %q, want %q", rc.baseURL, tc.wantURL)
			}
			if rc.client == nil {
				t.Error("client 不应为 nil")
			}
		})
	}
}
