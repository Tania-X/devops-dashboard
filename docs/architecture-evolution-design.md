# 架构演进设计(DDD 演进 + RBAC 权限)

> 状态：设计稿(DDD 部分方向已确认,RBAC 部分待评审)
> 创建：2026-08-08
> 技术栈：Go 后端(gin + GORM + SQLite)

---

## 一、DDD 演进设计

### 1.1 背景

- 现状：Transaction Script 架构 + **贫血模型**(5 个 model 纯字段，业务逻辑集中在 Service)
- 触发：用户学习 DDD，提出两个方向：
  1. 后端本体与 Agent 边界清晰，可各自演进
  2. 将 User 等模型设计为充血模型，行为内聚

### 1.2 设计决策：渐进式 DDD

| 阶段 | 内容 | 成本 | 时机 |
|------|------|------|------|
| ① 边界 + 充血 | 按限界上下文拆包；User 等实体改充血模型 | 低 | **现在可做** |
| ② 仓储接口 | domain 定义 Repository 接口，infrastructure 用 GORM 实现 | 中 | 业务复杂化后 |
| ③ 完整战术 DDD | 聚合、防腐层、领域事件、CQRS | 高 | 暂不做(过度设计) |

> 当前规模(5 model / 12 API)配全量 DDD 属于过度设计；阶段 ① 消除真实的贫血反模式，且不引入额外抽象成本。

### 1.3 限界上下文划分

- **上下文 A：DevOps 本体**(`backend/internal/api`)— 用户/服务器/部署/日志/监控/告警；部署为 Docker 镜像
- **上下文 B：Agent**(`cmd/agent` 独立二进制)— 宿主机指标采集；直连宿主机、不走 Docker

二者通过 HTTP 通信(AGENT_HOSTS)，已有物理边界，需在代码层(`internal/`)按上下文隔离，各自独立演进。

### 1.4 充血模型原则

1. **实体 = 状态 + 状态变更行为**，行为内聚在实体(如 `User.SetPassword` / `VerifyPassword` / `ChangeRole`)
2. **创建/删除归工厂 + 应用服务**，不进实体(创建时无实例状态可言)
3. **敏感字段私有化**(如 `passwordHash` 私有，外部只能通过 `VerifyPassword` 校验)
4. Service 从"上帝类"瘦身为**编排者**(NewUser → repo.Save → 返回)

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

### 1.5 演进目标结构

```
backend/internal/
├── agent/          # Agent 上下文(独立演进)
└── dashboard/      # 本体上下文
    ├── domain/     # 实体 + 值对象 + 工厂(不依赖 gin/gorm)
    ├── service/    # 应用服务(编排 + 事务)
    ├── api/        # HTTP handler(参数解析 → 应用服务 → 响应)
    └── monitor/    # 采集器(基础设施, 保持现状)
```

### 1.6 风险与权衡

- 阶段 ① 涉及文件移动与 model 改造，接口语义保持不变，可渐进式提交
- `AlertBus`(notify 包)已是**领域事件雏形**，阶段 ③ 可平滑升级，无需重写
- 试点：User 先行，Server/Deployment 等按需跟进，不强制全部改造

---

## 二、RBAC 权限设计(建议稿，待评审)

### 2.1 现状问题

1. **路由硬编码**：`router.go` 中 `admin := auth.Group("") + AdminMiddleware`，按路由分组生硬分离 admin/非 admin
2. **二值权限**：只有 admin / viewer，无法表达中间态(如"可管理 Agent 但不可管理用户")
3. **权限与路由耦合**：新增能力要改路由注册与中间件
4. JWT claims 带 `role`，但 `ValidateToken` 实际查 DB(claims 中 role 冗余)

### 2.2 目标

1. **权限点驱动**：每个 API 标注所需权限，如 `user:read` / `user:write` / `agent:manage` / `webhook:manage`
2. **角色 → 权限集合**，支持自定义角色(admin / viewer / operator…)
3. **中间件通用化**：`RequirePermission("user:create")` 取代 `AdminMiddleware`
4. 权限变更**即时生效**(不需要重新登录)

### 2.3 方案对比

| 方案 | 说明 | 适用场景 |
|------|------|----------|
| **Casbin** | Go 最成熟授权库；PERM 元模型，支持 RBAC/ABAC/ACL；策略可存 DB/文件；gin 集成 `github.com/casbin/gin-authz` | 复杂授权需求、团队项目 |
| **自研轻量 RBAC(推荐)** | `permissions` + `role_permissions` 表，权限点常量，启动加载内存 map，中间件 O(1) 查 | 本项目：可控、教学价值高、零依赖 |

> 选择自研：项目权限模型简单(约 10 个权限点)，Casbin 的元模型/策略语言属于额外学习成本；自研约百行代码即可覆盖，且能清晰理解 RBAC 全貌。

### 2.4 自研方案设计

**数据模型：**

```
users.role(string, 保留现状)         -- admin / viewer / operator...
roles.id, roles.code, roles.name
permissions.id, permissions.code     -- "user:read" / "agent:manage" ...
role_permissions.role_id, role_id.permission_id
```

**权限点清单(初稿)：**

| 权限点 | 说明 | admin | viewer | operator |
|--------|------|-------|--------|----------|
| `dashboard:view` | 查看仪表盘/趋势/告警 | ✅ | ✅ | ✅ |
| `server:read` | 查看服务器/日志/部署 | ✅ | ✅ | ✅ |
| `agent:manage` | 管理 Agent | ✅ | ❌ | ✅ |
| `user:read` | 查看用户列表 | ✅ | ❌ | ❌ |
| `user:write` | 创建/编辑/删除用户 | ✅ | ❌ | ❌ |
| `webhook:manage` | 配置 Webhook | ✅ | ❌ | ❌ |

**中间件设计：**

```go
// authz 包：启动时加载 role → permissions map，HasPermission O(1) 查
func RequirePermission(code string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := c.MustGet("currentUser").(*model.User)
        if !authz.HasPermission(user.Role, code) {
            ErrorJSON(c, http.StatusForbidden, "权限不足")
            c.Abort()
            return
        }
        c.Next()
    }
}
```

**路由声明式标注：**

```go
auth.GET("/users", h.GetUserList, RequirePermission("user:read"))
auth.POST("/users", h.CreateUser, RequirePermission("user:write"))
auth.POST("/agents", h.CreateAgent, RequirePermission("agent:manage"))
```

**权限即时生效：** token 只携带 `sub`(用户 ID)+ `role`，服务端每次从内存角色-权限表查询 → 修改角色权限立即生效，无需重新登录。

### 2.5 与 DDD 的结合

- `Role` 作为值对象/实体，`User.ChangeRole` 校验角色合法性
- 权限模型在 domain 层建模(roles/permissions 聚合)，infrastructure 持久化
- 依赖方向：HTTP → 应用服务 → 领域(authz 判断可下沉为领域服务)

### 2.6 待定问题

- [ ] 是否引入 Casbin(若后续权限复杂如 ABAC/多租户再升级)
- [ ] 角色是否可配置(需要前端管理界面？)
- [ ] 现有 viewer 角色保留为"只读"语义，是否拆分出 operator 中间角色

---

## 三、相关文档

- [架构总览](architecture.md) — 当前分层与决策
- [开发指南](development-guide.md) — 路线图(Phase 3/4/5)
- [告警 Webhook 设计](alert-webhook-design.md) — 告警推送设计
