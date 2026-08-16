# DevOps Dashboard — 开发指南

> 涵盖本地开发、前端模式、经验教训与未来规划。
> 最后更新: 2026-07-25

---

## 一、本地启动

### 前置条件

- **Go** >= 1.24
- **Node.js** >= 20
- **pnpm** 或 npm

### 启动后端

```bash
cd backend
go mod tidy
go run cmd/api/main.go

# 启动日志：
# 服务启动 address=http://localhost:8080
# 验证：curl http://localhost:8080/api/health
```

### 启动前端

```bash
cd frontend
npm install
npm run dev        # 默认通过 Vite proxy 连接后端
```

浏览器访问 `http://localhost:5173`。

### 前端 Mock 模式（不依赖后端）

```bash
cd frontend
VITE_USE_MSW=true npm run dev
```

---

## 二、前端开发模式

### 2.1 数据获取

当前各页面使用 `useState + useEffect` 直接调 API。未来计划引入 TanStack Query 统一管理服务端状态（见 §四）。

**现有 Hook 模式（后续将被 Query 替代）：**

```typescript
const [loading, setLoading] = useState(true);
const [data, setData] = useState<T>([]);

useEffect(() => {
  setLoading(true);
  api.getServerList(params).then(res => {
    setData(res.list);
  }).catch(console.error).finally(() => setLoading(false));
}, [params]);
```

### 2.2 Mock 策略

MSW (Mock Service Worker) 可通过环境变量控制是否启用。

**固定数据池原则**：Mock 数据使用固定池而非完全随机，保证筛选、分页、详情联查可复现：

```typescript
// mocks/browser.ts — 注册时自定义 handler 放前面，覆盖 Orval 生成
export const worker = setupWorker(
  customServerListHandler,    // ← 先匹配，拦截特定接口
  customServerDetailHandler,
  ...getDevOpsDashboardAPIMock(),  // 兜底其他接口
);
```

### 2.3 筛选与分页协同

筛选条件变化时强制重置页码为 1，避免数据不足时出现空白页：

```typescript
const handleStatusChange = (value: string | undefined) => {
  setStatusFilter(value);
  fetchList(1, pagination.pageSize, value);  // page = 1
};
```

### 2.4 图表与 DOM 操作

ECharts 直接操作 DOM，与 React 共存时需遵循：

1. **图表容器必须始终挂载** — 不能放在条件渲染分支里，否则 ref 绑定会失效
2. **Loading 遮罩和图表容器应为兄弟节点** — 禁止共享同一个父节点，否则 React 虚拟 DOM diff 会与 ECharts 插入的 `<canvas>` 冲突

```tsx
{/* ✅ 正确模式 */}
<div style={{ position: 'relative' }}>
  <div ref={chartRef} />                             {/* echarts 独占 */}
  {loading && <Spin style={{ position: 'absolute', inset: 0 }} />}  {/* 兄弟节点 */}
</div>
```

### 2.5 主题与样式

UI 视觉规范详见 `spec/ui-theme.md`。Ant Design 通过 ConfigProvider 配置暗色主题：

```typescript
// 核心映射
const theme = {
  token: {
    colorBgBase: '#141414',       // 页面背景
    colorBgContainer: '#1f1f1f',  // 卡片背景
    colorTextBase: '#ffffff',     // 主文本
    colorPrimary: '#177ddc',      // 强调色
    colorSuccess: '#73bf69',      // 正常态
    colorWarning: '#f2c94c',      // 警告态
    colorError: '#e02f44',        // 严重态
  },
  algorithm: theme.darkAlgorithm,
};
```

**自定义状态标签色值（避免使用纯饱和色）：**

| 状态 | 色值 | 语义 |
|--|--|--|
| running / success | `#73bf69` | 健康绿，比默认 green 更柔和 |
| stopped | `#aaaaaa` | 中性灰 |
| maintenance | `#f2c94c` | 警示黄 |
| critical / error | `#e02f44` | 严重红 |

---

