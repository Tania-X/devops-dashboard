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
// 生命周期：Close() 幂等，停止消费 goroutine 并等待其退出（应用关闭时调用，防泄漏）。
type AlertBus struct {
	mu        sync.RWMutex
	ch        chan model.AlertItem
	notifiers []Notifier
	stopCh    chan struct{}
	wg        sync.WaitGroup
	once      sync.Once
}

// NewAlertBus 创建告警总线，notifiers 为所有已注册渠道
func NewAlertBus(notifiers ...Notifier) *AlertBus {
	return &AlertBus{
		ch:        make(chan model.AlertItem, 100),
		notifiers: notifiers,
		stopCh:    make(chan struct{}),
	}
}

// Publish 向总线投递一条告警（非阻塞）
// channel 满时丢弃并记 warning，避免阻塞采集主链路；关闭后直接丢弃
func (b *AlertBus) Publish(entry model.AlertItem) {
	select {
	case b.ch <- entry:
	case <-b.stopCh: // 已关闭：丢弃
	default:
		slog.Warn("告警总线已满，丢弃告警", "alert", entry.ID, "level", entry.Level)
	}
}

// Run 启动后台消费协程：循环从 channel 取告警，分发给所有 Notifier
// 单个 Notifier 发送失败不影响其他渠道
func (b *AlertBus) Run() {
	b.wg.Add(1)
	go b.run()
}

// Close 停止消费 goroutine 并等待其退出（幂等；退出前尽力消费队列中剩余告警）。
// 不关闭 ch，避免并发 Publish 向已关闭 channel 发送 panic。
func (b *AlertBus) Close() {
	b.once.Do(func() {
		close(b.stopCh)
		b.wg.Wait()
	})
}

func (b *AlertBus) run() {
	defer b.wg.Done()
	for {
		select {
		case entry := <-b.ch:
			b.dispatch(entry)
		case <-b.stopCh:
			// 退出前尽力消费剩余告警（非阻塞 drain）
			for {
				select {
				case entry := <-b.ch:
					b.dispatch(entry)
				default:
					return
				}
			}
		}
	}
}

func (b *AlertBus) dispatch(entry model.AlertItem) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, n := range b.notifiers {
		if err := n.Send(entry); err != nil {
			slog.Error("告警推送失败", "notifier", n.Name(), "alert", entry.ID, "err", err)
		}
	}
}

// SetNotifiers 原子替换渠道列表（配置热更新时调用）
func (b *AlertBus) SetNotifiers(notifiers []Notifier) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.notifiers = notifiers
}
