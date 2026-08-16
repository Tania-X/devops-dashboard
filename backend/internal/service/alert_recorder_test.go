package service

import (
	"strings"
	"testing"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newRecorderTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	// 每个子测试独立内存库(共享缓存库会导致数据串扰)
	name := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试库失败: %v", err)
	}
	if err := db.AutoMigrate(&model.Alert{}); err != nil {
		t.Fatalf("AutoMigrate 失败: %v", err)
	}
	return db
}

func TestAlertRecorder_Close(t *testing.T) {
	t.Run("Close 幂等且不 panic", func(t *testing.T) {
		r := NewAlertRecorder(newRecorderTestDB(t))
		r.Close()
		r.Close() // 第二次调用应安全
	})

	t.Run("Close 后 Record 不 panic", func(t *testing.T) {
		r := NewAlertRecorder(newRecorderTestDB(t))
		r.Close()
		// 关闭后继续投递不应 panic(select 走 stopCh 分支丢弃)
		r.Record(model.AlertItem{Level: "info", Message: "x"})
	})

	t.Run("Close 等待消费完成", func(t *testing.T) {
		db := newRecorderTestDB(t)
		r := NewAlertRecorder(db)
		// 投递若干条后立即 Close,应等消费 goroutine 排空队列
		for i := 0; i < 5; i++ {
			r.Record(model.AlertItem{Level: "warning", Message: "test"})
		}
		r.Close() // 会等待 run() 退出(drain 剩余)
		time.Sleep(50 * time.Millisecond) // 留出落库时间

		var count int64
		db.Model(&model.Alert{}).Count(&count)
		if count == 0 {
			t.Fatal("Close 前投递的告警应已被消费落库")
		}
	})

	t.Run("正常消费后落库", func(t *testing.T) {
		db := newRecorderTestDB(t)
		r := NewAlertRecorder(db)
		r.Record(model.AlertItem{Level: "critical", Message: "cpu high", Source: "localhost", Time: "08-16 10:00"})
		time.Sleep(100 * time.Millisecond) // 等消费

		var list []model.Alert
		db.Find(&list)
		if len(list) != 1 || list[0].Message != "cpu high" || list[0].Level != "critical" {
			t.Fatalf("应落库 1 条 cpu high/critical, got %+v", list)
		}
	})
}
