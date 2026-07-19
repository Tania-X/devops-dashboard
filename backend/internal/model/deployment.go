package model

import "time"

type Deployment struct {
	ID             string              `gorm:"primaryKey" json:"id"`
	AppName        string              `json:"appName"`
	Version        string              `json:"version"`
	Env            string              `json:"env"`
	Status         string              `json:"status"`
	LastDeployedAt time.Time           `json:"lastDeployedAt"`
	Histories      []DeploymentHistory `gorm:"foreignKey:DeploymentID" json:"-"`
}

type DeploymentHistory struct {
	ID           uint      `gorm:"primaryKey" json:"-"`
	DeploymentID string    `json:"-"`
	Version      string    `json:"version"`
	Operator     string    `json:"operator"`
	DurationSec  int       `json:"durationSec"`
	Status       string    `json:"status"`
	DeployedAt   time.Time `json:"deployedAt"`
}
