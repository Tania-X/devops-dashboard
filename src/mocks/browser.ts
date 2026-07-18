import { setupWorker } from 'msw/browser';
import { http, HttpResponse, delay } from 'msw';
import { faker } from '@faker-js/faker';
import { getDevOpsDashboardAPIMock } from '../api/client';
import type { ServerItem, ServerDetail, PagedResultServerItem } from '../api/model';

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
  customServerListHandler,
  customServerDetailHandler,
  ...getDevOpsDashboardAPIMock(),
);
