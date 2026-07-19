# Phase 2 后端构建 — Go + Gin 软件设计文档 (SDD)

> **方向**: 后端服务化 — 从 MSW Mock 演进为真实 Go 后端  
> **技术栈**: Go 1.22+ / Gin / GORM / SQLite  
> **目标**: 构建可独立运行的 REST API 服务，前端直连真实后端，数据持久化到 SQLite  
> **日期**: 2026-07-19

---

## 1. 引言

### 1.1 为什么选 Go？

| 优势 | 在本项目中的体现 |
|---|---|
| **单二进制部署** | 编译后单个 `.exe` 文件，双击即运行，无需配环境 |
| **并发原生支持** | 日志查询、Dashboard 指标采集天然适合 goroutine |
| **性能极高** | 资源占用低，笔记本跑起来毫无压力 |
| **静态类型** | 与 TypeScript 前端配合，API 契约清晰 |
| **DevOps 生态** | 云原生标准语言，后续接入 Docker/K8s 极其自然 |

### 1.2 当前状态

- **前端**: 4 个页面已成型，调用 MSW 假数据
- **API 契约**: `spec/v1-api.yaml` 已定义 8 个 endpoint（OpenAPI 3.0）
- **后端**: 无。所有数据在 `src/mocks/browser.ts` 中内存生成

### 1.3 Phase 2 目标

1. 在 `backend/` 目录下新建 Go 项目
2. 100% 复用现有 `v1-api.yaml` 的接口契约（前端不改 API 调用逻辑）
3. SQLite 持久化，重启数据不丢
4. 前端通过 Vite proxy 直连后端，MSW 退居二线（仅前端单独开发时启用）
5. 启动时自动 seed 假数据，开箱即用

---

## 2. 技术选型

| 层级 | 选型 | 理由 |
|---|---|---|
| **Web 框架** | **Gin** | 社区最活跃、性能 top、中间件生态丰富、文档中文资料多 |
| **ORM** | **GORM** | 功能完整、自动迁移（AutoMigrate）、对初学者友好 |
| **数据库** | **SQLite** | 零配置、单文件存储、足够支撑本项目数据量、部署极简 |
| **配置管理** | 环境变量 + `.env` | 不引入复杂配置框架，12-factor 原则 |
| **日志** | `log/slog` (Go 1.21+) | 标准库，结构化日志，无需第三方依赖 |
| **数据生成** | 自写 seed 函数 | 不引入 faker 库，减少依赖，自己控制数据逻辑 |

> **为什么不选其他框架？**> - Echo/Fiber：也很优秀，但 Gin 的资料最多，遇到问题搜索最方便> - sqlx + 手写 SQL：更轻量，但 GORM 的 AutoMigrate 和关联查询能大幅减少 boilerplate> - PostgreSQL/MySQL：需要独立安装服务，SQLite 单文件更适合个人开发和演示

---

## 3. 项目结构设计

采用 Go 社区推荐的 **Standard Go Project Layout** 变体，按职责分层：

```
devops-dashboard/
├── backend/                    # Go 后端根目录
│   ├── cmd/
│   │   └── api/
│   │       └── main.go         # 唯一入口：初始化 DB、seed、启动 HTTP
│   ├── internal/               # 私有代码（禁止外部项目导入）
│   │   ├── api/                # HTTP 层：Gin handler
│   │   │   ├── dashboard.go
│   │   │   ├── server.go
│   │   │   ├── log.go
│   │   │   ├── deployment.go
│   │   │   └── router.go       # 路由注册 + 中间件挂载
│   │   ├── service/            # 业务逻辑层
│   │   │   ├── dashboard.go    # 指标计算、告警生成
│   │   │   ├── server.go
│   │   │   ├── log.go
│   │   │   └── deployment.go
│   │   ├── repository/         # 数据访问层（GORM）
│   │   │   ├── db.go           # 数据库连接 + AutoMigrate
│   │   │   ├── server.go
│   │   │   ├── log.go
│   │   │   └── deployment.go
│   │   └── model/              # 领域模型（GORM model + 请求/响应 DTO）
│   │       ├── server.go
│   │       ├── log.go
│   │       ├── deployment.go
│   │       └── dashboard.go
│   ├── pkg/
│   │   └── seed/               # 数据初始化
│   │       └── seed.go         # 启动时生成假数据
│   ├── storage/
│   │   └── devops.db           # SQLite 数据库文件（gitignore）
│   ├── .env                    # 环境变量模板
│   ├── go.mod
│   └── go.sum
├── frontend/                   # 现有前端（或保持当前 src/ 结构）
│   ├── src/
│   ├── vite.config.ts
│   └── ...
└── spec/
    └── v1-api.yaml             # API 契约（前后端共同遵守）
```

