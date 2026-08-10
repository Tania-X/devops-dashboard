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

test.describe('菜单级权限控制', () => {
  test('viewer(test/test) 看不到用户管理和 Agent 管理', async ({ page }) => {
    await login(page, 'test', 'test');

    const visible = await page.locator('.ant-menu-item').allTextContents();
    console.log('viewer 可见菜单:', JSON.stringify(visible));

    // 无权限的菜单项不应出现(用户管理=user:read, Agent 管理=agent:read)
    expect(visible.join()).not.toContain('用户管理');
    expect(visible.join()).not.toContain('Agent 管理');
    // 只读菜单应可见
    expect(visible.join()).toContain('系统概览');
    expect(visible.join()).toContain('服务器管理');
    // viewer 应恰好看到 5 个只读菜单
    expect(visible.length).toBe(5);
  });

  test('admin 可见全部 7 个菜单', async ({ page }) => {
    await login(page, 'admin', 'admin123');

    const visible = await page.locator('.ant-menu-item').allTextContents();
    console.log('admin 可见菜单:', JSON.stringify(visible));

    expect(visible.join()).toContain('用户管理');
    expect(visible.join()).toContain('Agent 管理');
    expect(visible.length).toBe(7);
  });
});
