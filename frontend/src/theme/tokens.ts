/**
 * Design Tokens — 设计令牌
 *
 * 全项目唯一的颜色/尺寸/字体来源，与 spec/ui-theme.md 一一对应。
 * 规则：组件中禁止写死色值（#1f1f1f 之类），一律从这里取。
 * 修改视觉风格只需改本文件。
 */

export const colors = {
  /** 背景色（spec 2.1） */
  bg: {
    /** 页面最底层背景 --bg-page */
    page: '#141414',
    /** Panel 卡片背景 --bg-panel */
    panel: '#1f1f1f',
    /** 卡片悬停/激活态 --bg-panel-hover */
    panelHover: '#2a2a2a',
    /** 顶部 Header 背景 --bg-header */
    header: '#111217',
    /** 侧边栏背景 --bg-sidebar */
    sidebar: '#000000',
  },

  /** 文本色（spec 2.2） */
  text: {
    /** 主标题、重要数值 --text-primary */
    primary: '#ffffff',
    /** 次要文本、描述 --text-secondary */
    secondary: '#aaaaaa',
    /** 禁用态、占位符 --text-muted */
    muted: '#666666',
  },

  /** 强调色/品牌色（spec 2.3） */
  brand: {
    /** 主按钮、链接、激活菜单 --accent-primary */
    primary: '#177ddc',
    /** 按钮悬停态 --accent-primary-hover */
    hover: '#3c9ae8',
    /** 品牌渐变终点色（AppLogo / 登录页氛围） */
    gradientEnd: '#3fb6d9',
  },

  /** 角色色（RBAC 角色标签，spec 无定义，沿用既有视觉语义） */
  role: {
    /** 管理员 —— 最高权限，紫色（与原 roleColorMap admin 一致） */
    admin: '#534ab7',
  },

  /** 状态色（spec 2.4，与 Grafana 对齐） */
  status: {
    /** 正常：运行中、成功、低负载 */
    success: '#73bf69',
    /** 警告：负载偏高、WARN 日志 */
    warning: '#f2c94c',
    /** 严重：宕机、ERROR 日志、高危告警 */
    critical: '#e02f44',
    /** 信息：提示、INFO 日志、部署中 */
    info: '#3274d9',
    /** 未知：无数据、未知状态 */
    unknown: '#aaaaaa',
  },

  /** 边框、分隔线 */
  border: '#333333',
  /** 弱边框（按钮描边等） */
  borderLight: '#555555',

  /** 代码块背景（命令行、环境变量等） */
  codeBg: '#1a1a1a',
  /** 代码块文字 */
  codeText: '#cccccc',
} as const;

/** 语义状态 → 文字色/背景色（spec 6.1 状态标签） */
export const semanticStatus = {
  success: { color: colors.status.success, bg: 'rgba(115, 191, 105, 0.2)' },
  warning: { color: colors.status.warning, bg: 'rgba(242, 201, 76, 0.2)' },
  critical: { color: colors.status.critical, bg: 'rgba(224, 47, 68, 0.2)' },
  info: { color: colors.status.info, bg: 'rgba(50, 116, 217, 0.2)' },
  unknown: { color: colors.status.unknown, bg: 'rgba(170, 170, 170, 0.2)' },
} as const;

/** 尺寸规范（spec 3.2） */
export const spacing = {
  /** Content 内边距 */
  page: 24,
  /** Panel 内边距 */
  panel: 16,
  /** Panel 间距 */
  gap: 16,
} as const;

/** 圆角（spec 3.2，与 Grafana 一致） */
export const radius = {
  panel: 4,
} as const;

/** 字体规范（spec 5） */
export const fonts = {
  /** 主字体栈 */
  family: '"Inter", "PingFang SC", "Microsoft YaHei", sans-serif',
  /** 等宽数字字体（IP、版本号、数值） */
  mono: '"Roboto Mono", "Consolas", "Courier New", monospace',
  size: {
    /** H1 页面标题 */
    h1: 20,
    /** H2 Panel 标题 */
    h2: 16,
    /** Body 正文 */
    body: 14,
    /** Caption 辅助 */
    caption: 12,
    /** Stat 大数字 */
    number: 32,
  },
  weight: {
    h1: 600,
    h2: 500,
    number: 700,
  },
} as const;

/** 布局尺寸（spec 3.2） */
export const layout = {
  headerHeight: 56,
  sidebarWidth: 200,
} as const;
