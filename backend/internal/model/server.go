package model

type Server struct {
	ID                string             `gorm:"primaryKey" json:"id"`
	Hostname          string             `json:"hostname"`
	IP                string             `json:"ip"`
	OS                string             `json:"os"`
	CpuCores          int                `json:"cpuCores"`
	MemoryGb          int                `json:"memoryGb"`
	Status            string             `json:"status"`
	Uptime            string             `json:"uptime"`
	DiskPartitions    []DiskPartition    `gorm:"foreignKey:ServerID" json:"diskPartitions,omitempty"`
	NetworkInterfaces []NetworkInterface `gorm:"foreignKey:ServerID" json:"networkInterfaces,omitempty"`
}

type DiskPartition struct {
	ID       uint   `gorm:"primaryKey" json:"-"`
	ServerID string `json:"-"`
	Mount    string `json:"mount"`
	TotalGb  int    `json:"totalGb"`
	UsedGb   int    `json:"usedGb"`
}

type NetworkInterface struct {
	ID       uint   `gorm:"primaryKey" json:"-"`
	ServerID string `json:"-"`
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Mac      string `json:"mac"`
}

type PagedResultServerItem struct {
	List     []Server `json:"list"`
	Total    int64    `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"pageSize"`
}
