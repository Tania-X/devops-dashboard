package model

type MetricValue struct {
	Current float64 `json:"current"`
	Status  string  `json:"status"`
}

type DashboardMetrics struct {
	CPU        MetricValue `json:"cpu"`
	Memory     MetricValue `json:"memory"`
	Disk       MetricValue `json:"disk"`
	AlertCount int         `json:"alertCount"`
}

type DashboardTrend struct {
	TimeLabels []string  `json:"timeLabels"`
	CpuData    []float64 `json:"cpuData"`
	MemoryData []float64 `json:"memoryData"`
}

type AlertItem struct {
	ID      string `json:"id"`
	Level   string `json:"level"`
	Message string `json:"message"`
	Source  string `json:"source"`
	Time    string `json:"time"`
}
