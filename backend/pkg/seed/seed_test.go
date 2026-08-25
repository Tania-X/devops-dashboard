package seed

import (
	"testing"
	"time"

	deploymentdomain "github.com/Tania-X/devops-dashboard/backend/internal/dashboard/deployment/domain"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// TestSeedDeployments_RootStateSyncsLatestHistory 验证 seed 生成的部署数据满足领域语义：
// 根状态/LastDeployedAt 必须与「时间最新」的一次部署一致（而非最旧一次）。
// 回归背景：历史曾按"最新→最旧"顺序追加，AddHistory 每次同步根状态，导致
// 循环结束后根状态指向最旧一次部署，与领域注释相悖（CodeRabbit 第 3 次评审）。
func TestSeedDeployments_RootStateSyncsLatestHistory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("内存 DB 打开失败: %v", err)
	}
	if err := db.AutoMigrate(&deploymentdomain.Deployment{}, &deploymentdomain.DeploymentHistory{}); err != nil {
		t.Fatalf("AutoMigrate 失败: %v", err)
	}

	seedDeployments(db)

	var deployments []deploymentdomain.Deployment
	if err := db.Preload("Histories").Find(&deployments).Error; err != nil {
		t.Fatalf("查询部署失败: %v", err)
	}
	if len(deployments) != 15 {
		t.Fatalf("期望 15 条部署, got %d", len(deployments))
	}

	for _, d := range deployments {
		// 找出时间最新的一条历史（不依赖加载顺序）
		var latest *deploymentdomain.DeploymentHistory
		for i := range d.Histories {
			h := &d.Histories[i]
			if latest == nil || h.DeployedAt.After(latest.DeployedAt) {
				latest = h
			}
		}
		if latest == nil {
			t.Fatalf("部署 %s 无历史", d.ID)
		}
		// deploying 是"新一轮发布中"的合法中间态（最新历史保留上一轮结果），跳过终态断言
		if d.Status == deploymentdomain.DeploymentStatusDeploying {
			continue
		}
		if d.Status != latest.Status {
			t.Errorf("部署 %s: 根状态 %q 与最新历史 %q 不一致", d.ID, d.Status, latest.Status)
		}
		if d.LastDeployedAt.Sub(latest.DeployedAt).Abs() > time.Second {
			t.Errorf("部署 %s: LastDeployedAt %v 与最新历史 %v 不一致", d.ID, d.LastDeployedAt, latest.DeployedAt)
		}
	}
}
