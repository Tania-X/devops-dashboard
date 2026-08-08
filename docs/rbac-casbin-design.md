# RBAC / Casbin 详细设计

> 状态：详细设计（实施依据）
> 创建：2026-08-08
> 上游文档：[架构文档](architecture.md) §八、§九（Step 1-2）
> 技术栈：Go（gin + GORM + SQLite）+ Casbin v2

---

## 一、背景与目标

### 1.1 现状问题

1. **路由硬编码**：`router.go` 中 `admin := auth.Group("") + AdminMiddleware`，按路由分组生硬分离 admin/非 admin
2. **二值权限**：只有 admin / viewer，无法表达中间态（如"可管理 Agent 但不可管理用户"）
3. **权限与路由耦合**：新增能力要改路由注册与中间件
4. **无按钮级控制**：前端无法按权限控制按钮显隐

### 1.2 目标

1. 权限点驱动：每个 API 标注所需权限（`obj + act`）
2. 角色 → 权限集合：admin / viewer / operator
3. 通用中间件 `RequirePermission(obj, act)` 取代 `AdminMiddleware`
4. 权限变更即时生效（无需重新登录）
5. 前端按钮级控制（UX 层），后端接口校验为安全边界（纵深防御）

### 1.3 非目标

- 不做角色管理前端界面（策略用代码 seed 预置）
- 不做 ABAC / 多租户（Casbin 已支持，后续扩展）
- 不做权限审计日志（后续可加）

---

## 二、Casbin 概念速览

| 概念 | 说明 | 本项目用法 |
|------|------|-----------|
| **PERM 模型** | Policy / Effect / Request / Matcher 四要素 | `model.conf` 定义 |
| **Request** | 访问请求 `(sub, obj, act)` | 用户角色、资源、动作 |
| **Policy** | 授权策略 `p = sub, obj, act` | 角色拥有的权限 |
| **Matcher** | 匹配规则 | `g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act` |
| **Role 继承** | `g = _, _` | 角色 → 权限的关系 |
| **Adapter** | 策略持久化 | gorm-adapter 存 SQLite `casbin_rule` 表 |

判断流程：`Enforce(role, obj, act)` → 按 matcher 扫描策略 → 命中返回 true。

---

## 三、依赖

```bash
go get github.com/casbin/casbin/v2        # 授权引擎
go get github.com/casbin/gorm-adapter/v3  # GORM 策略适配器(SQLite)
```

> gin-authz（`github.com/casbin/gin-authz`）可选：本项目自写 10 行中间件封装，不引入额外依赖。

---

## 四、模型定义（model.conf）

`backend/internal/authz/model.conf`（嵌入二进制，用 `//go:embed`）：

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && keyMatch(r.act, p.act)
```

> ⚠️ **踩坑记录（2026-08-08 实施时发现）**：
> - policy 里的 `*`（如 `p, admin, *, *`）**不是通配符**，字面比较 `r.obj == p.obj` 永远匹配不上；必须用 Casbin 内置函数 `keyMatch`（支持 `*` 通配）才能实现 admin 全通
> - 模型内容用 `model.NewModelFromString(modelText)` 加载（`NewEnforcer` 第一参是**文件路径**，不能直接传字符串）
> - `gorm-adapter/v3` 需使用 **v3.14.x**（适配 casbin/v2）；最新 v3.41.0 已切到 casbin/v3 接口（`missing method LoadPolicy` panic）

---

## 五、authz 包设计

`backend/internal/authz/authz.go`：

```go
package authz

import (
    _ "embed"

    "github.com/casbin/casbin/v2"
    "github.com/casbin/gorm-adapter/v3"
    "gorm.io/gorm"
)

//go:embed model.conf
var modelText string

// Enforcer 全局单例(启动初始化,进程内唯一)
var enforcer *casbin.Enforcer

