import { setupWorker } from 'msw/browser';
import { http, HttpResponse, delay } from 'msw';
import { faker } from '@faker-js/faker';
import { getDevOpsDashboardAPIMock } from '../api/client';
import type { ServerItem, ServerDetail, PagedResultServerItem, LogItem, PagedResultLogItem } from '../api/model';

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
// 启动 Worker — 自定义 handler 放前面优先匹配
// ============================================

export const worker = setupWorker(
  customLogListHandler,
  customServerListHandler,
  customServerDetailHandler,
  ...getDevOpsDashboardAPIMock(),
);
