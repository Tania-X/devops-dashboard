# DevOps Dashboard — 架构文档

> 记录项目整体架构、分层设计、关键决策、演进设计与实施计划。
> 最后更新: 2026-08-08

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
│       │   ├── monitor/           # 实时监控页（进程 + 主机信息，待建设）
│       │   ├── agent/             # Agent 管理页
│       │   └── user/              # 用户管理页
│       └── mocks/                 # MSW Mock Handlers
├── backend/
│   ├── cmd/api/main.go            # 入口：读配置 → 初始化 → 启动 HTTP
│   ├── cmd/agent/                 # Agent 独立二进制入口（远程采集代理）
│   ├── internal/
│   │   ├── api/                   # HTTP Handler 层
│   │   ├── service/               # 业务逻辑层
│   │   ├── monitor/               # 系统指标采集（gopsutil + 环形缓冲）
│   │   ├── model/                 # 数据模型
│   │   ├── repository/            # GORM 连接 + 迁移 + seed
│   │   ├── notify/                # 告警总线 + Webhook 通知器
│   │   ├── config/                # 环境变量加载
│   │   └── app/                   # 应用生命周期（Init/Run/Shutdown）
│   ├── pkg/seed/                  # 启动时自动填充种子数据
│   └── storage/                   # SQLite 数据库（自动生成）
├── docs/                          # 开发文档与设计文档
└── AGENTS.md                      # Agent 工作约束
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

全部接口基于 `spec/v1-api.yaml` 实现。当前共 12 个 Endpoint + 认证/用户/Agent/设置管理：

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
| POST | `/api/auth/login` | ✅ | SQLite | AuthService |
| POST | `/api/auth/logout` | ✅ | - | AuthService |
| GET | `/api/auth/me` | ✅ | SQLite | AuthService |
| GET/POST/PUT/DELETE | `/api/agents...` | ✅ 管理 + 部署/停止/状态 | SQLite | AgentService |
| GET/POST/PUT/DELETE | `/api/users...` | ✅ 管理 | SQLite | UserService |
| GET/PUT/POST | `/api/settings/webhook...` | ✅ 配置 + 测试 | SQLite | WebhookManager |

> ⚠️ **Alerts**：当前为 Handler 内硬编码 mock 数据，待告警采集系统接入后替换（Phase 4）。

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
- 项目规模（5 个 model、12 个接口）引入 DDD 属于过度设计

> **演进说明（2026-08-08）**：DDD 全量战术模式仍不引入，但已确定"渐进式 DDD"路线——先做低成本阶段（限界上下文拆包 + 充血模型），见「七、演进设计：DDD」。

### 5.4 为什么 PID 用 int32？

gopsutil 的 `process.NewProcess` 接收 `int32` 参数，因为操作系统 PID 类型（Linux `pid_t`）为 32 位有符号整数，最大 PID 约 419 万，`int32` 完全够用。

### 5.5 敏感字段策略：请求 DTO 分离

**实体 model 的敏感字段（Password/Secret）一律 `json:"-"` 永不序列化，入站参数用专用 Request DTO（binding + 校验）。** 禁止"`json:"-"` 绑定 + 返回处手动清空"的做法。

> 教训（2026-08-08）：`json:"-"` 在 encoding/json 中**同时屏蔽序列化与反序列化**，曾导致创建用户时密码永远绑定不上（返回 400"用户名和密码不能为空"），Agent 密码、Webhook 加签密钥同样存不进去。

已应用：User（`CreateUserRequest`/`UpdateUserRequest`）、Agent（`CreateAgentRequest`/`UpdateAgentRequest`）、Webhook（`UpdateWebhookConfigRequest`）。

---

## 六、种子数据

启动时自动检测数据库是否为空，若为空则自动填充：

| 表 | 数据量 | 说明 |
|--|--|--|
| servers | 35 条 | 含状态/磁盘/网卡 |
| logs | 200 条 | 混合 INFO/WARN/ERROR 级别 |
| deployments | 15 条 | 含关联的 deployment_histories |
| users | 1 条 | 默认 admin / admin123（repository/db.go seedAdminUser） |

`rm storage/devops.db` 后重启即可重新 seed。

---

## 七、演进设计：DDD（限界上下文 + 充血模型）

### 7.1 背景

