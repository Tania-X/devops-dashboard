package service

// Services 聚合所有 Service，方便 Handler 一次性获取依赖
type Services struct {
	ServerService *ServerService
}
