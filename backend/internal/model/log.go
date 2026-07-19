package model

type Log struct {
	ID         string `gorm:"primaryKey" json:"id"`
	Time       string `json:"time"`
	Level      string `json:"level"`
	Service    string `json:"service"`
	Content    string `json:"content"`
	SourceHost string `json:"sourceHost"`
	LogPath    string `json:"logPath"`
	TraceID    string `json:"traceId"`
}

type PagedResultLogItem struct {
	List     []Log `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}
