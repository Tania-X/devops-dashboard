import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';

// 浏览器通道由 playwright.config.js 自动探测（msedge→chrome→chromium），此处不硬编码
// 回归测试:新建角色后切到审计日志 Tab,应立即看到新记录(无需重登)
// 背景:Tabs 默认不销毁隐藏面板,审计组件只在首次挂载拉数据 → 需 destroyOnHidden

const ROLE_NAME = `audit-flush-${Date.now().toString().slice(-6)}`;

async function login(page: Page) {
  await page.goto('http://localhost:5173/login');
  await page.fill('#username', 'admin');
  await page.fill('#password', 'admin123');
  await page.click('button[type="submit"]');
  await page.waitForURL('**/');
  await page.locator('.ant-menu-item').first().waitFor({ timeout: 15000 });
}

async function goToSettings(page: Page) {
  await page.locator('.ant-menu-item').filter({ hasText: '系统设置' }).click();
  await page.waitForURL('**/settings');
}

test('新建角色后切审计 Tab 立即看到记录(无需重登)', async ({ page }) => {
  await login(page);
  await goToSettings(page);

  // 1. 先进审计 Tab,等表格出现,记录初始日志条数
  await page.locator('.ant-tabs-tab').filter({ hasText: '审计日志' }).click();
  await page.locator('tbody tr').first().waitFor({ timeout: 10000 });
  const initialCount = await page.locator('tbody tr').count();
  console.log('初始审计日志条数:', initialCount);

  // 2. 切到角色权限 Tab,新建一个角色
  await page.locator('.ant-tabs-tab').filter({ hasText: '角色权限' }).click();
  await page.waitForTimeout(500);
  await page.locator('button:has-text("新建角色")').click();
  const modal = page.locator('.ant-modal');
  await modal.waitFor({ state: 'visible' });
  await modal.locator('input').nth(0).fill(ROLE_NAME);
  await modal.locator('input').nth(1).fill('审计刷新测试');
  await modal.locator('.ant-modal-footer .ant-btn-primary').click();
  await page.waitForTimeout(1500); // 等创建完成

  // 3. 切回审计 Tab(不重登),应看到新增的 role.create 记录
  await page.locator('.ant-tabs-tab').filter({ hasText: '审计日志' }).click();
  await page.locator('tbody tr').first().waitFor({ timeout: 10000 });
  const afterRows = await page.locator('tbody tr').allTextContents();
  const afterCount = afterRows.length;
  console.log('切回后审计日志条数:', afterCount);
  console.log('含新角色记录:', afterRows.some((r) => r.includes(ROLE_NAME)));

  expect(afterRows.some((r) => r.includes(ROLE_NAME))).toBe(true);
});
