import { test, expect } from '@playwright/test';

test.describe('登出后登录页', () => {
  test('退出后登录页不应预填账号密码', async ({ page }) => {
    // 登录
    await page.goto('http://localhost:5173/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'admin123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/');

    // 点退出
    await page.click('button:has-text("退出")');
    await page.waitForURL('**/login');
    await page.waitForSelector('#username');
    await page.waitForTimeout(500);

    const username = await page.locator('#username').inputValue();
    const password = await page.locator('#password').inputValue();

    console.log('登出后 username:', JSON.stringify(username));
    console.log('登出后 password:', JSON.stringify(password));

    expect(username).toBe('');
    expect(password).toBe('');
  });
});
