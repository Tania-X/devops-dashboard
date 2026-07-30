# Go 后端测试编写约定

## 可测性第一

设计代码时先问自己：**"这段逻辑不依赖外部系统，我能直接注入数据来测吗？"**

```go
// ✅ 可测性好 — 被测对象只依赖传进来的参数
func (a *Alerter) Evaluate(snapshot *MetricSnapshot) { ... }

// ✅ 可测性好 — 可以在测试里的 History 塞假数据
func (h *History) Record(p HistoryPoint) { ... }
```

如果一段逻辑的非纯函数部分太多（有 DB 查询、外部 HTTP 调用等），先重构为可注入的依赖，再写业务逻辑。

## 分类方式：从被测代码特征推导测试模式

```
被测代码的每次调用是否独立于之前的调用？
├── 是 → 表驱动测试（纯函数场景）
└── 否 → 独立函数，按时间线写清每一步（有状态场景）

被测代码是否被并发调用？
└── 是 → 独立并发测试（只验"会不会崩"）
```

### 模式 1：表驱动（纯函数行为）

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

**为什么用表驱动？** 同一逻辑重复试不同数据。正常/警告/临界/边界值都在一张表里，断言逻辑只写一次，加一个新场景只需一行。

**t.Run 子测试**：每个 case 一个标签，`go test -run "TestAlerter_SingleMetric/cpu_warning"` 可以只跑一个 case。

### 模式 2：独立函数（有状态行为）

适用场景：第二步结果依赖第一步，需要精确控制时间线。

```go
func TestAlerter_Deduplication(t *testing.T) {
    a := NewAlerter()

    // 第 1 步：CPU 85% → critical
    a.Evaluate(&MetricSnapshot{CPUPercent: 85})
    // 第 2 步：CPU 85% 没变 → 不重复告警
    a.Evaluate(&MetricSnapshot{CPUPercent: 85})

    if len(a.GetAlerts(10)) != 1 {  // 断言
        t.Fatal("不应重复告警")
    }
}
```

### 模式 3：并发安全测试

适用场景：数据结构被多个 goroutine 共享，有手动管理的 `sync.Mutex` / `sync.RWMutex`。

**目的**：只验证"会不会崩" — 如果 Lock()/RLock() 位置有误，测试立刻暴露。不验证内容正确性（并发顺序不可控）。

```go
func TestAlerter_Concurrent(t *testing.T) {
    a := NewAlerter()
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(v float64) {
            defer wg.Done()
            for j := 0; j < 100; j++ {
                a.Evaluate(&MetricSnapshot{CPUPercent: v})  // 并发写
            }
        }(float64(30 + i))
    }

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < 100; j++ {
                a.GetAlerts(10)  // 并发读
            }
        }()
    }

    wg.Wait()
}
```

## 断言规则

```go
// Fatalf — 立即停止，后续断言没意义（如告警为空时元素会是 nil）
t.Fatalf("期望有告警，但结果为空")

// Errorf — 标记失败但继续检查其他条件
t.Errorf("期望 critical，得到 %s", alerts[0].Level)
```

**消息格式**：`t.Errorf("期望 X，得到 Y")` — 必须包含预期值和实际值。

## 边界值覆盖

阈值判断的 off-by-one 是最常见的 bug，表驱动里必须覆盖边界值：

```go
{"cpu_boundary_60", MetricSnapshot{CPUPercent: 60}, "warning", true},   // 恰等于阈值
{"cpu_boundary_80", MetricSnapshot{CPUPercent: 80}, "critical", true},
```

原则：每个阈值测三个值（阈值-1，阈值，阈值+1）。

## 文件组织

```
backend/internal/monitor/
├── alerter.go           # package monitor
├── alerter_test.go      # package monitor（同一个包，可访问未导出字段）
```

不需要子目录。`go test ./internal/monitor/` 自动发现所有 `*_test.go`。

## 运行方式

```bash
go test ./internal/monitor/              # 整个包
go test -run TestAlerter ./...           # 指定测试
go test -v ./internal/...                # 所有 internal 下测试
go test -count=1 ./...                   # 确保跑一次，不是缓存
```

## 不写的测试

- `func main()` — 入口函数，纯胶水代码
- 路由注册函数（如 `SetupRouter()`） — 只做注册，无业务逻辑
- GORM 数据库查询 — 需真实 SQLite 或 mock，测试价值低（此项目暂不涉及）

## 并发测试的适用条件

满足三个条件才写：

1. 数据结构被多个 goroutine 共享
2. 有手动管理的 `sync.Mutex` / `sync.RWMutex`
3. 读取和写入是独立的调用入口（不是同一个函数内部自创的 goroutine）

本项目仅 `Alerter` 满足此条件。