### 3.1 分层职责

| 层 | 职责 | 比喻 |
|---|---|---|
| **Handler (api)** | 解析 HTTP 请求参数 → 调用 Service → 返回 JSON | 餐厅服务员 |
| **Service** | 业务规则、数据组装、跨领域协调 | 厨师长 |
| **Repository** | 数据库 CRUD，纯数据操作 | 仓库管理员 |
| **Model** | 数据结构定义（DB 表映射 + JSON 序列化） | 菜单/食材清单 |

---

## 4. 数据模型设计

### 4.1 与 OpenAPI Schema 的映射

`v1-api.yaml` 中定义的所有 schema，在 Go 中对应 GORM model：

```go
// internal/model/server.go
package model

import "time"

type Server struct {
    ID                string    `gorm:"primaryKey" json:"id"`
    Hostname          string    `json:"hostname"`
    IP                string    `json:"ip"`
    OS                string    `json:"os"`
    CpuCores          int       `json:"cpuCores"`
    MemoryGb          int       `json:"memoryGb"`
    Status            string    `json:"status"` // running, stopped, maintenance
    Uptime            string    `json:"uptime"`
    DiskPartitions    []DiskPartition    `gorm:"foreignKey:ServerID" json:"diskPartitions,omitempty"`
    NetworkInterfaces []NetworkInterface `gorm:"foreignKey:ServerID" json:"networkInterfaces,omitempty"`
    CreatedAt         time.Time `json:"-"`
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
```

```go
// internal/model/log.go
package model

type Log struct {
    ID         string `gorm:"primaryKey" json:"id"`
    Time       string `json:"time"`
    Level      string `json:"level"` // INFO, WARN, ERROR
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
```

```go
// internal/model/deployment.go
package model

import "time"

type Deployment struct {
    ID             string    `gorm:"primaryKey" json:"id"`
    AppName        string    `json:"appName"`
    Version        string    `json:"version"`
    Env            string    `json:"env"` // dev, test, prod
    Status         string    `json:"status"` // pending, deploying, success, failed
    LastDeployedAt time.Time `json:"lastDeployedAt"`
    Histories      []DeploymentHistory `gorm:"foreignKey:DeploymentID" json:"-"`
}

type DeploymentHistory struct {
    ID           uint      `gorm:"primaryKey" json:"-"`
    DeploymentID string    `json:"-"`
    Version      string    `json:"version"`
    Operator     string    `json:"operator"`
    DurationSec  int       `json:"durationSec"`
    Status       string    `json:"status"` // success, failed
    DeployedAt   time.Time `json:"deployedAt"`
}
```

```go
// internal/model/dashboard.go
package model

type MetricValue struct {
    Current float64 `json:"current"`
    Status  string  `json:"status"` // normal, warning, critical
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
    Level   string `json:"level"` // info, warning, critical
    Message string `json:"message"`
    Source  string `json:"source"`
    Time    string `json:"time"`
}
```

### 4.2 数据库初始化策略

1. **AutoMigrate**：启动时 GORM 自动建表（`db.AutoMigrate(...)`）
2. **Seed 检测**：查询 `servers` 表，若为空则触发 seed
3. **数据生成**：完全复现当前 MSW 中的数据逻辑，用 Go 的 `math/rand` 和 `time` 包模拟

---

## 5. API 层详细设计（Gin Handler）

所有路由前缀为 `/api`，与前端现有调用路径一致。

### 5.1 Dashboard 模块

