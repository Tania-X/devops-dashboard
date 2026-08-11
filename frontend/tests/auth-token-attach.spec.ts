import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';

test.use({ channel: 'chrome' });

async function login(page: Page) {
  await page.goto('http://localhost:5173/login');
  await page.fill('#username', 'admin');
  await page.fill('#password', 'admin123');
  await page.click('button[type="submit"]');
  await page.waitForURL('**/');
  await page.waitForLoadState('networkidle');
}

// 回归测试:登录后受保护接口(agents/settings/webhook)不应返回 401
// 背景:client.ts 曾缺少 Authorization 附加逻辑,导致所有需认证接口 401
test('受保护接口自动携带 token,不应出现 4xx/5xx', async ({ page }) => {
  const errors: string[] = [];
  const responses: string[] = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error') errors.push(`[console] ${msg.text()}`);
  });
  page.on('pageerror', (err) => errors.push(`[pageerror] ${err.message}`));
  page.on('response', (res) => {
    if (res.status() >= 400) responses.push(`HTTP ${res.status()} ${res.url()}`);
  });

  await login(page);

  // Agent 管理页
  await page.locator('.ant-menu-item').filter({ hasText: 'Agent 管理' }).click();
  await page.waitForURL('**/agents');
  await page.waitForTimeout(2000);

  // 系统设置页
  await page.locator('.ant-menu-item').filter({ hasText: '系统设置' }).click();
  await page.waitForURL('**/settings');
  await page.waitForTimeout(2000);

  // 受保护接口不应出现 4xx/5xx(此前因缺 token 全部 401)
  const badApiResponses = responses.filter((r) => r.includes('/api/'));
  expect(badApiResponses).toEqual([]);
  expect(errors.filter((e) => !e.includes('Warning'))).toEqual([]);
});
