package notify

import (
	"sync"
	"testing"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
)

// mockNotifier 记录收到的告警数
type mockNotifier struct {
	mu    sync.Mutex
	count int
}

func (m *mockNotifier) Name() string { return "mock" }
func (m *mockNotifier) Send(e model.AlertItem) error {
	m.mu.Lock()
	m.count++
	m.mu.Unlock()
	return nil
}

func (m *mockNotifier) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.count
}

func TestAlertBus_Close(t *testing.T) {
	t.Run("Close 幂等且不 panic", func(t *testing.T) {
		b := NewAlertBus()
		b.Run()
		b.Close()
		b.Close() // 第二次应安全
	})

	t.Run("Close 后 Publish 不 panic 不阻塞", func(t *testing.T) {
		b := NewAlertBus()
		b.Run()
		b.Close()
		b.Publish(model.AlertItem{ID: "x", Level: "info"}) // 关闭后投递应被丢弃
	})

	t.Run("正常消费分发到 notifier", func(t *testing.T) {
		n := &mockNotifier{}
		b := NewAlertBus(n)
		b.Run()
		b.Publish(model.AlertItem{ID: "a1", Level: "warning"})
		b.Close() // 等待消费完成(drain)
		if n.Count() != 1 {
			t.Fatalf("notifier 应收到 1 条告警, got %d", n.Count())
		}
	})

	t.Run("SetNotifiers 热更新生效", func(t *testing.T) {
		n1, n2 := &mockNotifier{}, &mockNotifier{}
		b := NewAlertBus(n1)
		b.Run()
		b.SetNotifiers([]Notifier{n2})
		b.Publish(model.AlertItem{ID: "a2", Level: "critical"})
		b.Close()
		if n1.Count() != 0 || n2.Count() != 1 {
			t.Fatalf("热更新后应只发给 n2, got n1=%d n2=%d", n1.Count(), n2.Count())
		}
	})
}
