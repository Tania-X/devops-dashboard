# DevOps Dashboard — 架构文档

> 记录项目整体架构、分层设计、关键决策与当前实现状态。
> 最后更新: 2026-07-25

---

## 一、项目概览

运维监控仪表盘，Go + React 全栈学习项目。前端展示系统运行状态，后端通过 gopsutil 采集本机指标并提供 REST API。

### 技术栈

| 层 | 技术 |
|--|--|
| 前端框架 | React 19 + TypeScript |
| 构建工具 | Vite 8 |
| UI 组件库 | Ant Design 6 |
| 图表 | ECharts 6 |
| API 规范 | OpenAPI 3.0（SDD — Spec-Driven Development） |
| Mock | MSW (Mock Service Worker) |
| 后端框架 | Go 1.24+ / Gin |
| ORM | GORM |
| 数据库 | SQLite（modernc.org/sqlite，纯 Go 无 CGO） |
| 系统监控 | gopsutil (CPU/内存/磁盘/进程) |
| 日志 | log/slog（Go 标准库结构化日志） |
| 测试 | Go 标准 testing + -race |

---

## 二、项目结构

```
devops-dashboard/
├── spec/                          # OpenAPI 规范（SDD 契约核心）
│   ├── v1-api.yaml                # 所有 API 接口定义
│   └── ui-theme.md                # UI 视觉规范（色值/布局/组件）
├── frontend/
│   └── src/
│       ├── api/                   # Orval 生成（或手写）的 API 客户端
│       ├── components/layout/     # 全局布局（Header/Sidebar/Content）
│       ├── features/
│       │   ├── dashboard/         # 系统概览页
│       │   ├── server/            # 服务器管理页
│       │   ├── log/               # 日志查询页
│       │   ├── deployment/        # 部署状态页
│       │   └── monitor/           # 实时监控页（进程 + 主机信息，待建设）
│       └── mocks/                 # MSW Mock Handlers
├── backend/
│   ├── cmd/api/main.go            # 入口：读配置 → 初始化 → 启动 HTTP
│   ├── internal/
│   │   ├── api/                   # HTTP Handler 层
│   │   │   ├── router.go          # 路由注册 + CORS/日志/恢复中间件
│   │   │   ├── errors.go          # 统一错误响应 + panic 恢复
│   │   │   ├── server.go          # 服务器管理 handler
│   │   │   ├── dashboard.go       # Dashboard handler
│   │   │   ├── log.go             # 日志查询 handler
│   │   │   ├── deployment.go      # 部署管理 handler
│   │   │   ├── monitor.go         # 实时监控 handler
│   │   │   └── health.go          # 健康检查 handler
│   │   ├── service/               # 业务逻辑层
│   │   │   ├── service.go         # Services 聚合 + NewServices 构造函数
│   │   │   ├── server.go          # ServerService
│   │   │   ├── deployment.go      # DeploymentService
│   │   │   ├── log.go             # LogService
│   │   │   ├── dashboard.go       # DashboardService
│   │   │   └── monitor.go         # MonitorService
│   │   ├── monitor/               # 系统指标采集
│   │   │   ├── collector.go       # gopsutil 采集 CPU/内存/磁盘
│   │   │   ├── collector_test.go
│   │   │   ├── history.go         # 环形缓冲历史缓存
│   │   │   └── history_test.go
│   │   ├── model/                 # 数据模型
│   │   │   ├── server.go
│   │   │   ├── log.go
│   │   │   ├── deployment.go
│   │   │   ├── dashboard.go
│   │   │   └── monitor.go         # ProcessItem / ProcessDetail / HostInfo
│   │   ├── repository/
│   │   │   └── db.go              # GORM 连接 + AutoMigrate
│   │   ├── config/
│   │   │   └── config.go          # 环境变量加载
│   │   └── app/
│   │       └── app.go             # 应用生命周期（Init/Run/Shutdown）
│   ├── pkg/seed/
│   │   └── seed.go                # 启动时自动填充种子数据
│   └── storage/
│       └── devops.db              # SQLite 数据库（自动生成）
├── docs/
│   ├── architecture.md            # ← 本文档
│   ├── development-guide.md       # 开发指南 + 经验教训 + 路线图
│   └── env-setup-macos.md         # macOS 环境搭建
└── AGENTS.md                      # Claude Code 工作约束
```

