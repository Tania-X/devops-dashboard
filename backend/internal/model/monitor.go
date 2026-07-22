package model

// ProcessItem 进程列表项
type ProcessItem struct {
	PID           int     `json:"pid"`
	Name          string  `json:"name"`
	Cmdline       string  `json:"cmdline,omitempty"`
	CPUPercent    float64 `json:"cpuPercent"`
	MemoryPercent float64 `json:"memoryPercent"`
	MemoryMb      float64 `json:"memoryMb"`
	Status        string  `json:"status"`
}

// ProcessDetail 进程详情
type ProcessDetail struct {
	ProcessItem
	PPid           int      `json:"ppid"`
	NumThreads     int      `json:"numThreads"`
	NumOpenFiles   int      `json:"numOpenFiles"`
	NumConnections int      `json:"numConnections"`
	CreateTime     string   `json:"createTime"`
	Username       string   `json:"username"`
	WorkingDir     string   `json:"workingDir,omitempty"`
	Env            []string `json:"env,omitempty"`
}

// HostInfo 主机信息
type HostInfo struct {
	Hostname        string  `json:"hostname"`
	OS              string  `json:"os"`
	Platform        string  `json:"platform"`
	PlatformVersion string  `json:"platformVersion"`
	KernelVersion   string  `json:"kernelVersion"`
	Arch            string  `json:"arch"`
	BootTime        string  `json:"bootTime"`
	Uptime          string  `json:"uptime"`
	CPUModel        string  `json:"cpuModel"`
	CPUCores        int     `json:"cpuCores"`
	CPULogicalCores int     `json:"cpuLogicalCores"`
	TotalMemoryGb   float64 `json:"totalMemoryGb"`
	VirtualMemoryGb float64 `json:"virtualMemoryGb,omitempty"`
}
