import { test, expect } from '@playwright/test';

test.describe('用户管理页面加载', () => {
  test('不应无限刷新 /api/users', async ({ page }) => {
    // 拦截统计请求次数
    let apiCalls = 0;
    page.on('request', (req) => {
      if (req.url().includes('/api/users') && req.method() === 'GET') {
        apiCalls++;
      }
    });

    // 登录
    await page.goto('http://localhost:5173/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'admin123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/');

    // 进入用户管理页
    await page.click('text=用户管理');
    await page.waitForURL('**/users');

    // 等 3 秒，让可能的死循环暴露
    await page.waitForTimeout(3000);

    console.log('/api/users GET 请求次数（3 秒内）:', apiCalls);

    // 页面稳定后只应有 1 次（首次加载）；死循环会是几十次
    expect(apiCalls).toBeLessThanOrEqual(3);

    // 表格应显示用户数据而不是一直转圈
    await page.waitForSelector('table', { timeout: 5000 });
    await page.waitForSelector('tbody tr', { timeout: 5000 });
    const rowCount = await page.locator('tbody tr').count();
    console.log('表格行数:', rowCount);
    expect(rowCount).toBeGreaterThanOrEqual(1);
  });
});