---

## 三、后端架构

### 3.1 分层结构

```
┌──────────────────────────────────────────────────────────┐
│                     HTTP 层 (Handler)                     │
│  解析参数 → 调用 Service → 返回 JSON                      │
│  每个 Handler 对应一个 API Endpoint                       │
├──────────────────────────────────────────────────────────┤
│                   业务逻辑层 (Service)                     │
│  数据组装、跨领域协调、业务规则                            │
│  通过 Services 聚合结构体集中注入 Handler                  │
├──────────────────────────────────────────────────────────┤
│        ┌─────────────────┬──────────────────────┐       │
│        │   Repository    │      Monitor          │       │
│        │  (GORM CRUD)    │  (gopsutil 采集器)    │       │
│        │     ↓           │    ↓         ↓        │       │
│        │   SQLite        │  即时指标   环形缓冲   │       │
│        └─────────────────┴──────────────────────┘       │
└──────────────────────────────────────────────────────────┘
```

| 层 | 职责 | 类比 |
|--|--|--|
| **Handler (api/)** | 解析 HTTP 参数、调用 Service、返回 JSON 响应 | 餐厅服务员 |
| **Service (service/)** | 业务规则、数据组装、跨模块协调 | 厨师长 |
| **Repository (repository/)** | GORM 数据库 CRUD（当前 db.go 仅含连接/迁移） | 仓库管理员 |
| **Monitor (monitor/)** | gopsutil 系统指标采集 + 环形缓冲历史缓存 | 传感器 |
| **Model (model/)** | 数据结构定义（DB 映射 + JSON 序列化） | 食材清单 |

### 3.2 依赖注入模式

```
app.go New(cfg)
  → Init()
    → repository.InitDB(cfg)          // GORM 连接
    → monitor.NewHistory(retain, interval)  // 采集器
    → service.NewServices(db, history) // 聚合所有 Service
    → api.NewHandler(db, history, services)  // 注入 Handler
  
Handler 通过 h.services.XXXService 调用 Service
Service 通过自身持有的 db/monitor 访问数据
```

核心代码 (`internal/service/service.go`)：

```go
type Services struct {
    ServerService     *ServerService
    DeploymentService *DeploymentService
    LogService        *LogService
    DashboardService  *DashboardService
    MonitorService    *MonitorService
}

func NewServices(db *gorm.DB, history *monitor.History) *Services {
    return &Services{
        ServerService:     NewServerService(db),
        DeploymentService: NewDeploymentService(db),
        LogService:        NewLogService(db),
        DashboardService:  NewDashboardService(db, history),
        MonitorService:    NewMonitorService(db),
    }
}
```

### 3.3 应用生命周期

`internal/app/app.go` 管理整个应用的生命周期：

1. **Init()** — 初始化 logger、DB、种子数据、采集器、Service、HTTP 路由
2. **Run()** — 启动 HTTP 服务，监听 SIGINT/SIGTERM
3. **shutdown()** — 关闭采集器 goroutine，优雅关闭 HTTP（5s 超时）

### 3.4 系统指标采集

`internal/monitor/` 包包含两个核心组件：

- **Collector**：通过 gopsutil 实时采集本机 CPU、内存、磁盘使用率，数据不落库
- **History**：固定大小环形缓冲（ring buffer），每 10s 由独立 goroutine 写入一次，支持按小时查询趋势数据

```
History 构造参数：
  retain = 24h（保留时长）
  interval = 10s（采集间隔）
  maxSize = retain/interval = 8640 条
```

### 3.5 配置项

| 环境变量 | 默认值 | 说明 |
|--|--|--|
| `PORT` | `8080` | 监听端口 |
| `DB_PATH` | `storage/devops.db` | SQLite 数据库路径 |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `LOG_FORMAT` | `text` | `text`（开发）或 `json`（生产，可接入 Loki） |
| `ENV` | `dev` | 运行环境标识 |
| `VERSION` | `dev` | 版本号（日志全局字段） |
| `HISTORY_RETAIN` | `24h` | 趋势数据保留时长 |
| `HISTORY_INTERVAL` | `10s` | 采集间隔 |

