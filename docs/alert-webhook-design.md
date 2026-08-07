# 告警 Webhook 推送设计

> 状态：设计稿（待评审）
> 关联 Spec：`spec/v1-api.yaml`（webhook 配置端点）
> 技术栈：Go 后端（gin）+ React 前端（antd）

---

## 一、背景与目标

### 现状问题

当前告警引擎（`backend/internal/monitor/alerter.go`）已能正确评估 CPU/内存/磁盘阈值并生成告警，但告警**只存在内存中**，唯一出口是 Dashboard 页面的告警列表（`GET /api/dashboard/alerts`）。

**没有人实时看到告警** → 服务器 CPU 飙到 90%，要等人打开 Dashboard 才能发现，告警系统失去"预警"价值。

### 目标

1. 告警产生时**实时推送到外部通知渠道**（钉钉/企业微信机器人 Webhook）
2. 支持在**前端页面配置** Webhook 地址与类型（不依赖改环境变量重启）
3. 推送失败不阻塞主流程，失败可重试
4. 保持与现有架构一致：SDD 契约优先、配置驱动

### 非目标

- 不实现消息队列（当前单机场景，goroutine + 重试足够）
- 不做多渠道聚合（先支持钉钉 + 企微，架构预留扩展）

---

## 二、现有告警链路分析

```
Collector 采集（10s 一次）
    ↓
History.StartCollector 后台协程
    ↓
Alerter.Evaluate(snapshot)  ← 阈值评估（去重、恢复检测）
    ↓
Alerter.addAlert() → alerts 缓存（内存，最多 20 条）
    ↓
GET /api/dashboard/alerts → 前端展示
```

**关键事实**：
- `Alerter` 是纯内存结构体，`Evaluate` 是同步方法
- 告警去重逻辑：`prevStatus` 状态机，**状态变化才产生告警**（不会重复轰炸）
- 有"已恢复"告警（level=info）
- 配置从环境变量注入（`config.Load()`）

---

## 三、设计方案

### 3.1 架构：告警总线 + Notifier 解耦

**核心思路**：Alerter 产生告警后，通过一个**内部 channel（告警总线）**异步分发给多个 `Notifier`，每个 Notifier 负责一种渠道。发送失败由 Notifier 内部重试，不阻塞 Alerter。

```
Alerter.addAlert(entry)
    ↓ 同步
AlertBus.Publish(entry)   ← 内部 channel（缓冲 100）
    ↓ 异步（goroutine 消费）
WebhookNotifier.Send(entry)  ← 按类型格式化 → HTTP POST → 失败重试 3 次
```

**为什么用 channel 而非直接在 Alerter 里调用 HTTP**：
- Alerter.Evaluate 是采集主链路，**不能因为网络慢/超时拖慢采集**
- 解耦：以后加邮件/短信 Notifier 只需新实现接口
- 缓冲 + 丢弃策略：渠道故障时不堆积阻塞

### 3.2 模块划分（新增文件）

```
backend/internal/notify/
├── notifier.go        # Notifier 接口 + AlertBus（channel 分发）
├── webhook.go         # WebhookNotifier：钉钉/企微格式转换 + HTTP 发送
├── webhook_test.go    # 格式转换 + 发送测试
backend/internal/api/
└── settings.go        # GET/PUT /api/settings/webhook + POST .../test
backend/internal/model/
└── webhook.go         # WebhookConfig 模型（存数据库）
```

### 3.3 核心接口设计

```go
// notifier.go —— 渠道抽象
type Notifier interface {
    Name() string
    Send(entry model.AlertItem) error
}

// AlertBus 告警总线：Alerter → channel → 各 Notifier
type AlertBus struct {
    ch chan model.AlertItem
    notifiers []Notifier
}
func NewAlertBus(notifiers ...Notifier) *AlertBus
func (b *AlertBus) Publish(e model.AlertItem)   // 非阻塞发送，满了丢弃（记日志）
func (b *AlertBus) Run()                          // 后台消费 channel，逐个分发
```

```go
// webhook.go —— Webhook 渠道
type WebhookNotifier struct {
    url   string    // 钉钉/企微机器人地址
    kind  string    // "dingtalk" | "wecom"
    http  *http.Client  // 带超时（5s）
}
func (n *WebhookNotifier) Send(e model.AlertItem) error {
    payload := formatByKind(e, n.kind)   // 按渠道格式化 JSON
    // HTTP POST，超时 5s；失败重试 3 次（指数退避）
}
```

### 3.4 WebhookConfig 模型（存 SQLite）