// Init 初始化:模型 + gorm adapter(自动建 casbin_rule 表) + seed 预置策略
func Init(db *gorm.DB) error {
    adapter, err := gormadapter.NewAdapterByDB(db)
    if err != nil {
        return err
    }
    e, err := casbin.NewEnforcer(modelText, adapter)
    if err != nil {
        return err
    }
    enforcer = e
    return seedIfNeeded(db)
}

// HasPermission 权限判断(线程安全,Casbin 内部带锁)
func HasPermission(role, obj, act string) (bool, error) {
    return enforcer.Enforce(role, obj, act)
}

// PermissionsOf 返回某角色拥有的全部权限点(供 /auth/me 使用)
func PermissionsOf(role string) ([]string, error) {
    // 遍历 policy 中 sub == role 的 (obj, act),拼成 "obj:act"
    // admin 命中通配 * 时展开为全部权限点常量
    ...
}

// Reload 策略变更后重载(后续做管理界面时用)
func Reload() error { return enforcer.LoadPolicy() }
```

### 5.1 策略 seed（`seedIfNeeded`）

启动时检查 `casbin_rule` 是否为空，空则写入（角色权限矩阵见 §六）：

| 角色 | 策略 |
|------|------|
| admin | `p, admin, *, *`（全通） |
| viewer | `dashboard:view`、`server:read`、`log:read`、`deployment:read` |
| operator | `dashboard:view`、`server:read`、`log:read`、`deployment:read`、`agent:read`、`agent:create`、`agent:update`、`agent:delete`、`agent:deploy`、`agent:stop` |

```go
func seedIfNeeded(db *gorm.DB) error {
    var count int64
    db.Table("casbin_rule").Count(&count)
    if count > 0 {
        return nil
    }
    policies := [][]string{
        {"admin", "*", "*"},
        {"viewer", "dashboard", "view"}, {"viewer", "server", "read"},
        {"viewer", "log", "read"}, {"viewer", "deployment", "read"},
        {"operator", "dashboard", "view"}, {"operator", "server", "read"},
        {"operator", "log", "read"}, {"operator", "deployment", "read"},
        {"operator", "agent", "read"}, {"operator", "agent", "create"},
        {"operator", "agent", "update"}, {"operator", "agent", "delete"},
        {"operator", "agent", "deploy"}, {"operator", "agent", "stop"},
    }
    _, err := enforcer.AddPolicies(policies)
    return err
}
```

---

## 六、权限点清单（obj × act）

### 6.1 接口级（后端强制校验）

| 权限点 | obj | act | 覆盖接口 | admin | viewer | operator |
|--------|-----|-----|----------|-------|--------|----------|
| `dashboard:view` | dashboard | view | GET /api/dashboard/* | ✅ | ✅ | ✅ |
| `server:read` | server | read | GET /api/servers* | ✅ | ✅ | ✅ |
| `log:read` | log | read | GET /api/logs | ✅ | ✅ | ✅ |
| `deployment:read` | deployment | read | GET /api/deployments* | ✅ | ✅ | ✅ |
| `monitor:read` | monitor | read | GET /api/monitor/* | ✅ | ✅ | ✅ |
| `agent:read` | agent | read | GET /api/agents* | ✅ | ❌ | ✅ |
| `agent:create` | agent | create | POST /api/agents | ✅ | ❌ | ✅ |
| `agent:update` | agent | update | PUT /api/agents/:id | ✅ | ❌ | ✅ |
| `agent:delete` | agent | delete | DELETE /api/agents/:id | ✅ | ❌ | ✅ |
| `agent:deploy` | agent | deploy | POST /api/agents/:id/deploy | ✅ | ❌ | ✅ |
| `agent:stop` | agent | stop | POST /api/agents/:id/stop | ✅ | ❌ | ✅ |
| `user:read` | user | read | GET /api/users | ✅ | ❌ | ❌ |
| `user:create` | user | create | POST /api/users | ✅ | ❌ | ❌ |
| `user:update` | user | update | PUT /api/users/:id | ✅ | ❌ | ❌ |
| `user:delete` | user | delete | DELETE /api/users/:id | ✅ | ❌ | ❌ |
| `webhook:read` | webhook | read | GET /api/settings/webhook | ✅ | ❌ | ❌ |
| `webhook:update` | webhook | update | PUT /api/settings/webhook | ✅ | ❌ | ❌ |
| `webhook:test` | webhook | test | POST /api/settings/webhook/test | ✅ | ❌ | ❌ |

> 公开接口（health、dashboard 只读页可公开）不在权限范围，保持现状。

### 6.2 按钮级（前端 UX，权限点同上）

前端按钮 → 权限点映射（供 `AuthButton` 使用）：

| 页面 | 按钮 | 权限点 |
|------|------|--------|
| 用户管理 | 新增用户 | `user:create` |
| 用户管理 | 编辑 | `user:update` |
| 用户管理 | 删除 | `user:delete` |
| Agent 管理 | 新增 Agent | `agent:create` |
| Agent 管理 | 部署 / 停止 | `agent:deploy` / `agent:stop` |
| 设置 | 保存 Webhook | `webhook:update` |
| 设置 | 发送测试 | `webhook:test` |

---

## 七、中间件设计

`backend/internal/api/middleware.go` 新增（替换 `AdminMiddleware`）：

```go
// RequirePermission 权限校验中间件:校验当前用户角色是否拥有 (obj, act) 权限
func (h *Handler) RequirePermission(obj, act string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user, exists := c.Get("currentUser")
        if !exists {
            ErrorJSON(c, http.StatusUnauthorized, "未认证")
            c.Abort()
            return
        }
        u := user.(*model.User)
        ok, err := authz.HasPermission(u.Role, obj, act)
        if err != nil || !ok {
            ErrorJSON(c, http.StatusForbidden, "权限不足")
            c.Abort()
            return
        }
        c.Next()
    }
}
```

`AuthMiddleware` 保持不变（JWT 校验 + currentUser 注入）；**删除 `AdminMiddleware`**（不再需要）。

---

## 八、路由改造清单

`router.go` 中 admin 分组 → 逐路由标注权限：

```go
auth := api.Group("")
auth.Use(h.AuthMiddleware())
{
    auth.GET("/auth/me", h.GetMe)

    // 用户管理(原 admin 分组)
    auth.GET("/users", h.GetUserList, h.RequirePermission("user", "read"))
    auth.POST("/users", h.CreateUser, h.RequirePermission("user", "create"))
    auth.PUT("/users/:id", h.UpdateUser, h.RequirePermission("user", "update"))
    auth.DELETE("/users/:id", h.DeleteUser, h.RequirePermission("user", "delete"))

    // Agent 管理(operator 可用)
    auth.GET("/agents", h.GetAgentList, h.RequirePermission("agent", "read"))
    auth.POST("/agents", h.CreateAgent, h.RequirePermission("agent", "create"))
    auth.PUT("/agents/:id", h.UpdateAgent, h.RequirePermission("agent", "update"))
    auth.DELETE("/agents/:id", h.DeleteAgent, h.RequirePermission("agent", "delete"))
    auth.POST("/agents/:id/deploy", h.DeployAgent, h.RequirePermission("agent", "deploy"))
    auth.POST("/agents/:id/stop", h.StopAgent, h.RequirePermission("agent", "stop"))
    auth.GET("/agents/:id/status", h.CheckAgentStatus, h.RequirePermission("agent", "read"))

    // Webhook 配置(仅 admin)
    auth.GET("/settings/webhook", h.GetWebhookConfig, h.RequirePermission("webhook", "read"))
    auth.PUT("/settings/webhook", h.UpdateWebhookConfig, h.RequirePermission("webhook", "update"))
    auth.POST("/settings/webhook/test", h.TestWebhookConfig, h.RequirePermission("webhook", "test"))
}
```

> 改动后：`admin := auth.Group("")` 分组与 `AdminMiddleware` 删除，`/auth/me` 保持登录即可访问。

---

## 九、前端权限接入

### 9.1 后端返回权限点

`GET /api/auth/me` 响应增加 `permissions: string[]`（登录响应 `LoginResponse` 同样带上）：

```go
// AuthService.Login / GetMe 中计算
permissions, _ := authz.PermissionsOf(user.Role)
// LoginResponse / MeResponse 增加字段:
//   permissions []string `json:"permissions"`
```

> admin 命中通配 `*`：`PermissionsOf("admin")` 展开为权限点常量全表（§6.1 全部权限点）。

### 9.2 前端实现

`frontend/src/contexts/AuthContext.tsx` 保存 `permissions: string[]`；新增两个工具：

```tsx
// src/hooks/usePermission.ts
export function usePermission(perm: string): boolean {
  const { permissions } = useAuth();
  return permissions.includes(perm);
}