- 现状：Transaction Script 架构 + **贫血模型**（model 纯字段，业务逻辑集中在 Service）
- 触发（2026-08-08）：用户学习 DDD，提出两个方向：
  1. 后端本体与 Agent 边界清晰，可各自演进
  2. 将 User 等模型设计为充血模型，行为内聚

### 7.2 设计决策：渐进式 DDD

| 阶段 | 内容 | 成本 | 时机 |
|------|------|------|------|
| ① 边界 + 充血 | 按限界上下文拆包；User 等实体改充血模型 | 低 | **现在可做** |
| ② 仓储接口 | domain 定义 Repository 接口，infrastructure 用 GORM 实现 | 中 | 业务复杂化后 |
| ③ 完整战术 DDD | 聚合、防腐层、领域事件、CQRS | 高 | 暂不做（过度设计） |

> 当前规模（5 model / 12 API）配全量 DDD 属于过度设计；阶段 ① 消除真实的贫血反模式，且不引入额外抽象成本。

### 7.3 限界上下文划分

- **上下文 A：DevOps 本体**（`backend/internal/api`）— 用户/服务器/部署/日志/监控/告警；部署为 Docker 镜像
- **上下文 B：Agent**（`cmd/agent` 独立二进制）— 宿主机指标采集；直连宿主机、不走 Docker

二者通过 HTTP 通信（AGENT_HOSTS），已有物理边界，需在代码层（`internal/`）按上下文隔离，各自独立演进。

### 7.4 充血模型原则

1. **实体 = 状态 + 状态变更行为**，行为内聚在实体（如 `User.SetPassword` / `VerifyPassword` / `ChangeRole`）
2. **创建/删除归工厂 + 应用服务**，不进实体（创建时无实例状态可言）
3. **敏感字段私有化**（如 `passwordHash` 私有，外部只能通过 `VerifyPassword` 校验）
4. Service 从"上帝类"瘦身为**编排者**（NewUser → repo.Save → 返回）

```go
// 目标形态
type User struct {
    id, username string
    passwordHash string   // 私有，外部不可见
    role         Role
}
func (u *User) SetPassword(plain string) error
func (u *User) VerifyPassword(plain string) bool
func (u *User) ChangeRole(r Role) error
func NewUser(name, plain, role string) (*User, error) // 工厂
```

### 7.5 演进目标结构

```
backend/internal/
├── agent/          # Agent 上下文（独立演进）
└── dashboard/      # 本体上下文
    ├── domain/     # 实体 + 值对象 + 工厂（不依赖 gin/gorm）
    ├── service/    # 应用服务（编排 + 事务）
    ├── api/        # HTTP handler（参数解析 → 应用服务 → 响应）
    └── monitor/    # 采集器（基础设施，保持现状）
```

### 7.6 风险与权衡

- 阶段 ① 涉及文件移动与 model 改造，接口语义保持不变，可渐进式提交
- `AlertBus`（notify 包）已是**领域事件雏形**，阶段 ③ 可平滑升级，无需重写
- 试点：User 先行，Server/Deployment 等按需跟进，不强制全部改造

---

## 八、演进设计：RBAC 权限（Casbin + 按钮级）

### 8.1 现状问题

1. **路由硬编码**：`router.go` 中 `admin := auth.Group("") + AdminMiddleware`，按路由分组生硬分离 admin/非 admin
2. **二值权限**：只有 admin / viewer，无法表达中间态（如"可管理 Agent 但不可管理用户"）
3. **权限与路由耦合**：新增能力要改路由注册与中间件
4. JWT claims 带 `role`，但 `ValidateToken` 实际查 DB（claims 中 role 冗余）

### 8.2 目标

1. **权限点驱动**：每个 API 标注所需权限，如 `user:read` / `user:write` / `agent:manage` / `webhook:manage`
2. **角色 → 权限集合**，支持自定义角色（admin / viewer / operator…）
3. **中间件通用化**：`RequirePermission(obj, act)` 取代 `AdminMiddleware`
4. 权限变更**即时生效**（不需要重新登录）
5. **按钮级权限**：前端按权限点控制按钮显隐（UX 层），后端接口校验为安全边界

### 8.3 方案对比

