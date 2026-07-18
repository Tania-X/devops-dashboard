# DevOps Dashboard 项目问题记录

> 记录从环境搭建到项目构建全过程中遇到的问题及根因分析，供后续参考和避坑。

---

## 一、环境搭建

### 1.1 npm / node 命令不可用

**现象**：Git Bash 中执行 `npm` 或 `node` 提示 `command not found`。

**根因**：fnm 安装的 Node 未自动写入 Shell 的 `PATH`，Git Bash 不会继承 Windows 用户级 PATH 的变更，除非重新打开终端。

**解决**：将 Node 安装路径追加到用户级 PATH，关闭并重新打开终端。

```powershell
[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", "User") + ";D:\tools\fnm\node-versions\node-versions\v20.20.2\installation",
    "User"
)
```

---

### 1.2 C 盘空间紧张，Node.js 需安装到 D 盘

**现象**：默认 Node 安装到 C 盘，空间不足。

**根因**：fnm 默认版本目录在 `%LOCALAPPDATA%\fnm`（C 盘用户目录）。

**解决**：设置 `FNM_DIR` 环境变量指向 D 盘父目录。

```powershell
[Environment]::SetEnvironmentVariable("FNM_DIR", "D:\tools\fnm", "User")
```

---

### 1.3 fnm 出现双重 `node-versions` 嵌套路径

**现象**：Node 实际路径为 `D:/tools/fnm/node-versions/node-versions/v20.20.2/installation`。

**根因**：`FNM_DIR` 被误设为 `D:\tools\fnm\node-versions`，fnm 内部会再创建一层 `node-versions`。

**解决**：将 `FNM_DIR` 改为 `D:\tools\fnm`，清理重装。

```powershell
[Environment]::SetEnvironmentVariable("FNM_DIR", "D:\tools\fnm", "User")
fnm uninstall 20
# 手动删除 D:\tools\fnm\node-versions
fnm install 20
fnm default 20
```

正确结构：

```
D:/tools/fnm/
├── fnm.exe
└── node-versions/
    └── v20.20.2/
        └── installation/
```

---

## 二、项目初始化

### 2.1 `create-vite` 因目录非空而取消

**现象**：在项目根目录（已有 `spec/`、`docs/` 目录）执行 `npm create vite@latest .` 时，工具提示目录非空并拒绝初始化。

**根因**：`create-vite` 默认要求目标目录为空，防止覆盖现有文件。

**解决**：在 `/tmp/vite-init` 空目录中初始化，再将生成的文件（`package.json`、`vite.config.ts`、`tsconfig.json`、`index.html`、`src/` 等）手动复制回项目根目录，保留原有的 `spec/` 和 `docs/`。

---

## 三、工具链与依赖

### 3.1 Orval mutator 配置错误

**现象**：执行 `npx orval` 报错：`Your mutator file doesn't have the customInstance exported function`。

**根因**：`orval.config.ts` 中配置了 `mutator` 指向一个不存在的自定义 axios 实例文件。

**解决**：移除 `mutator` 配置块，让 Orval 使用默认的 axios 实例生成逻辑，或手动创建 `customInstance` 导出。

---

### 3.2 `@faker-js/faker` 缺失

**现象**：Orval 生成的 MSW mock 代码中使用了 `faker`，运行时报错模块找不到。

**根因**：Orval 的 mock 生成功能依赖 `@faker-js/faker`，但项目初始化时未安装。

**解决**：

```bash
npm install -D @faker-js/faker
```

---

## 四、类型与编译错误

### 4.1 `@ant-design/icons` 类型断言不兼容

**现象**：`(icons as Record<string, React.ComponentType>)` 编译失败。

**根因**：Ant Design Icons 的模块导出类型与 `Record<string, React.ComponentType>` 不完全匹配，TS 严格模式拒绝此断言。

**解决**：使用双重断言绕过类型检查：

```typescript
const IconComponent = (icons as unknown as Record<string, React.ComponentType<any>>)[config.icon];
```

---

### 4.2 ECharts `area` 不是有效的 series 类型

**现象**：配置 `chartType: 'area'` 时，ECharts 报错 `area` 不是有效的 series type。

**根因**：ECharts 中面积图是 `line` 类型配合 `areaStyle` 配置，没有独立的 `area` type。

**解决**：渲染时做类型映射：

```typescript
type: (config.chartType === 'area' ? 'line' : config.chartType) as 'line' | 'bar'
areaStyle: config.chartType === 'area' ? { opacity: 0.1, color: s.color } : undefined
```

---

### 4.3 React UMD global 错误

**现象**：运行时报错 `React is not defined`，尽管代码中使用了 JSX。

**根因**：文件中没有显式 `import React from 'react'`，在旧版 JSX transform 或某些 TS 配置下会隐式使用全局 `React`，但该全局不存在。

**解决**：在用到 `React.createElement` 的文件顶部显式导入：

