import { defineConfig } from '@playwright/test';

export default defineConfig({
    testDir: './tests',
    timeout: 30000,
    retries: 0,
    use: {
        baseURL: 'http://localhost:5173',
        screenshot: 'only-on-failure',
        channel: 'msedge',  // 使用本机 Edge，无需下载 Chromium
    },
});
