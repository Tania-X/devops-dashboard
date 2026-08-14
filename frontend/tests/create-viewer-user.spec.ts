import { test, expect } from '@playwright/test';
import type { Page, APIRequestContext } from '@playwright/test';

const BASE = 'http://localhost:8080';

// 登录 admin 并返回 token(供 API 直连测试用)
async function loginAndGetToken(request: APIRequestContext): Promise<string> {
  const res = await request.post(`${BASE}/api/auth/login`, {
    data: { username: 'admin', password: 'admin123' },
  });
  const body = await res.json();
  return body.token;
}

// UI 登录(admin/admin123)
async function uiLogin(page: Page) {
  await page.goto('http://localhost:5173/login');
  await page.fill('#username', 'admin');
  await page.fill('#password', 'admin123');
  await page.click('button[type="submit"]');
  await page.waitForURL('**/');
}

test.describe('新增观察者用户排查', () => {
  test('API 直连: 后端能否创建 viewer 用户(诊断后端,时间戳用户名不撞库)', async ({ request }) => {
    const token = await loginAndGetToken(request);
    expect(token).toBeTruthy();

    const username = `api-viewer-${Date.now()}`;
    const res = await request.post(`${BASE}/api/users`, {
      headers: { Authorization: `Bearer ${token}` },
      data: { username, password: 'test123', role: 'viewer' },
    });
    const body = await res.json();
    console.log(`POST /api/users (${username}) -> ${res.status()}`, JSON.stringify(body));

    expect(res.status()).toBe(201);
    expect(body.role).toBe('viewer');
    // 安全断言: 响应不得泄露密码字段(后端 User 模型 json:"-" 排除,故 password 应为 undefined)
    expect(body.password).toBeUndefined();
  });

  test('UI 全流程: 新增 test/test/观察者', async ({ page }) => {
    // 捕获 POST /api/users 的真实响应
    let postRes: { status: number; body: string } | null = null;
    page.on('response', async (res) => {
      if (res.url().includes('/api/users') && res.request().method() === 'POST') {
        postRes = { status: res.status(), body: await res.text().catch(() => '(无 body)') };
      }
    });
    page.on('console', (msg) => {
      if (msg.type() === 'error') console.log('[console.error]', msg.text());
    });

    await uiLogin(page);
    await page.click('text=用户管理');
    await page.waitForURL('**/users');
    await page.waitForSelector('table', { timeout: 10000 });

    // 打开新增用户弹窗
    await page.click('button:has-text("新增用户")');
    const modal = page.locator('.ant-modal');
    await modal.waitFor({ state: 'visible' });

    // 填写: test / test / 观察者
    await modal.locator('input').nth(0).fill('test');
    await modal.locator('input').nth(1).fill('test');
    await modal.locator('.ant-select').click();
    await page.locator('.ant-select-dropdown:visible').getByText('观察者', { exact: true }).click();

    console.log('表单已填写,提交中...');
    await modal.locator('button:has-text("确认")').click();

    // 等待提交结果
    await page.waitForTimeout(3000);

    console.log('POST /api/users 响应:', postRes ? `${postRes.status} ${postRes.body}` : '未捕获到请求');
    console.log('页面消息提示:', JSON.stringify(await page.locator('.ant-message').allTextContents()));
    console.log('弹窗是否仍打开:', await modal.isVisible().catch(() => false));

    // 点刷新,确认列表里是否出现 test
    await page.click('button:has-text("刷新")');
    await page.waitForTimeout(1500);
    const rows = await page.locator('tbody tr').allTextContents();
    console.log('用户列表行:', JSON.stringify(rows));
    const found = rows.some((r) => r.includes('test'));
    console.log('列表是否出现 test 用户:', found);

    await page.screenshot({ path: 'test-results/create-viewer-final.png', fullPage: true });
    console.log('截图: test-results/create-viewer-final.png');

    // 断言
    expect(postRes?.status).toBe(201);
    expect(found).toBeTruthy();
  });
});