```typescript
import React from 'react';
```

> 注：Vite + React 19 默认使用新的 JSX Transform（无需导入 React），但代码中显式调用了 `React.createElement`（动态创建图标组件），因此仍需导入。

---

### 4.4 `ServerOutlined` 图标不存在

**现象**：`ServerOutlined` 无法从 `@ant-design/icons` 导入。

**根因**：该图标在当前安装的 `@ant-design/icons` 版本中不存在（可能被重命名或移除）。

**解决**：改用 `CloudServerOutlined`。

---

## 五、布局与 UI 问题

### 5.1 Dashboard 页面内容只填充浏览器宽度的一半

**现象**：Dashboard 三行内容（4 个指标卡片、趋势图、告警列表+快捷入口）都只占 Content 区域的一半宽度，右侧留有大片空白。

**根因**：**对 Ant Design 栅格系统的列数理解错误**。Ant Design 5.x 的 `Row/Col` 采用 **24 列栅格系统**，但 `dashboard-config.ts` 中的 `colSpan` 值是按 **12 列系统** 设计的：

| 行 | 面板 | 配置 `colSpan` | 24 列实际占比 |
|---|---|---|---|
| 第 1 行 | 4 个 Stat 卡片 | `3` × 4 = 12 | 12/24 = **50%** |
| 第 2 行 | 趋势图 | `12` | 12/24 = **50%** |
| 第 3 行 | 告警列表 + 快捷入口 | `8 + 4` = 12 | 12/24 = **50%** |

每行栅格之和都只有 12，在 24 列系统中自然只占一半宽度。

**解决**：将所有 `colSpan` 值翻倍，适配 24 栅格：

- `3` → `6`（4 个 Stat 卡片：4×6=24 ✓）
- `4` → `8`
- `6` → `12`
- `8` → `16`（告警列表：16，快捷入口：8，16+8=24 ✓）
- `12` → `24`（趋势图：24 ✓）

同时更新类型定义：

```typescript
export type ColSpan = 6 | 8 | 12 | 16 | 24;
```

> **经验教训**：在使用 UI 框架的栅格系统前，务必确认其总列数。Ant Design 4.x/5.x 是 24 列，Bootstrap 是 12 列，不同框架不可混用配置习惯。

---

## 六、布局框架宽度溢出（连带修复）

**现象**：修复上述宽度问题后，页面出现横向滚动条，或内容区域宽度计算异常。

**根因**：`AppSidebar` 使用了 `position: 'fixed'` 脱离文档流，但 Ant Design 的 `Layout` 组件仍会检测到 `Sider` 子元素并启用 `flex-direction: row`。此时 Sider 不参与 flex 空间分配（不占宽度），右侧 `Layout` 却获得 `flex: auto`（占满 100vw），再加上显式设置的 `marginLeft: 200`，导致整体宽度 = 200px + 100vw，超出视口。

**解决**：放弃 `position: fixed`，改用 **固定视口高度 + 内容区滚动** 的标准 Dashboard 布局模式：

1. **AppLayout.tsx**：最外层 `Layout` 设置 `height: '100vh', overflow: 'hidden'`，右侧 `Layout` 同样限制高度，`Content` 设置 `overflow: 'auto'`（仅内容区可滚动）
2. **AppSidebar.tsx**：移除 `position: 'fixed'`、`left`、`top`、`bottom`，让 Sider 正常参与 flex 布局
3. **AppHeader.tsx**：移除 `position: 'sticky'`、`top`、`zIndex`，加上 `flexShrink: 0`

效果：
- 页面总高度严格锁定 `100vh`，无横向滚动
- Sidebar + Header 始终固定，仅 Content 区域纵向滚动
- Sider 正常参与 flex 空间分配，右侧宽度自动为 `100vw - 200px`

> **经验教训**：在 flex 布局中，避免将子元素设为 `position: fixed` 却仍期望父容器正确计算剩余空间。固定侧边栏的正确做法是限制父容器高度并让内容区滚动，而不是让侧边栏脱离文档流。

---

## 七、Dashboard 图表渲染问题链

### 7.1 ChartPanel 一直转圈，API 请求未发出

**现象**：Dashboard 趋势图区域持续显示 Spin 加载动画，控制台 Network 面板看不到 `/api/dashboard/trend` 请求。

**根因**：ChartPanel 使用条件渲染控制 `ref` 绑定的 DOM 节点：

```tsx
{loading ? (
  <div> <Spin /> </div>          // ← 这个 div 没有 ref
) : (
  <div ref={chartRef} />         // ← loading 为 true 时不渲染
)}
```

`useEffect` 一进入就判断 `if (!chartRef.current) return;`，导致后续的 `api.getDashboardTrend()` **一行都没执行**。

**解决**：让 `ref={chartRef}` 的 div **始终存在于 DOM** 中，loading 状态改用绝对定位的遮罩层覆盖：