## 三、常见问题排查

### 3.1 编译/运行

**Go 未使用的 import**

```text
internal/api/server.go:6:2: "time" imported and not used
```

Go 禁止未使用的 import。VS Code 自动导入后若删除了相关代码，需手动移除。

**前端 502 Bad Gateway**

Vite proxy 转发到后端但后端未启动。确认：`curl http://localhost:8080/api/health`。

**MSW 新增 Handler 未生效**

浏览器可能缓存了旧的 Service Worker。Chrome DevTools → Application → Service Workers → Unregister，然后刷新。

### 3.2 栅格布局

Ant Design 5.x 的 Row/Col 是 **24 列栅格**，不是 Bootstrap 的 12 列。配置 `colSpan` 时确保每行之和为 24：

| 内容 | colSpan 配置 |
|--|--|
| 4 个 Stat 卡片 | `6 + 6 + 6 + 6 = 24` |
| 全宽趋势图 | `24` |
| 告警列表 + 快捷入口 | `16 + 8 = 24` |

### 3.3 后端 API

新增接口或修改接口时，对照 `spec/v1-api.yaml` 逐项检查：**参数 → 筛选 → 排序 → 返回格式**。

---

## 四、未来路线图

以下方向按优先级排列，每个可独立开展。

### Phase 3：前端工程化

| 任务 | 说明 |
|--|--|
| TS strict: true | 开启严格模式，逐文件修复类型错误，消除 `any` |
| Prettier | 统一代码格式 |
| Git Hooks | Husky + lint-staged，提交前自动检查 |
| Vite 路径别名 | `@/` → `./src` |
| TanStack Query | 替换手写 `useState + useEffect` 数据获取模式，统一缓存/错误/重试 |
| 通用组件 | PageCard、StatusBadge 等消除跨页面重复样式 |
| Mock 拆分 | 将 414 行 god file 按职责拆分为 data/handlers/utils |
| CI | GitHub Actions，PR 自动 lint + typecheck + build |
| Vitest 测试 | 组件测试 + Hook 测试，覆盖率 ≥ 60% |
| 路由懒加载 | React.lazy + vendor 拆包（echarts / antd 分离） |

### Phase 4：告警体系

- `monitor.Collect()` 中加阈值判断，超阈值写入 `alert` 表
- `DashboardService.GetAlerts()` 从表查，替换 Handler 内 mock 数据
- 目前告警是 Handler 里硬编码的 5 条假数据

### Phase 5：前端监控页面

- `/api/monitor/processes`、`/api/monitor/host` 后端已实现，前端缺页面
- 可新增进程管理页（表格展示 + 排序/筛选/搜索 + 详情抽屉）
- 可新增主机信息页（OS/CPU/内存/启动时间展示）

---

## 四·五、已知安全取舍(登记)

> AI review 要求登记的安全取舍,记录原因与迁移计划。

| 取舍 | 风险 | 原因 | 迁移计划 |
|--|--|--|--|
| SSH `InsecureIgnoreHostKey`(跳过主机密钥校验) | MITM 可截获密码/篡改部署 | 个人学习项目 + 远程主机动态变化,known_hosts 维护成本高;仅用于部署自研 agent 的可信内网场景 | 部署到生产前改为 `ssh.KnownHosts` 或固定指纹校验(infra/ssh.go) |

---

## 五、环境速查

```bash
# Go 代理（国内）
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off

# 启动后端
cd backend && go run cmd/api/main.go

# 启动前端（联调模式）
cd frontend && npm run dev

# 启动前端（Mock 模式）
cd frontend && VITE_USE_MSW=true npm run dev

# 编译后端
cd backend && go build -o devops-api cmd/api/main.go

# 交叉编译
GOOS=linux GOARCH=amd64 go build -o devops-api-linux cmd/api/main.go
GOOS=darwin GOARCH=arm64 go build -o devops-api-mac cmd/api/main.go
```

详见 [环境搭建 (macOS)](env-setup-macos.md)。
