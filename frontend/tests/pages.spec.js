import { test, expect } from '@playwright/test';

const BASE = 'http://localhost:5173';

test.describe('页面截图', () => {

    test('Dashboard 概况', async ({ page }) => {
        await page.goto(BASE);
        await page.waitForSelector('.ant-card', { timeout: 10000 });
        await page.screenshot({ path: 'screenshots/01-dashboard.png', fullPage: true });
    });

    test('服务器管理', async ({ page }) => {
        await page.goto(`${BASE}/servers`);
        await page.waitForSelector('.ant-table', { timeout: 10000 });
        await page.screenshot({ path: 'screenshots/02-servers.png', fullPage: true });
    });

    test('日志查询', async ({ page }) => {
        await page.goto(`${BASE}/logs`);
        await page.waitForSelector('.ant-table', { timeout: 10000 });
        await page.screenshot({ path: 'screenshots/03-logs.png', fullPage: true });
    });

    test('部署状态', async ({ page }) => {
        await page.goto(`${BASE}/deployments`);
        await page.waitForSelector('.ant-table', { timeout: 10000 });
        await page.screenshot({ path: 'screenshots/04-deployments.png', fullPage: true });
    });

    test('实时监控', async ({ page }) => {
        await page.goto(`${BASE}/monitor`);
        await page.waitForSelector('.ant-table-row', { timeout: 15000 });
        await page.screenshot({ path: 'screenshots/05-monitor.png', fullPage: true });
    });
});

test.describe('基础断言', () => {

    test('Dashboard 有 Ant Design Card 组件', async ({ page }) => {
        await page.goto(BASE);
        await expect(page.locator('.ant-card')).not.toHaveCount(0);
    });

    test('服务器表格有数据', async ({ page }) => {
        await page.goto(`${BASE}/servers`);
        const rows = page.locator('.ant-table-row');
        await expect(rows).not.toHaveCount(0);
    });

    test('日志表格有数据', async ({ page }) => {
        await page.goto(`${BASE}/logs`);
        const rows = page.locator('.ant-table-row');
        await expect(rows).not.toHaveCount(0);
    });
});

test.describe('API 直连检查', () => {

    test('Dashboard metrics API 正常', async ({ request }) => {
        const resp = await request.get('http://localhost:8080/api/dashboard/metrics');
        await expect(resp).toBeOK();
        const body = await resp.json();
        expect(body.cpu.current).toBeGreaterThan(0);
    });

    test('服务器列表 API 正常', async ({ request }) => {
        const resp = await request.get('http://localhost:8080/api/servers?page=1&pageSize=5');
        await expect(resp).toBeOK();
    });

    test('日志 API 正常', async ({ request }) => {
        const resp = await request.get('http://localhost:8080/api/logs');
        await expect(resp).toBeOK();
    });
});

test.describe('进程列表分页一致性', () => {

    test('翻页 + 切换条数 + 页码 三栏自洽', async ({ page }) => {
        await page.goto(`${BASE}/monitor`);
        await page.waitForSelector('.ant-table-row', { timeout: 15000 });

        // ── 1. 获取基准：当前总数 ──
        const totalText = await page.locator('.ant-pagination-total-text').textContent();
        console.log('分页信息:', totalText);
        // 格式："共 200 个进程" → 提取 200
        const totalMatch = totalText?.match(/(\d+)/);
        const total = totalMatch ? parseInt(totalMatch[1]) : 0;

        // ── 2. "最多显示" 的当前值 ──
        // ProcessList 页面有 "最多显示：" <Select>，默认值 200
        const limitSelect = page.locator('.ant-select', { hasText: /1\d{2}|2\d{2}|5\d{2}/ }).first();
        const currentLimit = await limitSelect.textContent();
        console.log('当前最多显示:', currentLimit);

        // ── 3. "xx/页" 的当前值 ──
        // 表格下方分页栏的 pageSize selector
        const pageSizeSelect = page.locator('.ant-pagination .ant-select');
        const currentPageSize = await pageSizeSelect.textContent();
        console.log('当前每页条数:', currentPageSize);

        // ── 4. 自洽性校验 ──
        // 如果"最多显示 200"，则表格实际只有 200 条
        const limit = parseInt(currentLimit || '200');
        const displayedTotal = limit < total ? limit : total;
        console.log(`总 ${total}，限制 ${limit}，应显示 ${displayedTotal} 条`);

        // ── 5. 切到第 2 页，行内容应变 ──
        // 先数一下当前总行数
        const totalRows = await page.locator('.ant-table-row').count();
        console.log(`当前显示行数: ${totalRows}`);

        // 第 1 列是行号（#），第 2 列才是 PID
        const page1FirstPid = await page.locator('.ant-table-row').first().locator('td').nth(1).textContent();
        console.log(`第 1 页第一个 PID: ${page1FirstPid}`);

        // 点第 2 页
        const page2Btn = page.locator('.ant-pagination-item-2');
        const page2Exists = await page2Btn.count();
        console.log(`第 2 页按钮存在: ${page2Exists > 0}`);
        await page2Btn.click();
        await page.waitForTimeout(1000);

        const page2FirstPid = await page.locator('.ant-table-row').first().locator('td').nth(1).textContent();
        console.log(`第 2 页第一个 PID: ${page2FirstPid}`);

        if (page1FirstPid === page2FirstPid) {
            console.log('⚠️ 翻到第 2 页后 PID 没变，翻页可能未生效');
        } else {
            console.log(`✅ 翻页正常: P1=${page1FirstPid} → P2=${page2FirstPid}`);
        }

        // ── 6. 回到第 1 页，改"最多显示"为 100，总数应变 ──
        await page.locator('.ant-pagination-item-1').click();
        await page.waitForTimeout(300);

        // 点"最多显示"下拉，选 100
        const limitDropdowns = page.locator('.ant-select', { hasText: /1\d{2}|2\d{2}|5\d{2}/ });
        await limitDropdowns.first().click();
        await page.locator('.ant-select-item-option', { hasText: '100' }).click();
        await page.waitForTimeout(500);

        const newTotalText = await page.locator('.ant-pagination-total-text').textContent();
        console.log('切到 100 后分页信息:', newTotalText);

        // ── 7. 检查状态栏（第 6 列）是否全是空 ──
        const statusCells = page.locator('.ant-table-row td:nth-child(6)');
        const count = await statusCells.count();
        let emptyCount = 0;
        for (let i = 0; i < count; i++) {
            const text = (await statusCells.nth(i).textContent())?.trim();
            if (!text || text === '---') emptyCount++;
        }
        console.log(`状态栏: ${emptyCount}/${count} 行为空或 ---`);
        if (emptyCount === count) {
            console.log('⚠️ 全部状态栏为空！');
        }

        // ── 8. 验证自动刷新重置手动修改的问题 ──
        await limitDropdowns.first().click();
        await page.locator('.ant-select-item-option', { hasText: '100' }).click();
        await page.waitForTimeout(1000);
        const afterChange = await page.locator('.ant-pagination-total-text').textContent();
        console.log('手动改为 100 后:', afterChange);

        await page.waitForTimeout(15000); // 超过 10s 自动刷新
        const afterRefresh = await page.locator('.ant-pagination-total-text').textContent();
        console.log('15 秒后（自动刷新触发后）:', afterRefresh);

        if (afterChange !== afterRefresh) {
            console.log('⚠️ 自动刷新重置了手动修改！', afterChange, '→', afterRefresh);
        } else {
            console.log('✅ 自动刷新没有重置手动修改');
        }
    });
});
