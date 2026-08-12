import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';

// 浏览器通道由 playwright.config.js 自动探测（msedge→chrome→chromium），此处不硬编码
// Webhook 设置页权限分层测试:
//   webhook:read   → 菜单可见,表单只读,无操作按钮
//   webhook:update → 表单可编辑 + 保存按钮
//   webhook:test   → 测试推送按钮
// 依赖关系:read 是基础(菜单显隐),update/test 需配合 read 才可达

const VIEWER_READ_ONLY_PERMS = [
  'dashboard:view', 'server:read', 'log:read', 'deployment:read', 'monitor:read', 'webhook:read',
];

async function login(page: Page, username: string, password: string) {
  await page.goto('http://localhost:5173/login');
  await page.fill('#username', username);
  await page.fill('#password', password);
  await page.click('button[type="submit"]');
  await page.waitForURL('**/');
  await page.locator('.ant-menu-item').first().waitFor({ timeout: 15000 });
}

async function goToSettings(page: Page) {
  await page.locator('.ant-menu-item').filter({ hasText: '系统设置' }).click();
  await page.waitForURL('**/settings');
  await page.locator('form').first().waitFor({ timeout: 15000 });
}

async function readFormState(page: Page) {
  return page.evaluate(() => {
    const qa = (sel: string) => Array.from(document.querySelectorAll(sel));
    const isDisabled = (el: Element | null) => el ? (el as HTMLInputElement).disabled ?? el.hasAttribute('disabled') : null;
    return {
      urlInputDisabled: isDisabled(qa('input[placeholder*="qyapi"]')[0] ?? null),
      saveBtnCount: qa('button').filter((b) => b.textContent?.includes('保存')).length,
      testBtnCount: qa('button').filter((b) => b.textContent?.includes('测试推送')).length,
      readOnlyAlert: qa('.ant-alert').filter((a) => a.textContent?.includes('只读')).length,
    };
  });
}

// 通过 admin 重置 viewer 角色权限为仅只读(webhook:read),保证测试前置条件确定
async function resetViewerToReadOnly() {
  const res = await fetch('http://localhost:8080/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'admin123' }),
  });
  const { token } = await res.json();
  await fetch('http://localhost:8080/api/settings/roles/viewer', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ permissions: VIEWER_READ_ONLY_PERMS }),
  });
}

test.describe('Webhook 设置页权限分层', () => {
  test('viewer(仅 webhook:read) — 表单只读,无保存/测试按钮', async ({ page }) => {
    await resetViewerToReadOnly(); // 先重置角色权限
    await login(page, 'test', 'test');
    await goToSettings(page);

    const state = await readFormState(page);
    console.log('viewer 仅 read:', JSON.stringify(state));

    expect(state.urlInputDisabled).toBe(true);   // 输入框禁用
    expect(state.saveBtnCount).toBe(0);          // 无保存按钮
    expect(state.testBtnCount).toBe(0);          // 无测试按钮
    expect(state.readOnlyAlert).toBeGreaterThan(0); // 有只读提示
  });

  test('admin — 表单可编辑,有保存/测试按钮', async ({ page }) => {
    await login(page, 'admin', 'admin123');
    await goToSettings(page);

    const state = await readFormState(page);
    console.log('admin:', JSON.stringify(state));

    expect(state.urlInputDisabled).toBe(false);  // 输入框可编辑
    expect(state.saveBtnCount).toBe(1);          // 有保存按钮
    expect(state.testBtnCount).toBe(1);          // 有测试按钮
    expect(state.readOnlyAlert).toBe(0);         // 无只读提示
  });
});

