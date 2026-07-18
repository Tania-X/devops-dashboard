# DevOps Dashboard UI 视觉规范 v1.0

> 借鉴 Grafana 设计哲学，适配企业内网运维场景。  
> 规范日期：2026-07-18  
> 适用范围：一期所有页面（Dashboard、Server、Log、Deployment）

---

## 1. 设计原则

1. **深色优先**：运维人员长时间盯屏，深色主题减少眼部疲劳。
2. **信息密度高**：一屏尽可能展示更多有效信息，减少滚动。
3. **状态显性化**：正常/警告/严重 三种状态必须一眼可辨。
4. **一致性**：同类型元素（卡片、表格、标签）全系统统一。

---

## 2. 色彩规范（Color Palette）

### 2.1 背景色
| Token | 色值 | 用途 |
|-------|------|------|
| `--bg-page` | `#141414` | 页面最底层背景 |
| `--bg-panel` | `#1f1f1f` | Panel 卡片背景 |
| `--bg-panel-hover` | `#2a2a2a` | 卡片悬停/激活态 |
| `--bg-header` | `#111217` | 顶部 Header 背景 |
| `--bg-sidebar` | `#000000` | 侧边栏背景 |

### 2.2 文本色
| Token | 色值 | 用途 |
|-------|------|------|
| `--text-primary` | `#ffffff` | 主标题、重要数值 |
| `--text-secondary` | `#aaaaaa` | 次要文本、描述 |
| `--text-muted` | `#666666` | 禁用态、占位符 |

### 2.3 强调色（品牌色）
| Token | 色值 | 用途 |
|-------|------|------|
| `--accent-primary` | `#177ddc` | 主按钮、链接、激活菜单 |
| `--accent-primary-hover` | `#3c9ae8` | 按钮悬停态 |

### 2.4 状态色（Status Colors）—— 与 Grafana 对齐
| 状态 | 色值 | 用途 |
|------|------|------|
| **正常 Normal** | `#73bf69` | 运行中、成功、低负载 |
| **警告 Warning** | `#f2c94c` | 负载偏高、WARN 日志 |
| **严重 Critical** | `#e02f44` | 宕机、ERROR 日志、高危告警 |
| **信息 Info** | `#3274d9` | 提示、INFO 日志、部署中 |
| **未知 Unknown** | `#aaaaaa` | 无数据、未知状态 |

---

## 3. 布局规范（Layout）

### 3.1 整体结构
```
┌──────────────────────────────────────────────┐  ← Header (56px)
│ Logo        面包屑                    👤 Admin │
├──────────┬───────────────────────────────────┤
│          │                                   │
│ Sidebar  │         Content Area              │
│ (200px)  │         (自适应剩余宽度)            │
│          │                                   │
│          │                                   │
└──────────┴───────────────────────────────────┘
```

### 3.2 尺寸规范
| 元素 | 尺寸 | 说明 |
|------|------|------|
| Header 高度 | `56px` | 固定顶部 |
| Sidebar 宽度 | `200px` | 固定左侧，不可折叠（一期简化）|
| Content 内边距 | `24px` | 四边统一 |
| Panel 圆角 | `4px` | 与 Grafana 一致，微微圆角 |
| Panel 内边距 | `16px` | 卡片内部留白 |
| Panel 间距 | `16px` | 卡片之间间隙 |

### 3.3 Dashboard 栅格（参考 Grafana 12-column）
- 一屏最大容纳 3 列 Panel（每个约 4 columns）
- 趋势图等宽 Panel 占满 12 columns（整行）
- 响应式断点：一期只做 ≥1366px 桌面端

---

## 4. Panel（面板）规范

Panel 是 Dashboard 的最小展示单元，类比 Grafana 的 Panel。

### 4.1 Stat Panel（大数字卡片）
```
┌─────────────────────┐
│ 图标  标题          │  ← 14px, #aaaaaa
│                     │
│  45.2%              │  ← 32px, #ffffff, 加粗
│  ↑ 5.3% vs 上小时    │  ← 12px, 绿色↑/红色↓
└─────────────────────┘
```
- 宽度：自适应（一排放 4 个）
- 高度：`120px`
- 数值字体：`Roboto Mono` 或等宽字体（数字对齐美观）

### 4.2 Chart Panel（图表卡片）
```
┌──────────────────────────────────────┐
│  📈 CPU & 内存趋势        [?] [⋮]   │  ← 标题栏 16px
├──────────────────────────────────────┤
│                                      │
│         [  折线图区域  ]              │
│                                      │
└──────────────────────────────────────┘
```
- 标题栏底部带 `1px` 分隔线，颜色 `#333333`
- 图表区域高度：`300px`
-  Tooltip 背景：`rgba(0,0,0,0.8)`，边框 `#333`

