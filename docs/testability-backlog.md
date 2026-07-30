# 可测试性重构 TODO

> 记录当前代码中"因外部依赖而不可测试的分支逻辑"。
> 等项目规模变胖（新增 Service、多人参与）后，按此表逐项重构并补测试。

---

## 原则

不是每个函数都要测。靶子是**有分支逻辑但被外部依赖"锁住"的代码**。

重构方式：把"拿数据"和"处理数据拆开"——调用方传数据，被测函数只做纯计算。

---

## 待重构项

### 1. `MonitorService.GetHostInfo()` — CPU 核心数为 0 时的降级逻辑

- **文件**：`internal/service/monitor.go:143`
- **分支**：`if cpuCores == 0 && len(cpuInfo) > 0 { cpuCores = int(cpuInfo[0].Cores) }`
- **当前不可测原因**：`cpu.Counts()` 返回真实 CPU 核心数，开发机永远 > 0
- **重构方向**：抽 `buildHostInfo(cpuInfo, memInfo, uptime) model.HostInfo`，在测试里传 `cpuCores=0` 验证降级
- **测试场景**：cores=0 降级 / cores>0 正常 / cpuInfo 为空 / memInfo 为空

### 2. `MonitorService.GetProcessList()` — 排序逻辑

- **文件**：`internal/service/monitor.go:67`
- **分支**：`switch sortBy { case "pid": ... case "name": ... case "memory": ... default: ... }`
- **当前不可测原因**：依赖 `process.Pids()` 真实 OS 进程表
- **重构方向**：抽 `sortProcesses(items []model.ProcessItem, sortBy, order string)`，测试里传入构造好的 slice
- **测试场景**：6 种排序组合（3 个字段 × 2 个方向）+ 空列表

### 3. `DashboardService.GetMetrics()` — 远程/本地切换

- **文件**：`internal/service/dashboard.go:24`
- **分支**：`if d.remoteCollector != nil { ... } else { ... }`
- **当前不可测原因**：远程采集器要连真 Agent，本地采集要调 gopsutil
- **重构方向**：把 `RemoteCollector` 抽象为接口 `MetricsProvider`
- **测试场景**：Agent 在线 / Agent 离线降级 / Agent 返回错误

### 4. `Config.Load()` — AgentHosts 空格处理

- **文件**：`internal/config/config.go:55`
- **分支**：`if hosts != "" { cfg.AgentHosts = strings.Split(hosts, ",") }`
- **当前状态**：已测（`config_test.go`），但 Split 保留空格——未来可能加 TrimSpace
- **重构方向**：如果加 Trim，同步改测试（已在测试注释里标注）

---

## 不需要重构的（确认放弃）

| 代码 | 原因 |
|------|------|
| `Collect()` — gopsutil 的 cpu.Percent / mem.VirtualMemory / disk.Usage | 只有错误处理分支，且错误意味着采集失败——不值得为罕见路径写 mock |
| `History.Query()` | 已测（时间过滤、淘汰逻辑） |
| `History.StartCollector` | goroutine + 定时器，集成测试已覆盖 |
| Handler 层 — 读参数 → 调 service → 写 JSON | 纯胶水代码，没有分支 |

---

## 何时回头做

- 项目新增第二个 Service 层功能时
- 有人改 `monitor.go` / `dashboard.go` 引入 bug 后
- 项目交给第二个人维护时

---

*最后更新：2026-07-30。基于告警引擎和 Agent 化完成后首次测试覆盖评审的记录。*