```go
// GET /api/dashboard/metrics
func GetDashboardMetrics(c *gin.Context) {
    metrics := dashboardService.GetMetrics()
    c.JSON(200, metrics)
}

// GET /api/dashboard/trend?hours=6
func GetDashboardTrend(c *gin.Context) {
    hours := c.DefaultQuery("hours", "6")
    h, _ := strconv.Atoi(hours)
    trend := dashboardService.GetTrend(h)
    c.JSON(200, trend)
}

// GET /api/dashboard/alerts?limit=5
func GetDashboardAlerts(c *gin.Context) {
    limit := c.DefaultQuery("limit", "5")
    l, _ := strconv.Atoi(limit)
    alerts := dashboardService.GetAlerts(l)
    c.JSON(200, alerts)
}
```

### 5.2 Server 模块

```go
// GET /api/servers?page=1&pageSize=10&status=running
func GetServerList(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
    status := c.Query("status")

    result := serverService.List(page, pageSize, status)
    c.JSON(200, result)
}

// GET /api/servers/:id
func GetServerDetail(c *gin.Context) {
    id := c.Param("id")
    server, err := serverService.GetByID(id)
    if err != nil {
        c.JSON(404, gin.H{"error": "server not found"})
        return
    }
    c.JSON(200, server)
}
```

### 5.3 Log 模块

```go
// GET /api/logs?page=1&pageSize=10&level=INFO&keyword=xxx
func GetLogList(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
    level := c.Query("level")
    keyword := c.Query("keyword")

    result := logService.List(page, pageSize, level, keyword)
    c.JSON(200, result)
}
```

### 5.4 Deployment 模块

```go
// GET /api/deployments
func GetDeploymentList(c *gin.Context) {
    list := deploymentService.List()
    c.JSON(200, list)
}

// GET /api/deployments/:id/history
func GetDeploymentHistory(c *gin.Context) {
    id := c.Param("id")
    history := deploymentService.GetHistory(id)
    c.JSON(200, history)
}
```

### 5.5 路由注册（router.go）

```go
func SetupRouter() *gin.Engine {
    r := gin.Default()

    // CORS：允许前端 localhost:5173 访问
    r.Use(corsMiddleware())

    api := r.Group("/api")
    {
        api.GET("/dashboard/metrics", GetDashboardMetrics)
        api.GET("/dashboard/trend", GetDashboardTrend)
        api.GET("/dashboard/alerts", GetDashboardAlerts)

        api.GET("/servers", GetServerList)
        api.GET("/servers/:id", GetServerDetail)

        api.GET("/logs", GetLogList)

        api.GET("/deployments", GetDeploymentList)
        api.GET("/deployments/:id/history", GetDeploymentHistory)
    }

    return r
}
```

---

## 6. 前端适配方案

### 6.1 开发环境代理

前端 `vite.config.ts` 增加 `server.proxy`，将 `/api` 转发到 Go 后端：