### 4.3 Table Panel（表格卡片）
- 表头背景：`#1f1f1f`（与 Panel 同底）
- 表头文字：`#aaaaaa`，`12px`，大写
- 行悬停背景：`#2a2a2a`
- 行高：`48px`
- 分页器：右下角，简约风格

---

## 5. 字体规范（Typography）

| 层级 | 字号 | 字重 | 用途 |
|------|------|------|------|
| H1 页面标题 | `20px` | 600 | Dashboard / Server 等页面大标题 |
| H2 Panel 标题 | `16px` | 500 | 卡片标题 |
| Body 正文 | `14px` | 400 | 表格内容、描述文字 |
| Caption 辅助 | `12px` | 400 | 时间戳、次要信息 |
| Number 数值 | `32px` | 700 | Stat 大数字 |

**字体栈**：`"Inter", "PingFang SC", "Microsoft YaHei", sans-serif`  
**等宽数字**：`"Roboto Mono", monospace`（用于 IP、版本号、数值）

---

## 6. 组件规范

### 6.1 状态标签（Status Tag）
| 状态 | 背景色 | 文字色 | 圆角 |
|------|--------|--------|------|
| 运行中 / 成功 | `rgba(115, 191, 105, 0.2)` | `#73bf69` | `4px` |
| 警告 / 部署中 | `rgba(242, 201, 76, 0.2)` | `#f2c94c` | `4px` |
| 失败 / 严重 | `rgba(224, 47, 68, 0.2)` | `#e02f44` | `4px` |
| 已停机 / 待部署 | `rgba(170, 170, 170, 0.2)` | `#aaaaaa` | `4px` |

### 6.2 侧边栏菜单
- 默认态：文字 `#aaaaaa`，背景透明
- 悬停态：文字 `#ffffff`，背景 `#177ddc`，左侧 `3px` 蓝色竖线指示器
- 激活态：与悬停态一致，保持高亮

### 6.3 按钮
| 类型 | 背景 | 文字 | 边框 |
|------|------|------|------|
| Primary | `#177ddc` | `#ffffff` | 无 |
| Default | 透明 | `#aaaaaa` | `1px solid #555` |
| Danger | `#e02f44` | `#ffffff` | 无 |

---

## 7. 动画与交互

| 场景 | 动画效果 | 时长 |
|------|---------|------|
| Panel 悬停 | 轻微上浮 `translateY(-2px)` + 阴影加深 | `200ms` |
| 数字变化 | 数字滚动动画（可选） | `500ms` |
| 页面切换 | 淡入 `opacity 0→1` | `150ms` |
| 表格行悬停 | 背景色渐变过渡 | `150ms` |
| 加载中 | 骨架屏（Skeleton）或 Spin 旋转 | — |

---

## 8. 暗黑模式适配说明

一期**固定深色主题**，不做浅色模式切换。所有颜色 Token 均按深色场景定义。  
若二期增加浅色模式，需重新定义一套 Light Token，通过 CSS 变量切换。

---

## 9. 与 Ant Design 的映射

本项目使用 Ant Design 作为组件库，需要将上述规范映射到 Ant Design 的 `ConfigProvider` 主题配置：

```typescript
// 核心映射关系（后续代码实现时参考）
const theme = {
  token: {
    colorBgBase: '#141414',        // --bg-page
    colorBgContainer: '#1f1f1f',   // --bg-panel
    colorTextBase: '#ffffff',      // --text-primary
    colorPrimary: '#177ddc',       // --accent-primary
    colorSuccess: '#73bf69',       // 正常态
    colorWarning: '#f2c94c',       // 警告态
    colorError: '#e02f44',         // 严重态
    borderRadius: 4,               // Panel 圆角
    fontSize: 14,                  // Body 字号
  },
  algorithm: theme.darkAlgorithm,  // 启用 Ant Design 暗色算法
};
```

---

## 附录：Grafana 对比速查

| 元素 | Grafana 实现 | 我们的实现 |
|------|-------------|-----------|
| 主题 | Dark / Light / Custom | 固定 Dark（一期）|
| Panel | ReactGridLayout 拖拽 | 固定栅格布局（一期）|
| 图表 | uPlot / Flot | ECharts |
| 数据源 | Plugin 架构 | 抽象接口（二期）|
| 告警线 | 图表阈值配置 | 前端硬编码阈值（一期）|
| 变量模板 | `$var` 替换 | 无（一期）|