| 方案 | 说明 | 适用场景 |
|------|------|----------|
| **Casbin（推荐）** | Go 最成熟授权库；PERM 元模型，支持 RBAC/ABAC/ACL；策略可存 DB/文件；gin 集成 `github.com/casbin/gin-authz` | 复杂授权需求、团队项目 |
| 自研轻量 RBAC | `permissions` + `role_permissions` 表，权限点常量，启动加载内存 map，中间件 O(1) 查 | 极简场景、零依赖偏好 |

> 选择 Casbin（2026-08-08 用户评审决定）：生态成熟、社区方案完善，后续要扩展 ABAC/多租户无需重写；本项目的权限模型对 Casbin 属于轻量使用，接入成本可控。自研方案作废。详细设计见 [RBAC/Casbin 详细设计](rbac-casbin-design.md)。

### 8.4 Casbin 接入设计（概要）

**依赖**：`github.com/casbin/casbin/v2` + `github.com/casbin/gorm-adapter/v3`（策略存 SQLite `casbin_rule` 表）+ `github.com/casbin/gin-authz`

**模型（PERM + RBAC）**：`r = sub, obj, act`；匹配 `g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act`

**权限点映射（接口级）**：

| 权限点 | obj | act | admin | viewer | operator |
|--------|-----|-----|-------|--------|----------|
| `dashboard:view` | dashboard | view | ✅ | ✅ | ✅ |
| `server:read` | server | read | ✅ | ✅ | ✅ |
| `agent:manage` | agent | manage | ✅ | ❌ | ✅ |
| `user:read` | user | read | ✅ | ❌ | ❌ |
| `user:write` | user | write | ✅ | ❌ | ❌ |
| `webhook:manage` | webhook | manage | ✅ | ❌ | ❌ |

**策略 seed**：

```
p, admin,    *,        *
p, viewer,   dashboard, view
p, viewer,   server,   read
p, operator, dashboard, view
p, operator, server,   read
p, operator, agent,    manage
```

**中间件**：`RequirePermission(obj, act)` 封装 Enforcer，替换 `AdminMiddleware`；路由声明式标注。

### 8.5 按钮级权限（前端 UX 控制）

> 原则：按钮级控制是**体验层（前端）**，接口级校验是**安全边界（后端）**——按钮隐藏不等于接口安全，接口仍由 Casbin 强制校验（纵深防御）。

1. **权限点细化**（敏感操作独立权限点）：

| 权限点 | 控制的按钮 | admin | viewer | operator |
|--------|-----------|-------|--------|----------|
| `user:read` / `user:create` / `user:update` / `user:delete` | 用户页 查看/新增/编辑/删除 | ✅ | ❌ | ❌ |
| `agent:read` / `agent:create` / `agent:update` / `agent:delete` / `agent:deploy` / `agent:stop` | Agent 页 各操作按钮 | ✅ | ❌ | ✅ |
| `webhook:read` / `webhook:update` / `webhook:test` | 设置页 查看/保存/测试 | ✅ | ❌ | ❌ |
| `dashboard:view` / `server:read` / `log:read` / `deployment:read` | 只读页面（无按钮） | ✅ | ✅ | ✅ |

2. **前端获取权限**：`GET /api/auth/me` 返回 `permissions: string[]`（登录时由 Casbin 按角色计算）——单一事实源在后端
3. **前端实现**：`usePermission('user:delete')` hook + `<AuthButton perm="...">` 组件控制显隐
4. **后端路由仍 `RequirePermission(obj, act)` 校验**

### 8.6 与 DDD 的结合

- `Role` 作为值对象/实体，`User.ChangeRole` 校验角色合法性
- 权限模型在 domain 层建模，infrastructure 持久化
- 依赖方向：HTTP → 应用服务 → 领域（authz 判断可下沉为领域服务）

### 8.7 待定问题（2026-08-08 用户评审）

- [x] 是否引入 Casbin → **引入**（用户修正选择）
- [x] 角色是否可配置 → 第一期不做，策略用代码 seed 预置
- [x] viewer 是否拆分 operator → 保留 viewer + 预置 operator
- [ ] 按钮级权限点列表由后端 `/auth/me` 计算返回 → 建议采用（单一事实源），待确认

---

## 九、实施计划（RBAC + DDD 合并路线）

### 9.1 执行顺序与依赖