```typescript
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

### 6.2 MSW 切换策略

MSW 不再默认启动。改为通过环境变量控制：

```typescript
// src/main.ts 改造后
async function bootstrap() {
  if (import.meta.env.DEV && import.meta.env.VITE_USE_MSW === 'true') {
    const { worker } = await import('./mocks/browser')
    await worker.start()
  }
  // ...render
}
```

- **联调模式**：Go 后端运行，前端直连（默认）
- **纯前端模式**：`.env.local` 设置 `VITE_USE_MSW=true`，前端独立开发

### 6.3 Axios BaseURL

`src/api/axios-instance.ts` 中的 `baseURL` 改为相对路径 `/` 或 `/api`，由 Vite proxy 转发：

```typescript
const customInstance = axios.create({
  baseURL: '/api', // 开发时 proxy 到 localhost:8080，生产时由 nginx 转发
  timeout: 10000,
})
```

---

## 7. 实施路线图

### Sprint 1：Go 项目骨架 + 数据库（1 周）

| ID | 任务 | 产出 |
|---|---|---|
| S1-1 | 初始化 Go 模块，安装 Gin + GORM + SQLite driver | `backend/go.mod`，可编译运行 |
| S1-2 | 搭建分层目录结构，编写 `main.go` 入口 | `cmd/api/main.go` |
| S1-3 | 配置 GORM + SQLite，实现 AutoMigrate | `internal/repository/db.go`，自动建表 |
| S1-4 | 编写 Model 层（所有 schema 的 Go struct） | `internal/model/*.go` |
| S1-5 | 实现 Seed 逻辑（生成假数据写入 DB） | `pkg/seed/seed.go`，启动自动填充 |
| S1-6 | 验证：运行后端，用 curl/浏览器访问 `GET /api/servers` | 返回 JSON 数组 |

**Sprint 1 门禁**：
- `go run cmd/api/main.go` 成功启动
- SQLite 文件生成，表结构正确
- curl `localhost:8080/api/servers` 有数据返回

---

### Sprint 2：API 实现 + 前端联调（1.5 周）

| ID | 任务 | 产出 |
|---|---|---|
| S2-1 | 实现 Server 模块（List + Detail + 分页筛选） | 完整 REST API |
| S2-2 | 实现 Log 模块（List + 级别筛选 + 关键词搜索） | 完整 REST API |
| S2-3 | 实现 Deployment 模块（List + History） | 完整 REST API |
| S2-4 | 实现 Dashboard 模块（Metrics + Trend + Alerts） | 内存计算或轻量查询 |
| S2-5 | 前端配置 Vite proxy，联调 4 个页面 | 前端显示真实后端数据 |
| S2-6 | 移除 MSW 默认启动，改为环境变量控制 | `main.tsx` 改造 |

**Sprint 2 门禁**：
- 前端 4 个页面全部从 Go 后端取数据，功能与 Phase 1 一致
- Dashboard 趋势图正常渲染
- 分页、筛选、搜索交互正常

---

### Sprint 3：工程化完善（1 周）

| ID | 任务 | 产出 |
|---|---|---|
| S3-1 | 统一错误处理：Gin 全局错误中间件 | 404/500 返回统一 JSON 结构 |
| S3-2 | 结构化日志：接入 `log/slog` | 请求日志、慢查询日志 |
| S3-3 | 配置外部化：`.env` 支持端口、DB 路径配置 | `backend/.env` |
| S3-4 | 优雅关闭：捕获 SIGINT/SIGTERM，关闭 DB 连接 | `main.go` signal handling |
| S3-5 | 后端 README：如何运行、如何重新 seed、目录说明 | `backend/README.md` |
| S3-6 | 构建验证：`go build` 产出可执行文件 | `backend/api.exe`（Windows） |

**Sprint 3 门禁**：
- 编译后单文件可独立运行（不依赖 Go 环境）
- 配置文件可修改端口和 DB 路径
- Ctrl+C 优雅退出，无数据损坏

---

## 8. 关键设计决策

### 8.1 Dashboard Metrics 怎么存？

**决策**：Metrics 和 Trend **不存数据库**，由 Service 层内存计算 + 定时 goroutine 更新。

**理由**：
- Dashboard 数据是"瞬时值"和"最近 N 小时"，不需要长期历史
- 用 goroutine 每 30 秒生成一次新指标，模拟真实监控采集
- Alerts 也放在内存中，定期轮转（保留最近 20 条）

**实现思路**：

```go
type DashboardCollector struct {
    metrics   model.DashboardMetrics
    trend     model.DashboardTrend
    alerts    []model.AlertItem
    mu        sync.RWMutex
}

func (c *DashboardCollector) Start() {
    ticker := time.NewTicker(30 * time.Second)
    go func() {
        for range ticker.C {
            c.collect()
        }
    }()
}
```

### 8.2 日志表数据量大，SQLite 撑得住吗？

**决策**：Seed 时只生成 200 条日志，足够前端分页演示。

**理由**：
- 本项目是 Dashboard 演示，非真实日志系统
- 200 条在 SQLite 中毫秒级查询
- 若未来扩展，可无缝迁移到 PostgreSQL（GORM 支持多种 dialect）

### 8.3 是否需要认证（JWT）？

**决策**：Phase 2 **不做认证**。

**理由**：
- 当前前端无登录页面，引入 JWT 需要改动前端路由和状态管理
- 如需保护，Sprint 3 可加一个简单的 API Key 中间件（环境变量配置）作为过渡

---

## 9. 开发环境运行指南（目标状态）

### 9.1 启动后端

```bash
cd backend
cp .env.example .env
go mod tidy
go run cmd/api/main.go

# 输出：
# 2026/07/19 10:00:00 INFO 数据库已连接: storage/devops.db
# 2026/07/19 10:00:00 INFO 数据已初始化: 35 servers, 200 logs, 15 deployments
# 2026/07/19 10:00:00 INFO 服务启动: http://localhost:8080
```

### 9.2 启动前端（联调模式）

```bash
cd frontend
npm run dev
# 自动通过 proxy 访问 localhost:8080/api
```

### 9.3 构建生产包

```bash
# 后端：编译为单文件
cd backend
GOOS=linux GOARCH=amd64 go build -o devops-api cmd/api/main.go

# 前端：静态资源
cd frontend
npm run build

# 部署：将 devops-api + frontend/dist/ 上传到服务器
# ./devops-api  # 启动后端，同时提供静态文件（Gin static middleware）
```

---

## 10. 验收标准

| 检查项 | 标准 |
|---|---|
| API 兼容性 | 所有 8 个 endpoint 的响应结构与 `v1-api.yaml` 100% 一致，前端无需修改 API 调用代码 |
| 数据持久化 | 重启后端，服务器/日志/部署数据不丢失 |
| 单文件运行 | `go build` 后单个可执行文件可直接运行，无需安装 Go 环境 |
| 前端联调 | 4 个页面全部正常展示后端数据，无 MSW 也能工作 |
| 并发安全 | Dashboard 指标采集 goroutine 与 HTTP handler 无数据竞争 |
| 优雅关闭 | Ctrl+C 后数据库连接正常关闭，无日志报错 |

---

## 11. 风险与缓解

| 风险 | 影响 | 缓解 |
|---|---|---|
| GORM + SQLite 在 Windows 上的 CGO 编译问题 | 阻塞开发 | 使用纯 Go 的 SQLite driver（`modernc.org/sqlite`），无需 CGO |
| 前端联调时 CORS 报错 | 体验差 | Gin 中间件统一配置 CORS，允许 localhost:5173 |
| Go 学习曲线导致进度延迟 | 延期 | Sprint 1 只要求"能跑通一个 API"，后续增量添加；每步提供代码模板 |
| Seed 数据逻辑复杂，与前端期望不符 | 联调失败 | 完全复刻 MSW 中的数据结构、字段名、值范围 |

---

## 12. 附录

### 12.1 前端与后端目录关系

```
devops-dashboard/           # Git 根目录
├── backend/                # Go 后端（新增）
├── frontend/               # React 前端（由当前 src/ 迁移或保留原样）
│   └── src/
│       ├── api/            # Orval 生成的 client（复用）
│       ├── mocks/          # MSW（保留，可选启用）
│       └── features/       # 页面（基本不改动）
├── spec/
│   └── v1-api.yaml         # 契约（前后端共用）
└── docs/
    └── phase2-go-backend-sdd.md
```

> **关于 frontend 目录**：为保持清晰，建议将现有前端代码整体移入 `frontend/` 目录，与 `backend/` 并列。如果担心改动大，也可以保持当前根目录结构，仅在根目录新增 `backend/` 文件夹。两种方式均可，实施时根据你的偏好决定。

### 12.2 推荐学习路径（Go 初学者）

如果你 Go 基础还不扎实，按这个顺序边做边学：

1. **先跑通官方 Tour**：https://tour.golang.org/ （半天，重点看 goroutine、channel、struct、interface）
2. **Gin 快速入门**：https://gin-gonic.com/docs/quickstart/ （1 小时，跑通 Hello World + 路由参数）
3. **GORM 快速开始**：https://gorm.io/docs/index.html （重点看 Model 定义、AutoMigrate、CRUD、Preload）
4. **跟着 Sprint 1 做**：每一步我提供完整代码，你负责编译运行和微调

---

*文档版本: v1.0*  
*维护人: DevOps Dashboard Team*  
*下次评审: Sprint 2 联调完成后*
