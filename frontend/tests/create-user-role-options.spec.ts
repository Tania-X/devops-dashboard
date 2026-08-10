import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';

test.use({ channel: 'chrome' });

async function login(page: Page, username: string, password: string) {
  await page.goto('http://localhost:5173/login');
  await page.fill('#username', username);
  await page.fill('#password', password);
  await page.click('button[type="submit"]');
  await page.waitForURL('**/');
  await page.waitForLoadState('networkidle');
}

test.describe('新增用户角色选项', () => {
  test('角色下拉应包含 admin/viewer/operator 三种角色', async ({ page }) => {
    await login(page, 'admin', 'admin123');
    await page.locator('.ant-menu-item').filter({ hasText: '用户管理' }).click();
    await page.waitForURL('**/users');
    await page.waitForSelector('table');

    // 打开新增用户弹窗
    await page.click('button:has-text("新增用户")');
    const modal = page.locator('.ant-modal');
    await modal.waitFor({ state: 'visible' });

    // 打开角色下拉
    await modal.locator('.ant-select').click();
    // 等待下拉渲染(antd 有动画)
    await page.waitForSelector('.ant-select-dropdown:visible .ant-select-item-option', { timeout: 5000 });
    const options = await page
      .locator('.ant-select-dropdown:visible .ant-select-item-option')
      .allTextContents();
    console.log('角色下拉选项:', JSON.stringify(options));

    // 期望:admin / viewer / operator 三种角色都在(当前缺 operator)
    expect(options.length).toBe(3);
    expect(options.join()).toContain('管理员');
    expect(options.join()).toContain('观察者');
    expect(options.join()).toContain('运维'); // operator 角色
  });
});
