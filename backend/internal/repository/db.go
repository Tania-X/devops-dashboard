package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// InitDB 初始化 SQLite 数据库连接并自动迁移表结构
// 数据库文件所在目录不存在时会自动创建
func InitDB(dbPath string) (*gorm.DB, error) {
	if dbPath == "" {
		dbPath = "storage/devops.db"
	}

	// 自动创建数据库文件所在目录（SQLite 不会自动建父目录）
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	if err := db.AutoMigrate(
		&model.Server{},
		&model.DiskPartition{},
		&model.NetworkInterface{},
		&model.Log{},
		&model.Deployment{},
		&model.DeploymentHistory{},
	); err != nil {
		return nil, fmt.Errorf("AutoMigrate 失败: %w", err)
	}

	return db, nil
}
