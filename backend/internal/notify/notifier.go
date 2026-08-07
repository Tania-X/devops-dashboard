package notify

import (
	"log/slog"
	"sync"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
)

// Notifier 通知渠道抽象接口
// 实现类：WebhookNotifier（钉钉/企微），未来可扩展 EmailNotifier 等
type Notifier interface {
	Name() string
	Send(entry model.AlertItem) error
}

// AlertBus 告警总线：生产者（Alerter）→ channel → 各 Notifier
//
// 设计目的：把「告警产生」和「通知发送」解耦。
// Alerter.Evaluate 是采集主链路，不能因为 HTTP 发送慢/超时而阻塞采集，
// 所以 Publish 只是往 channel 里丢一条（微秒级），真正的发送由后台
// goroutine 异步完成。channel 满时丢弃新告警（记录日志），绝不阻塞生产者。
type AlertBus struct {
	mu        sync.RWMutex
	ch        chan model.AlertItem
	notifiers []Notifier
}

// NewAlertBus 创建告警总线，notifiers 为所有已注册渠道
func NewAlertBus(notifiers ...Notifier) *AlertBus {
	return &AlertBus{
		ch:        make(chan model.AlertItem, 100),
		notifiers: notifiers,
	}
}

// Publish 向总线投递一条告警（非阻塞）
// channel 满时丢弃并记 warning，避免阻塞采集主链路
func (b *AlertBus) Publish(entry model.AlertItem) {
	select {
	case b.ch <- entry:
	default:
		slog.Warn("告警总线已满，丢弃告警", "alert", entry.ID, "level", entry.Level)
	}
}

// Run 启动后台消费协程：循环从 channel 取告警，分发给所有 Notifier
// 单个 Notifier 发送失败不影响其他渠道
func (b *AlertBus) Run() {
	go func() {
		for entry := range b.ch {
			b.mu.RLock()
			for _, n := range b.notifiers {
				if err := n.Send(entry); err != nil {
					slog.Error("告警推送失败", "notifier", n.Name(), "alert", entry.ID, "err", err)
				}
			}
			b.mu.RUnlock()
		}
	}()
}

// SetNotifiers 原子替换渠道列表（配置热更新时调用）
func (b *AlertBus) SetNotifiers(notifiers []Notifier) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.notifiers = notifiers
}