```tsx
<div style={{ position: 'relative' }}>
  <div ref={chartRef} />           {/* 始终渲染 */}
  {loading && <div style={{ position: 'absolute', inset: 0 }}> <Spin /> </div>}
</div>
```

> **经验教训**：React 中需要操作 DOM（如 echarts.init）的节点，**不能放在条件渲染分支里**。`ref` 必须绑定到始终挂载的元素上，否则 `useEffect` + `ref` 的组合会失效。

---

### 7.2 MSW Service Worker 缓存导致新增 Handler 未生效

**现象**：注销 Service Worker 并重新注册后，`/api/dashboard/metrics` 和 `/api/dashboard/alerts` 能被拦截，但刷新前一直报 `Failed to fetch`。

**根因**：MSW v2 使用 Service Worker 拦截请求，浏览器会缓存旧的 Service Worker 脚本。新增 `customDashboardTrendHandler` 后，旧的 SW 仍在运行，不认识新的 handler，请求被 passthrough 到网络（`return fetch(requestClone)`），而本地没有真实后端，直接报错。

**解决**：Chrome DevTools → Application → Service Workers → 找到 `mockServiceWorker.js` → 点击 **Unregister**，然后刷新页面。

> **经验教训**：在 MSW 开发中新增/修改 handler 后，如果请求行为异常，优先检查 Service Worker 是否已更新。普通 F5 刷新不一定能触发 SW 更新，手动注销是最可靠的方式。

---

### 7.3 echarts 与 React 虚拟 DOM 冲突（`removeChild` 报错）

**现象**：折线图一闪而过，然后整个图表区域空白，控制台报错：

```text
NotFoundError: Failed to execute 'removeChild' on 'Node':
The node to be removed is not a child of this node.
```

**根因**：echarts 调用 `echarts.init(chartRef.current)` 时，会在该 DOM 节点内部插入 `<canvas>` 等元素。React 不知道这些元素的存在。当 `loading` 从 `true` 变为 `false` 时，React 尝试移除遮罩层的 `div`，发现真实 DOM 结构和虚拟 DOM 不一致（多了一个 echarts 插入的 canvas 容器），`removeChild` 失败。

**解决**：**echarts 容器 和 loading 遮罩层 必须是兄弟节点**，不能嵌套在同一个 React 管理的父节点内：

```tsx
{/* 错误：echarts 和 Spin 在同一个 div 内 */}
<div ref={chartRef}>             {/* echarts 会往这里插 canvas */}
  {loading && <div> <Spin /> </div>}
</div>

{/* 正确：兄弟节点，互不干扰 */}
<div style={{ position: 'relative' }}>
  <div ref={chartRef} />         {/* echarts 独占，React 不管内部 */}
  {loading && <div> <Spin /> </div>}  {/* React 只管理这个节点 */}
</div>
```

> **经验教训**：任何直接操作 DOM 的库（echarts、D3、Chart.js 等）与 React 共存时，**不能让第三方库和 React 条件渲染共享同一个父节点**。给第三方库一个独立的容器，React 条件渲染放在兄弟节点上，是避免 DOM diff 冲突的标准做法。

---

### 7.4 栅格校验硬编码了 12 列

**现象**：控制台报错 `[DashboardConfig] 配置校验失败: 第 X 行栅格之和不等于 12`。

**根因**：`validateRow` 函数写死了 `=== 12`，但 Ant Design 5.x 的 `Row/Col` 是 **24 列栅格**。

**解决**：校验逻辑改为 `=== 24`。

> **经验教训**：见 §5.1，再次印证了使用 UI 框架前必须确认其栅格列数。Ant Design 是 24 列，不要按 Bootstrap 的 12 列习惯配置。

---

### 7.5 图例与时间轴重叠

**现象**：echarts 默认图例（"CPU 使用率 / 内存使用率"）在顶部居中，与 xAxis 的时间标签区域发生视觉重叠。

**根因**：echarts 默认 `legend` 位置为顶部居中，未给图表内容留出足够间距。

**解决**：显式设置图例定位到左上角：

```typescript
legend: {
  data: config.series.map((s) => s.name),
  textStyle: { color: '#aaaaaa' },
  top: 0,
  left: 0,
}
```

---

### 7.6 时间标签缺少日期

**现象**：xAxis 时间标签只显示 `14:00`，无法区分是哪一天的 14:00。

**根因**：Mock 数据生成时只取了 `HH:mm`。

**解决**：时间标签格式改为 `MM-DD HH:mm`：

```typescript
const month = String(d.getMonth() + 1).padStart(2, '0');
const day = String(d.getDate()).padStart(2, '0');
const hour = String(d.getHours()).padStart(2, '0');
const minute = String(d.getMinutes()).padStart(2, '0');
timeLabels.push(`${month}-${day} ${hour}:${minute}`);
```

---

*文档维护：随项目迭代持续更新。*
