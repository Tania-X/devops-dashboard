/**
 * Dashboard 页面配置规范
 *
 * 设计思想：Dashboard 页面的结构不由代码硬编码，而由本配置文件驱动。
 * 新增/删除/修改卡片，只需改此文件，无需改动 React 组件代码。
 *
 * 这是 SDD 和 Grafana "Dashboard as Code" 理念的轻量级实践。
 */

// ============================================
// 类型定义（规范层）
// ============================================

/** 支持的 Panel 类型 */
export type PanelType = 'stat' | 'chart' | 'alert-list' | 'quick-links';

/** 数据 API 标识，对应 v1-api.yaml 中的 operationId */
export type ApiEndpoint =
  | 'getDashboardMetrics'
  | 'getDashboardTrend'
  | 'getDashboardAlerts';

/** 数值单位 */
export type UnitType = 'percent' | 'bytes' | 'count' | 'duration' | 'none';

/** 布局列宽（24 栅格系统） */
export type ColSpan = 6 | 8 | 12 | 16 | 24;

/** Stat 卡片配置 */
export interface StatPanelConfig {
  type: 'stat';
  title: string;
  /** 对应 API 返回数据中的字段路径，如 "cpu.current" */
  dataKey: string;
  /** 关联的 API */
  api: ApiEndpoint;
  unit: UnitType;
  /** 阈值配置：超过则变色 */
  thresholds?: {
    warning: number;
    critical: number;
  };
  /** 图标名称（使用 Ant Design Icon 的标识） */
  icon: string;
}

/** 图表 Panel 配置 */
export interface ChartPanelConfig {
  type: 'chart';
  title: string;
  api: ApiEndpoint;
  /** 图表子类型 */
  chartType: 'line' | 'area' | 'bar';
  /** 数据序列配置 */
  series: Array<{
    name: string;
    dataKey: string;
    color: string;
  }>;
  /** 高度（像素） */
  height: number;
}

/** 告警列表 Panel 配置 */
export interface AlertListPanelConfig {
  type: 'alert-list';
  title: string;
  api: ApiEndpoint;
  /** 展示条数 */
  limit: number;
}

/** 快捷入口 Panel 配置 */
export interface QuickLinksPanelConfig {
  type: 'quick-links';
  title: string;
  /** 跳转链接配置 */
  links: Array<{
    label: string;
    path: string;
    icon: string;
  }>;
}

/** Panel 配置联合类型 */
export type PanelConfig =
  | StatPanelConfig
  | ChartPanelConfig
  | AlertListPanelConfig
  | QuickLinksPanelConfig;

/** Dashboard 行配置，一行可包含多个 Panel */
export interface DashboardRow {
  /** 行高（像素），可选，默认 auto */
  height?: number;
  /** 该行包含的 Panels */
  panels: Array<{
    /** 栅格宽度 */
    colSpan: ColSpan;
    /** Panel 配置 */
    config: PanelConfig;
  }>;
}

/** 完整的 Dashboard 配置 */
export interface DashboardConfig {
  version: string;
  title: string;
  description: string;
  rows: DashboardRow[];
}

// ============================================
// 一期 Dashboard 实际配置（实现层）
// ============================================

export const dashboardConfig: DashboardConfig = {
  version: '1.0.0',
  title: '系统概览',
  description: '运维监控仪表盘首页，展示核心指标、趋势和告警',
  rows: [
    // ---------- 第 1 行：4 个核心指标卡片 ----------
    {
      panels: [
        {
          colSpan: 6,
          config: {
            type: 'stat',
            title: 'CPU 使用率',
            dataKey: 'cpu.current',
            api: 'getDashboardMetrics',
            unit: 'percent',
            thresholds: { warning: 60, critical: 80 },
            icon: 'DashboardOutlined',
          },
        },
        {
          colSpan: 6,
          config: {
            type: 'stat',
            title: '内存使用率',
            dataKey: 'memory.current',
            api: 'getDashboardMetrics',
            unit: 'percent',
            thresholds: { warning: 70, critical: 85 },
            icon: 'DatabaseOutlined',
          },
        },
        {
          colSpan: 6,
          config: {
            type: 'stat',
            title: '磁盘使用率',
            dataKey: 'disk.current',
            api: 'getDashboardMetrics',
            unit: 'percent',
            thresholds: { warning: 75, critical: 90 },
            icon: 'HddOutlined',
          },
        },
        {
          colSpan: 6,
          config: {
            type: 'stat',
            title: '活跃告警',
            dataKey: 'alertCount',
            api: 'getDashboardMetrics',
            unit: 'count',
            thresholds: { warning: 1, critical: 5 },
            icon: 'AlertOutlined',
          },
        },
      ],
    },

    // ---------- 第 2 行：趋势图（占满整行） ----------
    {
      panels: [
        {
          colSpan: 24,
          config: {
            type: 'chart',
            title: 'CPU & 内存趋势（最近 6 小时）',
            api: 'getDashboardTrend',
            chartType: 'line',
            series: [
              { name: 'CPU 使用率', dataKey: 'cpuData', color: '#73bf69' },
              { name: '内存使用率', dataKey: 'memoryData', color: '#3274d9' },
            ],
            height: 300,
          },
        },
      ],
    },

    // ---------- 第 3 行：告警列表 + 快捷入口 ----------
    {
      panels: [
        {
          colSpan: 16,
          config: {
            type: 'alert-list',
            title: '最近告警',
            api: 'getDashboardAlerts',
            limit: 5,
          },
        },
        {
          colSpan: 8,
          config: {
            type: 'quick-links',
            title: '快捷入口',
            links: [
              { label: '服务器管理', path: '/servers', icon: 'CloudServerOutlined' },
              { label: '日志查询', path: '/logs', icon: 'FileTextOutlined' },
              { label: '部署状态', path: '/deployments', icon: 'RocketOutlined' },
            ],
          },
        },
      ],
    },
  ],
};

// ============================================
// 辅助函数（规范校验）
// ============================================

/** 校验一行中的 colSpan 之和是否等于 12 */
export function validateRow(row: DashboardRow): boolean {
  const totalSpan = row.panels.reduce((sum, p) => sum + p.colSpan, 0);
  return totalSpan === 12;
}

/** 校验整个 Dashboard 配置 */
export function validateDashboard(config: DashboardConfig): string[] {
  const errors: string[] = [];

  config.rows.forEach((row, index) => {
    if (!validateRow(row)) {
      errors.push(`第 ${index + 1} 行栅格之和不等于 12`);
    }
  });

  return errors;
}

// 运行时报错（开发期检查）
const errors = validateDashboard(dashboardConfig);
if (errors.length > 0) {
  console.error('[DashboardConfig] 配置校验失败:', errors);
}

export default dashboardConfig;