RBAC 是横切改动（middleware/router），DDD 拆包同样会动 router 的 import 路径 → **先 RBAC 后 DDD**，拆包时只改路径不改逻辑，返工最小；User 充血模型在 RBAC 之后做，顺手把 Role 抽象为值对象，两个方向自然衔接。

| 步骤 | 内容 | 主要涉及 | 产出 |
|------|------|----------|------|
| **Step 1 Casbin 集成** | 引入 `casbin/v2` + `gorm-adapter/v3`；model.conf 定义；enforcer 初始化；策略 seed（admin/viewer/operator） | `internal/authz`、`app.go`、`go.mod` | Casbin 引擎就绪，策略落库 |
| **Step 2 RBAC 接入** | `RequirePermission(obj, act)` 中间件替换 `AdminMiddleware`；路由声明式标注权限 | `api/middleware.go`、`api/router.go` | 路由权限声明化，admin 分组消除 |
| **Step 3 DDD 拆包** | `internal/agent/` 上下文独立；本体内部按 domain/service/api 分层 | `internal/` 目录结构 | 限界上下文隔离，各自演进 |
| **Step 4 User 充血** | `domain.User`（私有 `passwordHash` + `SetPassword`/`VerifyPassword`/`ChangeRole`）+ `NewUser` 工厂；service 变薄；`Role` 值对象 | `model/user.go`、`service/user.go` | 贫血模型消除，行为内聚 |
| **Step 5 回归验证** | go build/vet/test + API 冒烟 + Playwright 回归（`create-viewer-user.spec.ts`） | 全链路 | 行为不变，无回归 |

### 9.2 待定问题决定（2026-08-08 用户评审）

- [x] **引入 Casbin**（用户修正选择，自研方案作废）
- [x] **第一期不做角色管理前端界面**：策略用代码 seed 预置到 `casbin_rule` 表
- [x] **保留 viewer（只读）**，新增预置 `operator` 中间角色（可管理 Agent，不可管理用户/Webhook）

### 9.3 验收标准

1. 所有现有 API 行为不变：admin 全通、viewer 只读语义不倒退
2. 新增 `operator` 角色可管理 Agent，访问用户/Webhook 接口返回 403
3. User 实体行为内聚，service 无业务规则残留（不再有 bcrypt/uuid 逻辑）
4. go build/vet/test 全过；Playwright 回归（创建观察者用户）通过

### 9.4 里程碑

- **M1（RBAC 可用）**：Step 1-2 完成，`RequirePermission` 全量替换
- **M2（DDD 阶段①）**：Step 3-4 完成，限界上下文拆分 + User 充血
- **M3（验收）**：Step 5 完成，回归通过，更新 `development-guide.md`

### 9.5 二期规划：角色权限配置页面（暂缓，2026-08-10 用户确认）

> 现状：策略由代码 `rolePolicies()` seed 预置，改权限需改代码 + 重启。
> 二期目标：前端可视化配置角色权限，热生效。Casbin 架构已预留此能力（策略在 DB + `Reload()`），无需改动判断链路。

| 项 | 设计要点 |
|----|----------|
| 后端 | `GET/PUT /api/settings/roles`、`GET/PUT /api/settings/policies`（仅 admin，权限点如 `settings:manage`）；写操作执行 `AddPolicy/RemovePolicy` + `Reload()` 热生效，无需重新登录 |
| 前端 | 设置页新增"角色权限"Tab：角色 × 权限点矩阵，勾选保存（AntD Table + Checkbox） |
| 权限点 | 复用 §8.4/§8.5 权限点清单（`obj:act`），前端矩阵按权限点渲染 |
| 前置 | Step 2（RequirePermission 全量接入）完成；Phase 3 前端工程化（可选） |
| 状态 | **暂缓**——单人项目 + 角色固定，当前代码 seed 足够；出现"自定义角色"需求时启动 |

---

## 十、相关文档

- [开发指南](development-guide.md) — 本地启动、前端模式、常见问题、路线图
- [RBAC/Casbin 详细设计](rbac-casbin-design.md) — 权限模型落地细节
- [告警 Webhook 设计](alert-webhook-design.md) — 告警推送设计
- [环境搭建 (macOS)](env-setup-macos.md) — 从零配置开发环境
- [API 契约](../../spec/v1-api.yaml) — OpenAPI 3.0 接口定义
- [UI 视觉规范](../../spec/ui-theme.md) — 色值、布局、组件设计
