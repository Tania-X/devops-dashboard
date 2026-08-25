package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Deployment 部署状态（集中定义，替代散落的魔法字符串）
const (
	DeploymentStatusPending   = "pending"
	DeploymentStatusDeploying = "deploying"
	DeploymentStatusSuccess   = "success"
	DeploymentStatusFailed    = "failed"
)

// DeploymentHistoryStatus 部署历史状态
const (
	HistoryStatusSuccess = "success"
	HistoryStatusFailed  = "failed"
)

// Deployment 部署聚合根（充血模型）。
//
// 聚合边界 = Deployment + DeploymentHistory：
//   - 历史只能经 AddHistory 追加，根状态与最新一次部署结果保持同步（不变式）
//   - ChangeStatus 仅用于部署流程中间态（pending → deploying），终态一律经 AddHistory
//   - 外部不得直接修改 Histories
//
// 说明：Histories 因 GORM 关联持久化（Preload/外键保存）需要保持导出字段；
// 业务上必须经 AddHistory 变更，只读访问走 HistoryCount / LatestHistory。
// 待演进阶段②（Repository 接口 + 组装职责下沉）后可将子实体私有化。
type Deployment struct {
	ID             string              `gorm:"primaryKey" json:"id"`
	AppName        string              `json:"appName"`
	Version        string              `json:"version"`
	Env            string              `json:"env"`
	Status         string              `json:"status"`
	LastDeployedAt time.Time           `json:"lastDeployedAt"`
	Histories      []DeploymentHistory `gorm:"foreignKey:DeploymentID" json:"-"`
}

// DeploymentHistory 部署历史子实体。
// 无独立对外身份：仅存在于 Deployment 聚合内，不可脱离根单独创建/修改。
type DeploymentHistory struct {
	ID           uint      `gorm:"primaryKey" json:"-"`
	DeploymentID string    `json:"-"`
	Version      string    `json:"version"`
	Operator     string    `json:"operator"`
	DurationSec  int       `json:"durationSec"`
	Status       string    `json:"status"`
	DeployedAt   time.Time `json:"deployedAt"`
}

// NewDeployment 工厂：创建聚合根，初始状态 pending、无历史（未落库）。
func NewDeployment(appName, version, env string) (*Deployment, error) {
	if appName == "" {
		return nil, errors.New("应用名不能为空")
	}
	if version == "" {
		return nil, errors.New("版本号不能为空")
	}
	if env == "" {
		return nil, errors.New("环境不能为空")
	}
	return &Deployment{
		ID:      uuid.New().String(),
		AppName: appName,
		Version: version,
		Env:     env,
		Status:  DeploymentStatusPending,
	}, nil
}

// AddHistory 新增一次部署结果：校验不变式 → 同步根状态/版本/时间 → 追加历史。
// 两条写操作在这里绑定为一次业务原子操作，调用方事务内保存即可，无中间态窗口。
func (d *Deployment) AddHistory(version, operator string, durationSec int, status string, deployedAt time.Time) error {
	if version == "" {
		return errors.New("版本号不能为空")
	}
	if operator == "" {
		return errors.New("操作人不能为空")
	}
	if status != HistoryStatusSuccess && status != HistoryStatusFailed {
		return fmt.Errorf("非法历史状态: %q", status)
	}
	// 不变式：根状态与最新一次部署结果一致；Version/LastDeployedAt 同步
	d.Status = status
	d.Version = version
	if !deployedAt.IsZero() {
		d.LastDeployedAt = deployedAt
	}
	d.Histories = append(d.Histories, DeploymentHistory{
		DeploymentID: d.ID,
		Version:      version,
		Operator:     operator,
		DurationSec:  durationSec,
		Status:       status,
		DeployedAt:   deployedAt,
	})
	return nil
}

// ChangeStatus 变更根状态（仅限部署流程中间态：pending → deploying）。
// 终态（success/failed）一律由 AddHistory 写入——AddHistory 内部会同步根状态与
// 最新历史；绕过它直接写终态会使根状态与最新历史脱节（违反不变式）。
func (d *Deployment) ChangeStatus(status string) error {
	switch status {
	case DeploymentStatusPending, DeploymentStatusDeploying:
	default:
		return fmt.Errorf("非法状态: %q（终态必须经 AddHistory 写入）", status)
	}
	d.Status = status
	return nil
}

// HistoryCount 历史条数（只读）
func (d *Deployment) HistoryCount() int {
	return len(d.Histories)
}

// LatestHistory 最新一条历史（只读；无历史返回 nil）
func (d *Deployment) LatestHistory() *DeploymentHistory {
	if len(d.Histories) == 0 {
		return nil
	}
	h := d.Histories[len(d.Histories)-1]
	return &h
}
