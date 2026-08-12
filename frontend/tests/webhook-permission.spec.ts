import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';

// 浏览器通道由 playwright.config.js 自动探测（msedge→chrome→chromium），此处不硬编码
// Webhook 设置页权限组合矩阵测试:
//   权限点:webhook:read(查看) / webhook:update(配置+测试推送)
//   组合空间 2²=4,全部覆盖:
//     ① 无 webhook 权限      → 菜单不可见
//     ② 仅 read              → 菜单可见,表单只读,无操作按钮
//     ③ 仅 update(无 read)   → 菜单不可见(依赖 read 才可达)
//     ④ read + update        → 菜单可见,可编辑,有保存/测试按钮
// 依赖关系:read 是入口(菜单显隐),update 依赖 read 才有意义

const BASE_PERMS = [
  'dashboard:view', 'server:read', 'log:read', 'deployment:read', 'monitor:read',
];

async function login(page: Page, username: string, password: string) {
  await page.goto('http://localhost:5173/login');
  await page.fill('#username', username);
  await page.fill('#password', password);
  await page.click('button[type="submit"]');
  await page.waitForURL('**/');
  await page.locator('.ant-menu-item').first().waitFor({ timeout: 15000 });
}

async function readFormState(page: Page) {
  return page.evaluate(() => {
    const qa = (sel: string) => Array.from(document.querySelectorAll(sel));
    const isDisabled = (el: Element | null) => el ? (el as HTMLInputElement).disabled ?? el.hasAttribute('disabled') : null;
    return {
      urlInputDisabled: isDisabled(qa('input[placeholder*="qyapi"]')[0] ?? null),
      saveBtnCount: qa('button').filter((b) => b.textContent?.includes('保存')).length,
      testBtnCount: qa('button').filter((b) => b.textContent?.includes('测试推送')).length,
    };
  });
}

// 通过 admin 设置 viewer 角色权限
async function setViewerPerms(perms: string[]) {
  const res = await fetch('http://localhost:8080/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'admin123' }),
  });
  const { token } = await res.json();
  await fetch('http://localhost:8080/api/settings/roles/viewer', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ permissions: perms }),
  });
}

test.describe('Webhook 设置页权限组合矩阵', () => {
  test('① 无 webhook 权限 — 菜单不可见', async ({ page }) => {
    await setViewerPerms(BASE_PERMS);
    await login(page, 'test', 'test');

    const menuItems = await page.locator('.ant-menu-item').allTextContents();
    console.log('无 webhook 权限菜单:', JSON.stringify(menuItems));
    expect(menuItems.join()).not.toContain('系统设置');
  });

  test('② 仅 webhook:read — 菜单可见,表单只读,无按钮', async ({ page }) => {
    await setViewerPerms([...BASE_PERMS, 'webhook:read']);
    await login(page, 'test', 'test');
    await page.locator('.ant-menu-item').filter({ hasText: '系统设置' }).click();
    await page.waitForURL('**/settings');
    await page.locator('form').first().waitFor({ timeout: 15000 });

    const state = await readFormState(page);
    console.log('仅 read:', JSON.stringify(state));

    expect(state.urlInputDisabled).toBe(true);
    expect(state.saveBtnCount).toBe(0);
    expect(state.testBtnCount).toBe(0);
  });

  test('③ 仅 webhook:update(无 read) — 菜单不可见', async ({ page }) => {
    await setViewerPerms([...BASE_PERMS, 'webhook:update']);
    await login(page, 'test', 'test');

    const menuItems = await page.locator('.ant-menu-item').allTextContents();
    console.log('仅 update 菜单:', JSON.stringify(menuItems));
    expect(menuItems.join()).not.toContain('系统设置');
  });

  test('④ read + update — 可编辑,有保存/测试按钮', async ({ page }) => {
    await setViewerPerms([...BASE_PERMS, 'webhook:read', 'webhook:update']);
    await login(page, 'test', 'test');
    await page.locator('.ant-menu-item').filter({ hasText: '系统设置' }).click();
    await page.waitForURL('**/settings');
    await page.locator('form').first().waitFor({ timeout: 15000 });

    const state = await readFormState(page);
    console.log('read+update:', JSON.stringify(state));

    expect(state.urlInputDisabled).toBe(false);
    expect(state.saveBtnCount).toBe(1);
    expect(state.testBtnCount).toBe(1);
  });

  test('admin — 完整功能', async ({ page }) => {
    await login(page, 'admin', 'admin123');
    await page.locator('.ant-menu-item').filter({ hasText: '系统设置' }).click();
    await page.waitForURL('**/settings');
    await page.locator('form').first().waitFor({ timeout: 15000 });

    const state = await readFormState(page);
    console.log('admin:', JSON.stringify(state));

    expect(state.urlInputDisabled).toBe(false);
    expect(state.saveBtnCount).toBe(1);
    expect(state.testBtnCount).toBe(1);
  });
});
