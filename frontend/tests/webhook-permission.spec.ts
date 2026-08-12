import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';

// 浏览器通道由 playwright.config.js 自动探测（msedge→chrome→chromium），此处不硬编码
// Webhook 权限依赖联动测试:
//   权限点:webhook:read(查看) / webhook:update(配置+测试推送),update 依赖 read
//   依赖规则:仅提交 update 时后端自动补 read(隐式继承);前端勾 update 自动勾 read
//   组合空间:死配置"仅 update 无 read"被依赖机制消除

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

test.describe('Webhook 权限依赖联动', () => {
  test('后端:提交仅 update 自动补 read(死配置从数据层消除)', async ({ page }) => {
    // 提交仅 webhook:update(不提交 read),依赖规则应自动补全
    await setViewerPerms([...BASE_PERMS, 'webhook:update']);

    // 用 admin 读回 viewer 实际权限
    const res = await fetch('http://localhost:8080/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: 'admin', password: 'admin123' }),
    });
    const { token } = await res.json();
    const rolesRes = await fetch('http://localhost:8080/api/settings/roles', {
      headers: { Authorization: `Bearer ${token}` },
    });
    const { roles } = await rolesRes.json();
    const viewer = roles.find((r: { name: string }) => r.name === 'viewer');
    console.log('viewer 实际权限:', JSON.stringify(viewer.permissions));

    expect(viewer.permissions).toContain('webhook:update');
    expect(viewer.permissions).toContain('webhook:read'); // 依赖自动补全
  });

  test('UI:viewer(update 自动含 read) 设置页可编辑有按钮', async ({ page }) => {
    await login(page, 'test', 'test');
    await page.locator('.ant-menu-item').filter({ hasText: '系统设置' }).click();
    await page.waitForURL('**/settings');
    await page.locator('form').first().waitFor({ timeout: 15000 });

    const state = await readFormState(page);
    console.log('update 自动含 read 后:', JSON.stringify(state));

    expect(state.urlInputDisabled).toBe(false); // read 补上后可看
    expect(state.saveBtnCount).toBe(1);         // update 有保存
    expect(state.testBtnCount).toBe(1);         // 测试推送随 update
  });

  test('UI:权限配置矩阵勾 update 自动勾 read(联动)', async ({ page }) => {
    await login(page, 'admin', 'admin123');
    // 进入设置页 → 角色权限 Tab
    await page.locator('.ant-menu-item').filter({ hasText: '系统设置' }).click();
    await page.waitForURL('**/settings');
    await page.locator('.ant-tabs-tab').filter({ hasText: '角色权限' }).click();
    await page.waitForTimeout(1000);

    // 选择 viewer 角色(默认选中第一个非锁定,可能就是 viewer)
    // 找到 webhook 分组中的"配置 Webhook"(webhook:update) 复选框,点击勾选
    const updateCheckbox = page.locator('label').filter({ hasText: '配置 Webhook' }).first();
    await updateCheckbox.waitFor({ timeout: 10000 });
    await updateCheckbox.click(); // 勾选 update

    // 断言:查看 Webhook(webhook:read) 被自动勾上
    const readCheckbox = page.locator('label').filter({ hasText: '查看 Webhook' }).first();
    const readChecked = await readCheckbox.locator('input').isChecked();
    console.log('勾 update 后 read 是否自动勾选:', readChecked);
    expect(readChecked).toBe(true);
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
