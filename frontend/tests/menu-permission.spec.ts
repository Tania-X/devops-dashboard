import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';

// 浏览器通道由 playwright.config.js 自动探测（msedge→chrome→chromium），此处不硬编码

// viewer 标准只读权限(不含 webhook,保证菜单数量确定)
const VIEWER_STD_PERMS = [
  'dashboard:view', 'server:read', 'log:read', 'deployment:read', 'monitor:read',
];

async function login(page: Page, username: string, password: string) {
  await page.goto('http://localhost:5173/login');
  await page.fill('#username', username);
  await page.fill('#password', password);
  await page.click('button[type="submit"]');
  // 等跳转完成 + 菜单渲染(不用 networkidle,页面有轮询等不到)
  await page.waitForURL('**/');
  await page.locator('.ant-menu-item').first().waitFor({ timeout: 10000 });
}

// 通过 admin 把 viewer 角色重置为标准只读权限,保证菜单数量断言稳定(其他测试可能改动角色权限)
async function resetViewerToStd() {
  const res = await fetch('http://localhost:8080/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'admin123' }),
  });
  const { token } = await res.json();
  await fetch('http://localhost:8080/api/settings/roles/viewer', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ permissions: VIEWER_STD_PERMS }),
  });
}

test.describe('菜单级权限控制', () => {
  test('viewer(test/test) 看不到用户管理和 Agent 管理', async ({ page }) => {
    await resetViewerToStd(); // 重置角色为标准只读,避免其他测试改动影响

    await login(page, 'test', 'test');

    // 等待侧边栏菜单渲染完成(登录后跳转需要时间,直接读取可能拿到空列表)
    await page.locator('.ant-menu-item').first().waitFor({ timeout: 10000 });
    const visible = await page.locator('.ant-menu-item').allTextContents();
    console.log('viewer 可见菜单:', JSON.stringify(visible));

    // 无权限的菜单项不应出现(用户管理=user:read, Agent 管理=agent:read, 系统设置=webhook:read)
    expect(visible.join()).not.toContain('用户管理');
    expect(visible.join()).not.toContain('Agent 管理');
    expect(visible.join()).not.toContain('系统设置');
    // 只读菜单应可见
    expect(visible.join()).toContain('系统概览');
    expect(visible.join()).toContain('服务器管理');
    // viewer 应恰好看到 5 个只读菜单
    expect(visible.length).toBe(5);
  });

  test('admin 可见全部 8 个菜单', async ({ page }) => {
    await login(page, 'admin', 'admin123');

    // 等待侧边栏菜单渲染完成
    await page.locator('.ant-menu-item').first().waitFor({ timeout: 10000 });
    const visible = await page.locator('.ant-menu-item').allTextContents();
    console.log('admin 可见菜单:', JSON.stringify(visible));

    expect(visible.join()).toContain('用户管理');
    expect(visible.join()).toContain('Agent 管理');
    expect(visible.join()).toContain('系统设置');
    expect(visible.length).toBe(8);
  });
});
