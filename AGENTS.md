# 项目测试约定

## 首条原则：测试是代码的"防手贱"机制

一个测试存在的唯一理由：**防止未来有人（包括你自己）改代码时引入 bug**。

不是"演示功能正常"，不是"凑覆盖率"。如果在重构某段代码时删了这个测试、功能依然没 bug，那这个测试从一开始就不该写。

## 可测性第一

设计代码时先问自己：**"这段逻辑不依赖外部系统，我能直接注入数据来测吗？"**

```go
// ✅ 可测性好——被测对象只依赖传进来的参数
func (a *Alerter) Evaluate(snapshot *MetricSnapshot) { ... }

// ✅ 可测性好——可以在测试里的 History 塞假数据
func (h *History) Record(p HistoryPoint) { ... }
```

如果一段逻辑的非纯函数部分太多（有 DB 查询、外部 HTTP 调用等），先重构为可注入的依赖，再写业务逻辑。

## 分类方式：从被测代码特征推导测试模式

```
被测代码的每次调用是否独立于之前的调用？
├── 是 → 表驱动测试（单指标场景）
└── 否 → 独立函数，按时间线写清每一步（多状态场景）

被测代码是否被并发调用？
└── 是 → 独立并发测试（只验"会不会崩"）
```

### 表驱动（纯函数行为）

适用场景：输入 → 输出，不依赖历史状态。

```go
func TestAlerter_SingleMetric(t *testing.T) {
    cases := []struct {
        name      string
        snapshot  MetricSnapshot
        wantLevel string
        wantAlert bool
    }{
        {"cpu_normal",    MetricSnapshot{CPUPercent: 30}, "",       false},
        {"cpu_warning",   MetricSnapshot{CPUPercent: 65}, "warning", true},
        {"cpu_critical",  MetricSnapshot{CPUPercent: 85}, "critical", true},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            a := NewAlerter()
            a.Evaluate(&tc.snapshot)
            alerts := a.GetAlerts(10)

            if !tc.wantAlert {
                if len(alerts) != 0 { t.Fatalf(...) }
                return
            }
            if alerts[0].Level != tc.wantLevel { t.Errorf(...) }
        })
    }
}
```

### 独立函数（有状态行为）

适用场景：第二步结果依赖第一步，需要精确控制时间线。

```go
func TestAlerter_Escalate(t *testing.T) {
    a := NewAlerter()

    a.Evaluate(...)    // 第 1 步
    a.Evaluate(...)    // 第 2 步

    alerts := a.GetAlerts(10)
    // 断言
}
```

### 并发安全测试

适用场景：数据结构被多个 goroutine 共享。

目的：**只验证"会不会崩"**——如果 `Lock()`/`RLock()` 位置有误，测试立刻暴露。不验证告警内容是否正确（并发顺序不可控）。

```go
func TestAlerter_Concurrent(t *testing.T) {
    a := NewAlerter()
    var wg sync.WaitGroup

    // 并行写
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(v float64) {
            defer wg.Done()
            for j := 0; j < 100; j++ {
                a.Evaluate(&MetricSnapshot{CPUPercent: v})
            }
        }(float64(30 + i))
    }

    // 并行读
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < 100; j++ {
                a.GetAlerts(10)
            }
        }()
    }

    wg.Wait()
}
```

## 断言规则

### Fatalf vs Errorf

```go
// Fatalf — 立即停止，后续断言没意义
if len(alerts) == 0 {
    t.Fatalf("期望有告警，但结果为空")  // 没有告警，后面判断 Level 会 nil panic
}

// Errorf — 标记失败但继续，其他断言仍执行
if alerts[0].Level != "critical" {
    t.Errorf("期望 critical，得到 %s", alerts[0].Level)
}
```

### 断言消息格式

```go
t.Errorf("期望 X，得到 Y")          // ✅ 明确：预期值 vs 实际值
t.Errorf("something wrong")         // ❌ 含糊：看不出哪不对
```

### 使用子测试

```go
for _, tc := range cases {
    t.Run(tc.name, func(t *testing.T) {  // ← 每个 case 是独立子测试
        // ...
    })
}
```

这样 `go test -run "TestAlerter_SingleMetric/cpu_warning"` 可以只跑一个 case。

## 边界值覆盖

阈值判断的 off-by-one 是最常见的 bug，表驱动里必须覆盖边界值：

```go
{"cpu_boundary_60",  MetricSnapshot{CPUPercent: 60}, "warning", true},   // 刚好等于阈值
{"cpu_boundary_80",  MetricSnapshot{CPUPercent: 80}, "critical", true},
```

建议：每个阈值测三个值（阈值-1，阈值，阈值+1）。

## 文件名

测试代码放在与实现代码相同的包中：

```
backend/internal/monitor/
├── alerter.go           # package monitor
├── alerter_test.go      # package monitor（同一个包，可以访问未导出字段）
```

不需要改名，不需要子目录。`go test ./internal/monitor/` 自动发现所有 `*_test.go` 文件。

## 运行方式

```bash
go test ./internal/monitor/              # 整个包
go test -run TestAlerter_Normal ./...    # 单个测试
go test -v -race ./internal/monitor/     # 带竞态检测（需要 CGO）
```

## 不写的测试

以下情况不写测试：

- `func main()` — 入口函数，纯胶水代码
- 路由注册函数（如 `SetupRouter()`） — 只做注册，没有业务逻辑
- GORM 数据库查询 — 需要真实 SQLite 或 mock，测试价值/维护成本不划算（此项目暂不涉及）

## 什么时候应当加测试

不是每一个函数都要测。按优先级从高到低：

1. 有复杂逻辑的函数（`evaluateMetric` 的阈值判断分支）
2. 有内部状态变化的函数（`Alerter` 的 prevStatus 更新 / 去重）
3. 返回值有可能被误解的函数（返回 `nil` 和返回空 `[]AlertItem{}` 含义不同）
4. 纯胶水代码（路由注册、main 组装依赖） — 不测

## 何时写并发测试

满足三个条件才需要：

1. 数据结构被多个 goroutine 共享
2. 有手动管理的 `sync.Mutex` / `sync.RWMutex`
3. 读取和写入是独立的调用入口（不是同一个函数内部自创的 goroutine）

本项目仅 `Alerter` 满足此条件。

## 测试命令

```bash
go test ./internal/monitor/         # 当前包
go test -run TestAlerter ./...      # 指定测试
go test -v ./internal/...           # 所有 internal 下的测试
go test -count=1 ./...               # 确保跑一次，不是缓存结果
```
