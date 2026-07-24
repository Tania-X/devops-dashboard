# DevOps Dashboard Backend

Go + Gin + GORM + SQLite（纯 Go，无需 CGO）

## 快速开始

```bash
# 安装依赖
cd backend
go mod tidy

# 启动服务（默认端口 8080）
go run ./cmd/api

# 验证
curl http://localhost:8080/api/health
```

服务启动后自动建表并写入种子数据（35 台服务器、200 条日志、15 个部署记录）。

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `8080` | 服务监听端口 |
| `DB_PATH` | `storage/devops.db` | SQLite 数据库文件路径 |
| `LOG_LEVEL` | `info` | 日志级别：`debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `text` | 日志格式：`text`（人类可读）或 `json`（接入 Loki） |
| `ENV` | `dev` | 运行环境：`dev`, `prod` |
| `VERSION` | `dev` | 服务版本号（用于日志全局字段） |
| `HISTORY_RETAIN` | `24h` | 指标历史保留时长 |
| `HISTORY_INTERVAL` | `10s` | 指标采集间隔 |

示例：

```bash
# 生产模式：JSON 日志 + 环境标识
LOG_FORMAT=json ENV=prod VERSION=1.0.0 go run ./cmd/api
```

## API 接口

所有路由前缀 `/api`，共 13 个 endpoint：

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/health` | 健康检查（含数据库连通性检测） |
| GET | `/api/servers` | 服务器列表 |
| GET | `/api/servers/:id` | 单台服务器详情 |
| GET | `/api/dashboard/metrics` | Dashboard 实时指标（CPU/内存/磁盘） |
| GET | `/api/dashboard/trend` | 趋势历史数据 |
| GET | `/api/dashboard/alerts` | 告警列表 |
| GET | `/api/logs` | 日志列表（支持分页/级别/关键词筛选） |
| GET | `/api/deployments` | 部署列表 |
| GET | `/api/deployments/:id/history` | 单次部署历史 |
| GET | `/api/monitor/processes` | 本机进程列表 |
| GET | `/api/monitor/processes/:pid` | 单进程详情 |
| GET | `/api/monitor/host` | 本机主机信息 |

快速测试：

```bash
curl http://localhost:8080/api/health
curl http://localhost:8080/api/servers | head -c 200
curl http://localhost:8080/api/dashboard/metrics
```

## 项目结构

```
backend/
├── cmd/
│   └── api/
│       └── main.go          # 入口：读配置 → 初始化 → 启动 HTTP 服务
├── internal/
│   ├── api/
│   │   ├── router.go        # Gin 路由注册 + CORS/日志中间件
│   │   ├── errors.go        # 统一错误响应 + panic 恢复中间件
│   │   ├── server.go        # 服务器相关 handler
│   │   ├── dashboard.go     # Dashboard 看板 handler
│   │   ├── log.go           # 日志查询 handler
│   │   ├── deployment.go    # 部署管理 handler
│   │   ├── monitor.go       # 实时监控 handler
│   │   └── health.go        # 健康检查 handler
│   ├── app/
│   │   └── app.go           # 应用生命周期（Init/Run/Shutdown）
│   ├── config/
│   │   └── config.go        # 环境变量配置加载
│   ├── model/
│   │   ├── server.go
│   │   ├── log.go
│   │   ├── deployment.go
│   │   ├── dashboard.go
│   │   └── monitor.go
│   ├── monitor/
│   │   ├── collector.go     # gopsutil 系统指标采集
│   │   ├── collector_test.go
│   │   ├── history.go       # 环形缓冲历史缓存
│   │   └── history_test.go
│   └── repository/
│       └── db.go            # GORM 连接 + AutoMigrate
├── pkg/
│   └── seed/
│       └── seed.go          # 启动时自动填充种子数据
├── storage/
│   └── devops.db            # SQLite 数据库文件（自动生成）
├── .env                     # 环境变量模板
├── go.mod
└── go.sum
```

## 架构说明

**分层结构**：

```
Handler（参数解析 + JSON 返回）
    ↓ 调用
Repository（GORM 数据库操作）
    ↓ 自动
SQLite 持久化

Monitor 采集器（goroutine + ticker 后台运行）
    ↓ 写入
环形缓冲（内存，线程安全）
    ↓ 查询
Dashboard handler 读取最新数据返回前端
```

- **Handler 层**：接收 HTTP 请求，解析参数，调用下层，返回 JSON
- **Repository 层**：GORM 数据库 CRUD
- **Monitor 采集器**：独立 goroutine 每 10s 采集系统指标，存入环形缓冲
- **App 生命周期**：依赖注入 + 信号监听（SIGINT/SIGTERM）实现优雅关闭

## 调试

```bash
# dlv 调试
dlv debug ./cmd/api

# 启动时重置数据（删除数据库后重启自动重新 seed）
rm storage/devops.db  # Windows: del storage\devops.db
go run ./cmd/api

# 启用 debug 日志
LOG_LEVEL=debug go run ./cmd/api
```

## 构建

```bash
# 当前平台
cd backend
go build -o devops-api ./cmd/api
# Windows 下生成 devops-api.exe，双击即可运行

# 交叉编译
GOOS=linux GOARCH=amd64 go build -o devops-api-linux ./cmd/api
GOOS=darwin GOARCH=arm64 go build -o devops-api-mac ./cmd/api
```

## 技术栈

| 组件 | 选型 |
|------|------|
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | SQLite（pure Go：modernc.org/sqlite，无需 CGO） |
| 系统监控 | gopsutil |
| 日志 | log/slog（Go 标准库） |
| 测试 | Go 标准 testing + go test -race |