// src/components/AuthButton.tsx
export function AuthButton({ perm, children, ...props }: { perm: string } & ButtonProps) {
  const can = usePermission(perm);
  if (!can) return null;               // 无权限不渲染(或 <Tooltip>禁用)
  return <Button {...props}>{children}</Button>;
}
```

使用示例（UserPage 删除按钮）：

```tsx
<AuthButton perm="user:delete" size="small" danger onClick={() => ...}>
  删除
</AuthButton>
```

---

## 十、测试与验收

### 10.1 单元测试（authz 包）

| 用例 | 断言 |
|------|------|
| `HasPermission("viewer", "user", "read")` | false |
| `HasPermission("operator", "agent", "deploy")` | true |
| `HasPermission("operator", "user", "read")` | false |
| `HasPermission("admin", "anything", "anything")` | true（通配） |
| `PermissionsOf("viewer")` | 含 `dashboard:view`、`server:read`，不含 `user:read` |

### 10.2 集成验证（Playwright / curl）

| 场景 | 预期 |
|------|------|
| admin 登录 → 全部接口 | 200 |
| viewer 登录 → GET /api/users | 403 |
| operator 登录 → POST /api/agents | 201 |
| operator 登录 → POST /api/users | 403 |
| operator 登录 → 用户页无"新增用户"按钮 | 按钮不渲染 |

### 10.3 验收标准

1. 现有 API 行为不变：admin 全通、viewer 只读不倒退
2. operator 可管理 Agent，访问用户/Webhook 接口 403
3. `/auth/me` 与登录响应返回 permissions
4. 前端按钮按权限点显隐，与后端接口校验一致
5. go build/vet/test 全过；Playwright 回归（create-viewer-user.spec.ts）通过

---

## 十一、风险与注意事项

1. **Casbin 依赖体积**：+~8MB 依赖，一次性代价
2. **策略与代码漂移**：新增权限点时，需同步更新 seed 策略 + 路由标注 + 前端按钮 —— 权限点清单（§6）是唯一事实源，变更需三处同步
3. **admin 通配策略**：`p, admin, *, *` 使 `PermissionsOf("admin")` 需要展开常量表，实现时注意一致性
4. **权限即时生效**：seed 后策略在 DB，`LoadPolicy()` 重载即可，token 无需重新签发（claims 只带 role）
5. **性能**：Casbin 策略量极小（<20 条），Enforcer 内存缓存，单次 Enforce 微秒级，无需优化

---

## 十二、实施检查单（Step 1-2）

- [ ] `go get` 两个 casbin 依赖
- [ ] `internal/authz/`：model.conf + authz.go（Init/HasPermission/PermissionsOf/Reload + seed）
- [ ] `app.go` Init 中调用 `authz.Init(db)`（Service 注入前）
- [ ] `api/middleware.go`：新增 `RequirePermission`，删除 `AdminMiddleware`
- [ ] `api/router.go`：admin 分组 → 逐路由标注权限
- [ ] `auth.go`：Login/GetMe 返回 permissions
- [ ] 前端：AuthContext 存 permissions + usePermission + AuthButton + 页面按钮替换
- [ ] 单元测试 + Playwright 回归
- [ ] 更新 [架构文档](architecture.md) 状态（待评审 → 已实施）
