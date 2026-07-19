import { setupWorker } from 'msw/browser';
import { http, HttpResponse, delay } from 'msw';
import { faker } from '@faker-js/faker';
import { getDevOpsDashboardAPIMock } from '../api/client';
import type { ServerItem, ServerDetail, PagedResultServerItem, LogItem, PagedResultLogItem, DeploymentItem, DeploymentHistoryItem, DashboardMetrics, DashboardTrend, AlertItem } from '../api/model';
import { MetricValueStatus } from '../api/model';

// ============================================
// Log Mock 数据池（支持级别筛选 + 关键词搜索 + 分页）
// ============================================

const SERVICE_LIST = [
  'api-gateway', 'user-service', 'order-service', 'payment-service',
  'notification-service', 'auth-service', 'log-collector', 'monitor-agent',
];

const LOG_CONTENT_TEMPLATES: Record<string, string[]> = {
  INFO: [
    'Request processed successfully',
    'User login: user_{id}',
    'Cache refreshed for key: {key}',
    'Health check passed',
    'Scheduled task completed: {task}',
    'Config reloaded from remote',
    'Backup completed: {file}',
  ],
  WARN: [
    'High memory usage detected: {percent}%',
    'Slow query detected: {time}ms',
    'Retry attempt {n}/3 failed',
    'Connection pool approaching limit: {count}/{max}',
    'Disk usage above 80% on {mount}',
    'Rate limit threshold reached: {percent}%',
  ],
  ERROR: [
    'Database connection timeout after {time}ms',
    'NullPointerException in {service}',
    'Payment gateway returned status 500',
    'Disk full: {mount}',
    'Authentication failed for user {user}',
    'Message queue consumer crashed: {reason}',
    'Failed to connect to upstream: {host}',
  ],
};

function generateLogContent(level: string): string {
  const templates = LOG_CONTENT_TEMPLATES[level] || LOG_CONTENT_TEMPLATES.INFO;
  let content = faker.helpers.arrayElement(templates);
  content = content.replace('{id}', faker.string.alphanumeric(8));
  content = content.replace('{key}', faker.word.sample());
  content = content.replace('{task}', faker.word.words(2));
  content = content.replace('{file}', faker.system.fileName());
  content = content.replace('{percent}', String(faker.number.int({ min: 80, max: 99 })));
  content = content.replace('{time}', String(faker.number.int({ min: 100, max: 5000 })));
  content = content.replace('{n}', String(faker.number.int({ min: 1, max: 3 })));
  content = content.replace('{count}', String(faker.number.int({ min: 8, max: 10 })));
  content = content.replace('{max}', '10');
  content = content.replace('{mount}', faker.helpers.arrayElement(['/', '/data', '/var/log']));
  content = content.replace('{service}', faker.helpers.arrayElement(SERVICE_LIST));
  content = content.replace('{user}', faker.internet.username());
  content = content.replace('{reason}', faker.word.sample());
  content = content.replace('{host}', faker.internet.domainName());
  return content;
}

function generateLogList(count: number): LogItem[] {
  const LEVELS = ['INFO', 'WARN', 'ERROR'] as const;
  return Array.from({ length: count }, (_, i) => {
    const level = faker.helpers.arrayElement(LEVELS);
    const service = faker.helpers.arrayElement(SERVICE_LIST);
    const date = faker.date.recent({ days: 7 });
    const hostIndex = faker.number.int({ min: 0, max: 34 });
    return {
      id: `log-${String(i + 1).padStart(5, '0')}`,
      time: date.toISOString().replace('T', ' ').slice(0, 19),
      level: level as any,
      service,
      content: generateLogContent(level),
      sourceHost: `srv-${String(hostIndex + 1).padStart(3, '0')}`,
      logPath: `/var/log/${service}/app.log`,
      traceId: faker.string.alphanumeric(16),
    };
  });
}

/** 固定日志数据池，按时间降序排列 */
const ALL_LOGS = generateLogList(80).sort((a, b) => b.time.localeCompare(a.time));