```go
// model/webhook.go —— 存 settings 表，单条记录（ID=1）
type WebhookConfig struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Enabled   bool      `json:"enabled"`    // 开关
    Kind      string    `json:"kind"`       // dingtalk | wecom
    URL       string    `json:"url"`        // 机器人 Webhook 地址
    Secret    string    `json:"-"`          // 加签密钥（钉钉安全设置，不回传）
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

### 3.5 消息格式

**钉钉机器人**（`kind=dingtalk`）：

```json
{
  "msgtype": "markdown",
  "markdown": {
    "title": "【DevOps 告警】CPU 使用率异常",
    "text": "### DevOps Dashboard 告警\n\n- **级别**: 🔴 critical\n- **指标**: CPU 使用率 87.5%\n- **来源**: localhost\n- **时间**: 08-07 15:30\n\n> 触发条件：超过 critical 阈值 (80%)"
  }
}
```

**企业微信机器人**（`kind=wecom`）：

```json
{
  "msgtype": "markdown",
  "markdown": {
    "content": "**【DevOps 告警】CPU 使用率异常**\n> 级别: 🔴 critical\n> 指标: CPU 使用率 87.5%\n> 来源: localhost\n> 时间: 08-07 15:30"
  }
}
```

**级别映射**：critical → 🔴，warning → 🟡，info（恢复）→ 🟢

### 3.6 配置存储与加载流程

```
启动时：
  config.Load() 读环境变量（默认空 = 不启用）
    ↓
  DB 读取 WebhookConfig（ID=1）
    ↓
  有配置 → 创建 WebhookNotifier → 注册进 AlertBus
    ↓
  前端改配置 → PUT /api/settings/webhook → 存 DB → 重建 Notifier（热更新）
```

**要点**：配置以 **DB 为准**（前端可改），环境变量仅作启动兜底。改动 API 后 Notifier 需要重建——用一个 `webhookManager` 封装"当前生效的 Notifier"引用，改配置后原子替换（sync.RWMutex 保护）。

### 3.7 失败处理与重试

| 场景 | 处理 |
|------|------|
| 网络超时/连接失败 | 重试 3 次，指数退避（1s/2s/4s），仍失败记 error 日志 |
| channel 已满 | 丢弃新告警，记 warning 日志（不阻塞采集） |
| 机器人返回非 200 | 同上重试逻辑 |
| 未配置 Webhook | 跳过，零开销（不存在 Notifier） |

---

## 四、API 设计（见 spec/v1-api.yaml）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/settings/webhook` | 获取当前 Webhook 配置（secret 不回传） |
| PUT | `/api/settings/webhook` | 更新配置（url/kind/enabled），热生效 |
| POST | `/api/settings/webhook/test` | 发送测试告警到配置的 URL，验证连通 |

权限：`admin` 角色可读写（viewer 只读）。

---

## 五、前端页面设计

**新页面**：`frontend/src/features/settings/` → `SettingsPage.tsx`（或挂到现有"设置"入口）

```
┌─ 告警通知设置 ─────────────────────────────┐
│  启用推送        [开关]                     │
│  渠道类型        (●) 钉钉  ( ) 企业微信      │
│  Webhook 地址    [____________________]     │
│  （加签密钥       [____] 可选）              │
│  [测试推送]  [保存]                          │
└──────────────────────────────────────────┘
```

- 保存 → PUT API → message.success("已保存")
- 测试推送 → POST test → 渠道收到测试消息 → message.success("已发送，请检查群里")
- 侧边栏新增"设置"入口（或并入现有导航）

---

## 六、测试计划

| 层 | 用例 |
|----|------|
| 单测：webhook 格式 | 钉钉/企微 payload 结构正确（markdown 字段、级别 emoji 映射） |
| 单测：AlertBus | Publish 分发到所有 Notifier；channel 满不阻塞；顺序性 |
| 单测：重试 | 模拟 500/超时 → 重试 3 次后失败，返回 error 不 panic |
| 集成：API | GET/PUT/POST test 端点（admin 权限、secret 不泄露） |
| 手动验证 | 配置钉钉机器人 → 触发告警（阈值调低）→ 群里收到消息 |

---

## 七、实施步骤（后续）

1. `model/webhook.go` + AutoMigrate
2. `notify/notifier.go` + `notify/webhook.go` + 单测
3. `api/settings.go`（GET/PUT/test）
4. `webhookManager` 热更新接入 app 启动流程
5. `swag` 重新生成 docs
6. 前端 SettingsPage + 侧边栏入口
7. Playwright 冒烟

## 八、验证方式（验收标准）

1. 单测全绿（`go test ./...`）
2. 前端配置钉钉机器人地址 → 点"测试推送" → 钉钉群收到消息
3. 调低 CPU 阈值（或手动触发）→ 告警产生 → 钉钉群收到真实告警
4. 停掉 Webhook 服务 → 告警不阻塞采集（Dashboard 正常刷新）
