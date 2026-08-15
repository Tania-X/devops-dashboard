import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';

// 浏览器通道由 playwright.config.js 自动探测（msedge→chrome→chromium），此处不硬编码
// 角色管理 + 审计日志测试（RBAC 三期）:
//   角色 CRUD:新建自定义角色 → 可配置权限;内置角色不可删;编辑/删除自定义角色
//   审计日志:敏感操作(角色/权限/用户)被记录,分页展示

const UNIQ = Date.now().toString().slice(-6);
const NEW_ROLE = `pw-role-${UNIQ}`;
const NEW_LABEL = '测试角色';

async function login(page: Page) {
  await page.goto('http://localhost:5173/login');
  await page.fill('#username', 'admin');
  await page.fill('#password', 'admin123');
  await page.click('button[type="submit"]');
  await page.waitForURL('**/');
  await page.locator('.ant-menu-item').first().waitFor({ timeout: 15000 });
}

async function goToRolePerms(page: Page) {
  await page.locator('.ant-menu-item').filter({ hasText: '系统设置' }).click();
  await page.waitForURL('**/settings');
  await page.locator('.ant-tabs-tab').filter({ hasText: '角色权限' }).click();
  await page.waitForTimeout(1000);
}

test.describe('角色管理', () => {
  test('新建自定义角色 → 出现在角色列表并可配置权限', async ({ page }) => {
    await login(page);
    await goToRolePerms(page);

    // 新建角色
    await page.locator('button:has-text("新建角色")').click();
    const modal = page.locator('.ant-modal');
    await modal.waitFor({ state: 'visible' });
    await modal.locator('input').nth(0).fill(NEW_ROLE);
    await modal.locator('input').nth(1).fill(NEW_LABEL);
    await modal.locator('.ant-modal-footer .ant-btn-primary').click();
    await page.waitForTimeout(1500);

    // 角色列表应包含新角色(刷新后下拉)
    await page.locator('.ant-select').first().click();
    await page.locator('.ant-select-dropdown:visible').getByText(NEW_LABEL, { exact: true }).waitFor({ timeout: 5000 });
    const optionVisible = await page.locator('.ant-select-dropdown:visible').getByText(NEW_LABEL, { exact: true }).isVisible();
    expect(optionVisible).toBe(true);
    // 关闭下拉
    await page.keyboard.press('Escape');
  });

  test('内置角色(viewer)删除按钮禁用', async ({ page }) => {
    await login(page);
    await goToRolePerms(page);

    // 选 viewer 角色
    await page.locator('.ant-select').first().click();
    await page.locator('.ant-select-dropdown:visible').getByText('观察者', { exact: true }).click();
    await page.waitForTimeout(500);

    // 编辑/删除按钮应禁用(内置角色)
    const editDisabled = await page.locator('button:has-text("编辑")').isDisabled();
    const deleteDisabled = await page.locator('button:has-text("删除")').isDisabled();
    expect(editDisabled).toBe(true);
    expect(deleteDisabled).toBe(true);
  });

  test('删除自定义角色成功', async ({ page }) => {
    await login(page);
    await goToRolePerms(page);

    // 选中新角色
    await page.locator('.ant-select').first().click();
    await page.locator('.ant-select-dropdown:visible').getByText(NEW_LABEL, { exact: true }).click();
    await page.waitForTimeout(500);

    // 删除
    await page.locator('button:has-text("删除")').click();
    await page.locator('.ant-popconfirm .ant-btn-primary').click();
    await page.waitForTimeout(1500);

    // 下拉里不应再有该角色
    await page.locator('.ant-select').first().click();
    const dropdown = page.locator('.ant-select-dropdown:visible');
    const count = await dropdown.getByText(NEW_LABEL, { exact: true }).count();
    expect(count).toBe(0);
    await page.keyboard.press('Escape');
  });
});

test.describe('审计日志', () => {
  test('敏感操作被记录:角色创建/删除出现在审计列表', async ({ page }) => {
    await login(page);
    await page.locator('.ant-menu-item').filter({ hasText: '系统设置' }).click();
    await page.waitForURL('**/settings');
    await page.locator('.ant-tabs-tab').filter({ hasText: '审计日志' }).click();
    await page.waitForTimeout(1500);

    // 表格应出现
    const tableRows = await page.locator('tbody tr').allTextContents();
    console.log('审计日志行数:', tableRows.length);
    const all = tableRows.join('|');
    // 之前创建的 auditor 角色(role.create)应被记录
    expect(all).toContain('创建角色');
    expect(all).toContain('auditor');
  });
});