/** 自定义 Log List Handler — 支持级别/服务名/时间范围筛选 + 关键词搜索 + 分页 */
const customLogListHandler = http.get('*/api/logs', async ({ request }) => {
  await delay(300);

  const url = new URL(request.url);
  const level = url.searchParams.get('level');
  const service = url.searchParams.get('service');
  const keyword = url.searchParams.get('keyword');
  const startTime = url.searchParams.get('startTime');
  const endTime = url.searchParams.get('endTime');
  const page = Math.max(1, parseInt(url.searchParams.get('page') || '1', 10));
  const pageSize = Math.max(1, parseInt(url.searchParams.get('pageSize') || '10', 10));

  let filtered = [...ALL_LOGS];
  if (level) {
    filtered = filtered.filter((l) => l.level === level);
  }
  if (service) {
    filtered = filtered.filter((l) => l.service === service);
  }
  if (startTime) {
    filtered = filtered.filter((l) => l.time >= startTime);
  }
  if (endTime) {
    filtered = filtered.filter((l) => l.time <= endTime);
  }
  if (keyword) {
    const lowerKeyword = keyword.toLowerCase();
    filtered = filtered.filter(
      (l) =>
        l.content.toLowerCase().includes(lowerKeyword) ||
        l.service.toLowerCase().includes(lowerKeyword)
    );
  }

  const total = filtered.length;
  const start = (page - 1) * pageSize;
  const list = filtered.slice(start, start + pageSize);

  return HttpResponse.json({
    list,
    total,
    page,
    pageSize,
  } as PagedResultLogItem);
});

// ============================================
// Server List Mock 数据池（支持筛选 + 分页）
// ============================================

const OS_LIST = ['CentOS 7.9', 'Ubuntu 22.04', 'Debian 12', 'Rocky Linux 9', 'AlmaLinux 8'];
const STATUS_LIST = ['running', 'stopped', 'maintenance'] as const;

function generateServerList(count: number): ServerItem[] {
  return Array.from({ length: count }, (_, i) => ({
    id: `srv-${String(i + 1).padStart(3, '0')}`,
    hostname: `${faker.word.sample()}-${faker.word.sample()}-${String(i + 1).padStart(2, '0')}`,
    ip: faker.internet.ip(),
    os: faker.helpers.arrayElement(OS_LIST),
    cpuCores: faker.helpers.arrayElement([4, 8, 16, 32, 64]),
    memoryGb: faker.helpers.arrayElement([8, 16, 32, 64, 128]),
    status: faker.helpers.arrayElement(STATUS_LIST),
    uptime: `${faker.number.int({ min: 1, max: 365 })}d ${faker.number.int({ min: 0, max: 23 })}h ${faker.number.int({ min: 0, max: 59 })}m`,
  }));
}

/** 固定数据池，页面刷新前数据一致 */
const ALL_SERVERS = generateServerList(35);

/** 磁盘分区模板 */
function generateDiskPartitions(): ServerDetail['diskPartitions'] {
  const mounts = ['/', '/data', '/var/log', '/backup'];
  return mounts.map((mount) => {
    const totalGb = faker.helpers.arrayElement([100, 250, 500, 1000, 2000]);
    const usedGb = Math.round(totalGb * faker.number.float({ min: 0.2, max: 0.95 }));
    return { mount, totalGb, usedGb };
  });
}

/** 网卡模板 */
function generateNetworkInterfaces(): ServerDetail['networkInterfaces'] {
  const names = ['eth0', 'eth1', 'lo'];
  return names.map((name) => ({
    name,
    ip: name === 'lo' ? '127.0.0.1' : faker.internet.ip(),
    mac: faker.internet.mac(),
  }));
}

