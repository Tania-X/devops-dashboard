# DevOps Dashboard Backend

Go + Gin + GORM + SQLite (pure Go, no CGO)

## 运行

```bash
cd backend
go mod tidy
go run cmd/api/main.go
```

服务默认启动在 http://localhost:8080

测试接口：
```bash
curl http://localhost:8080/api/servers
curl http://localhost:8080/api/dashboard/metrics
```

## 重新生成数据

删除 `storage/devops.db`，重新启动服务即可自动重新 seed。

## 构建生产包

```bash
cd backend
go build -o devops-api cmd/api/main.go
```

Windows 下会生成 `devops-api.exe`，双击即可运行。
