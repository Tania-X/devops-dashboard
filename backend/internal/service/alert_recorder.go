package service

import (
	"log/slog"
	"sync"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"gorm.io/gorm"
)

// AlertRecorder 告警历史落库器(异步,不阻塞采集)。
// 设计:Alerter.OnAlert 同步回调中只做 channel 投递(ch 满则丢弃,告警内存缓冲已有),
// 独立 goroutine 批量落库;查询走 List(分页 + 级别筛选)。
// 生命周期:Close() 幂等,停止消费 goroutine 并等待其退出(应用关闭时调用,防 goroutine 泄漏)。
type AlertRecorder struct {
	db     *gorm.DB
	ch     chan model.AlertItem
	stopCh chan struct{}
	wg     sync.WaitGroup
	once   sync.Once
}

// NewAlertRecorder 创建落库器并启动消费 goroutine
func NewAlertRecorder(db *gorm.DB) *AlertRecorder {
	r := &AlertRecorder{
		db:     db,
		ch:     make(chan model.AlertItem, 128),
		stopCh: make(chan struct{}),
	}
	r.wg.Add(1)
	go r.run()
	return r
}

// Record 非阻塞投递一条告警到落库队列(关闭后或满则丢弃,不阻塞采集主链路)
func (r *AlertRecorder) Record(e model.AlertItem) {
	select {
	case r.ch <- e:
	case <-r.stopCh: // 已关闭:丢弃
	default: // 队列满:丢弃(内存告警缓冲已保留,查询仪表盘仍可见)
	}
}

// Close 停止消费 goroutine 并等待其退出(幂等;退出前尽力消费队列中剩余告警)。
// 不关闭 ch,避免并发 Record 向已关闭 channel 发送 panic。
func (r *AlertRecorder) Close() {
	r.once.Do(func() {
		close(r.stopCh)
		r.wg.Wait()
	})
}

func (r *AlertRecorder) run() {
	defer r.wg.Done()
	for {
		select {
		case e := <-r.ch:
			r.save(e)
		case <-r.stopCh:
			// 退出前尽力消费剩余告警(非阻塞 drain)
			for {
				select {
				case e := <-r.ch:
					r.save(e)
				default:
					return
				}
			}
		}
	}
}

func (r *AlertRecorder) save(e model.AlertItem) {
	alert := model.Alert{
		Level:   e.Level,
		Message: e.Message,
		Source:  e.Source,
		Time:    e.Time,
	}
	if err := r.db.Create(&alert).Error; err != nil {
		slog.Error("告警落库失败", "err", err)
	}
}

// List 分页查询告警历史(按落库时间倒序,最新在前);level 为空表示全部级别
func (r *AlertRecorder) List(page, pageSize int, level string) ([]model.Alert, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	q := r.db.Model(&model.Alert{})
	if level != "" {
		q = q.Where("level = ?", level)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.Alert
	if err := q.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