/** 自定义 Server List Handler — 支持筛选 + 分页 */
const customServerListHandler = http.get('*/api/servers', async ({ request }) => {
  await delay(300);

  const url = new URL(request.url);
  const status = url.searchParams.get('status');
  const page = Math.max(1, parseInt(url.searchParams.get('page') || '1', 10));
  const pageSize = Math.max(1, parseInt(url.searchParams.get('pageSize') || '10', 10));

  let filtered = [...ALL_SERVERS];
  if (status) {
    filtered = filtered.filter((s) => s.status === status);
  }

  const total = filtered.length;
  const start = (page - 1) * pageSize;
  const list = filtered.slice(start, start + pageSize);

  return HttpResponse.json({
    list,
    total,
    page,
    pageSize,
  } as PagedResultServerItem);
});

/** 自定义 Server Detail Handler — 根据 ID 返回对应数据 + 磁盘/网卡 */
const customServerDetailHandler = http.get('*/api/servers/:id', async ({ params }) => {
  await delay(200);

  const { id } = params;
  const base = ALL_SERVERS.find((s) => s.id === id);

  if (!base) {
    return new HttpResponse(null, { status: 404 });
  }

  const detail: ServerDetail = {
    ...base,
    diskPartitions: generateDiskPartitions(),
    networkInterfaces: generateNetworkInterfaces(),
  };

  return HttpResponse.json(detail);
});

// ============================================
// Deployment Mock 数据池
// ============================================

const APP_NAMES = [
  'api-gateway', 'user-service', 'order-service', 'payment-service',
  'notification-service', 'auth-service', 'log-collector', 'monitor-agent',
  'config-server', 'cache-proxy', 'search-engine', 'report-generator',
  'data-sync', 'file-storage', 'web-frontend',
];

const ENV_LIST = ['dev', 'test', 'prod'] as const;

function generateDeploymentHistory(count: number): DeploymentHistoryItem[] {
  const histories: DeploymentHistoryItem[] = [];
  const baseDate = new Date();
  for (let i = 0; i < count; i++) {
    const major = 2;
    const minor = faker.number.int({ min: 0, max: 4 });
    const patch = faker.number.int({ min: 0, max: 9 });
    const version = `v${major}.${minor}.${patch}`;
    const status = faker.helpers.arrayElement(['success', 'failed'] as const);
    const date = new Date(baseDate.getTime() - i * faker.number.int({ min: 3600000, max: 86400000 * 3 }));
    histories.push({
      version,
      operator: faker.person.fullName(),
      durationSec: faker.number.int({ min: 30, max: 600 }),
      status: status as any,
      deployedAt: date.toISOString(),
    });
  }
  return histories;
}

const DEPLOYMENT_HISTORY_MAP: Record<string, DeploymentHistoryItem[]> = {};

const ALL_DEPLOYMENTS: DeploymentItem[] = APP_NAMES.map((appName, i) => {
  const env = faker.helpers.arrayElement(ENV_LIST);
  const historyCount = faker.number.int({ min: 3, max: 8 });
  const histories = generateDeploymentHistory(historyCount);
  const id = `app-${String(i + 1).padStart(3, '0')}`;
  DEPLOYMENT_HISTORY_MAP[id] = histories;

  const latest = histories[0];

  // 大部分与最新历史状态一致，少量 pending/deploying
  const rand = Math.random();
  let status: string = latest.status;
  if (rand > 0.90) status = 'deploying';
  else if (rand > 0.80) status = 'pending';

  return {
    id,
    appName,
    version: latest.version,
    env: env as any,
    status: status as any,
    lastDeployedAt: latest.deployedAt,
  };
});

/** 自定义 Deployment List Handler */
const customDeploymentListHandler = http.get('*/api/deployments', async () => {
  await delay(300);
  return HttpResponse.json(ALL_DEPLOYMENTS);
});

/** 自定义 Deployment History Handler */
const customDeploymentHistoryHandler = http.get('*/api/deployments/:id/history', async ({ params }) => {
  await delay(200);
  const { id } = params;
  const history = DEPLOYMENT_HISTORY_MAP[id as string] || [];
  return HttpResponse.json(history);
});

// ============================================
// Dashboard Mock 数据池
// ============================================

const DASHBOARD_METRICS: DashboardMetrics = {
  cpu: { current: 67.5, status: MetricValueStatus.warning },
  memory: { current: 54.2, status: MetricValueStatus.normal },
  disk: { current: 78.0, status: MetricValueStatus.warning },
  alertCount: 3,
};