---

## 四、API 实现状态

全部接口基于 `spec/v1-api.yaml` 实现。当前共 11 个 Endpoint + 1 个健康检查：

| 方法 | 路径 | 实现 | 数据源 | Service |
|--|--|--|--|--|
| GET | `/api/health` | ✅ | 内存 | 无 |
| GET | `/api/dashboard/metrics` | ✅ 实时采集 | gopsutil | DashboardService |
| GET | `/api/dashboard/trend` | ✅ 最近 N 小时 | 环形缓冲 | DashboardService |
| GET | `/api/dashboard/alerts` | ⚠️ Mock 数据，无真实采集 | 硬编码 | 无（Handler 内） |
| GET | `/api/servers` | ✅ 分页 + 状态筛选 | SQLite | ServerService |
| GET | `/api/servers/:id` | ✅ 含磁盘/网卡详情 | SQLite | ServerService |
| GET | `/api/logs` | ✅ 分页 + 级别/关键词/服务名筛选 | SQLite | LogService |
| GET | `/api/deployments` | ✅ 按部署时间排序 | SQLite | DeploymentService |
| GET | `/api/deployments/:id/history` | ✅ | SQLite | DeploymentService |
| GET | `/api/monitor/processes` | ✅ 排序/搜索/条数限制 | gopsutil | MonitorService |
| GET | `/api/monitor/processes/:pid` | ✅ 进程详情 | gopsutil | MonitorService |
| GET | `/api/monitor/host` | ✅ 主机信息 | gopsutil | MonitorService |

> ⚠️ **Alerts**：当前为 Handler 内硬编码 mock 数据，待告警采集系统接入后替换。

---

## 五、关键设计决策

### 5.1 为什么不用 Repository 层？

`docs/phase2-go-backend-sdd.md`（已归档）中规划了独立的 Repository 层，但实际实现中：
- 模块仅 5 个简单 model，CRUD 操作单一（基本都是 `WHERE + ORDER + LIMIT`）
- GORM 的链式调用已经足够表达查询，多加一层 Repository 反而增加 boilerplate
- **最终分层**：Handler → Service → db/monitor（无独立 Repository 层）

若后续业务复杂化（多表关联、事务、读写分离），可随时从 Service 中拆出 Repository。

### 5.2 为什么用环形缓冲存趋势数据？

采集到的指标需要保存最近 N 小时以供趋势图查询。选择内存环形缓冲而非 SQLite：
- 时序数据写入频繁（每 10s 一次），SQLite 写压力大
- 数据有时效性（24h 后丢弃），不适合长期存储
- 环形缓冲 O(1) 写入，固定内存开销

### 5.3 为什么不是 DDD？

当前架构是典型的 **Transaction Script / 分层架构**，按"数据驱动"组织代码：
- Service 方法对应一个 API 操作（List、GetByID 等）
- 业务逻辑简单，不存在复杂的领域规则和聚合根
- 项目规模（5 个 model、11 个接口）引入 DDD 属于过度设计

### 5.4 为什么 PID 用 int32？

gopsutil 的 `process.NewProcess` 接收 `int32` 参数，因为操作系统 PID 类型（Linux `pid_t`）为 32 位有符号整数，最大 PID 约 419 万，`int32` 完全够用。

---

## 六、种子数据

启动时自动检测数据库是否为空，若为空则自动填充：

| 表 | 数据量 | 说明 |
|--|--|--|
| servers | 35 条 | 含状态/磁盘/网卡 |
| logs | 200 条 | 混合 INFO/WARN/ERROR 级别 |
| deployments | 15 条 | 含关联的 deployment_histories |

`rm storage/devops.db` 后重启即可重新 seed。

---

## 七、相关文档

- [开发指南](development-guide.md) — 本地启动、前端模式、常见问题、路线图
- [环境搭建 (macOS)](env-setup-macos.md) — 从零配置开发环境
- [API 契约](../../spec/v1-api.yaml) — OpenAPI 3.0 接口定义
- [UI 视觉规范](../../spec/ui-theme.md) — 色值、布局、组件设计
