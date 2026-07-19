package repository

import (
	"log/slog"
	"os"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dbPath string) *gorm.DB {
	if dbPath == "" {
		dbPath = "storage/devops.db"
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		slog.Error("数据库连接失败", "error", err)
		os.Exit(1)
	}

	slog.Info("数据库已连接", "path", dbPath)

	err = db.AutoMigrate(
		&model.Server{},
		&model.DiskPartition{},
		&model.NetworkInterface{},
		&model.Log{},
		&model.Deployment{},
		&model.DeploymentHistory{},
	)
	if err != nil {
		slog.Error("AutoMigrate 失败", "error", err)
		os.Exit(1)
	}

	slog.Info("数据库表结构已同步")

	DB = db
	return db
}