function generateTrendData(): DashboardTrend {
  const now = new Date();
  const timeLabels: string[] = [];
  const cpuData: number[] = [];
  const memoryData: number[] = [];

  for (let i = 6; i >= 0; i--) {
    const d = new Date(now.getTime() - i * 3600000);
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    const hour = String(d.getHours()).padStart(2, '0');
    const minute = String(d.getMinutes()).padStart(2, '0');
    timeLabels.push(`${month}-${day} ${hour}:${minute}`);

    // 模拟有波动的趋势数据，基础值 + 随机波动
    const baseCpu = 40 + Math.sin(i * 0.8) * 15;
    cpuData.push(Math.round(Math.max(10, Math.min(95, baseCpu + faker.number.int({ min: -8, max: 10 }))) * 10) / 10);

    const baseMem = 50 + Math.cos(i * 0.6) * 12;
    memoryData.push(Math.round(Math.max(15, Math.min(90, baseMem + faker.number.int({ min: -6, max: 8 }))) * 10) / 10);
  }

  return { timeLabels, cpuData, memoryData };
}

const DASHBOARD_TREND = generateTrendData();

const DASHBOARD_ALERTS: AlertItem[] = [
  {
    id: 'alert-001',
    level: 'critical' as any,
    message: '服务器 srv-012 磁盘使用率超过 90%，当前 93%',
    source: 'srv-012 (192.168.1.45)',
    time: new Date(Date.now() - 15 * 60000).toISOString(),
  },
  {
    id: 'alert-002',
    level: 'warning' as any,
    message: 'api-gateway 服务响应时间 P99 超过 500ms',
    source: 'api-gateway',
    time: new Date(Date.now() - 42 * 60000).toISOString(),
  },
  {
    id: 'alert-003',
    level: 'warning' as any,
    message: '支付服务 payment-service 内存使用率持续上升',
    source: 'payment-service',
    time: new Date(Date.now() - 78 * 60000).toISOString(),
  },
  {
    id: 'alert-004',
    level: 'info' as any,
    message: '每日备份任务已完成，耗时 4m32s',
    source: 'backup-agent',
    time: new Date(Date.now() - 120 * 60000).toISOString(),
  },
  {
    id: 'alert-005',
    level: 'critical' as any,
    message: '数据库主从延迟超过 5 秒，当前 7.3s',
    source: 'db-master (192.168.1.10)',
    time: new Date(Date.now() - 180 * 60000).toISOString(),
  },
];

/** 自定义 Dashboard Metrics Handler */
const customDashboardMetricsHandler = http.get('*/api/dashboard/metrics', async () => {
  console.log('[MSW] intercepted GET /api/dashboard/metrics');
  await delay(200);
  return HttpResponse.json(DASHBOARD_METRICS);
});

/** 自定义 Dashboard Trend Handler */
const customDashboardTrendHandler = http.get('*/api/dashboard/trend', async () => {
  console.log('[MSW] intercepted GET /api/dashboard/trend');
  await delay(300);
  return HttpResponse.json(DASHBOARD_TREND);
});

/** 自定义 Dashboard Alerts Handler */
const customDashboardAlertsHandler = http.get('*/api/dashboard/alerts', async ({ request }) => {
  console.log('[MSW] intercepted GET /api/dashboard/alerts');
  await delay(200);
  const url = new URL(request.url);
  const limit = parseInt(url.searchParams.get('limit') || '5', 10);
  return HttpResponse.json(DASHBOARD_ALERTS.slice(0, limit));
});

// ============================================
// 启动 Worker — 自定义 handler 放前面优先匹配
// ============================================

export const worker = setupWorker(
  customDashboardMetricsHandler,
  customDashboardTrendHandler,
  customDashboardAlertsHandler,
  customLogListHandler,
  customServerListHandler,
  customServerDetailHandler,
  customDeploymentListHandler,
  customDeploymentHistoryHandler,
  ...getDevOpsDashboardAPIMock(),
);
