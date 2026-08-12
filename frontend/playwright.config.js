import { defineConfig } from '@playwright/test';
import { existsSync } from 'fs';

/**
 * 浏览器通道自动探测 — 按优先级选择本机可用的浏览器。
 *
 * 优先级：msedge → chrome → chromium(playwright 内置，需 install)
 * 检测方式：检查 Edge/Chrome 在 Windows/macOS/Linux 的常见安装路径是否存在。
 * 同一份配置多电脑通用：本机用 Edge；装了 Chrome 的机器自动用 Chrome；
 * 都没有则回退 playwright 自带 chromium（需先 `npx playwright install chromium`）。
 */
function detectBrowserChannel() {
    const candidates = [
        {
            channel: 'msedge',
            paths: [
                'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
                'C:/Program Files/Microsoft/Edge/Application/msedge.exe',
                '/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge',
                '/usr/bin/microsoft-edge',
                '/usr/bin/microsoft-edge-stable',
            ],
        },
        {
            channel: 'chrome',
            paths: [
                'C:/Program Files/Google/Chrome/Application/chrome.exe',
                'C:/Program Files (x86)/Google/Chrome/Application/chrome.exe',
                '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
                '/usr/bin/google-chrome',
                '/usr/bin/google-chrome-stable',
            ],
        },
        { channel: 'chromium', paths: [] }, // playwright 内置，兜底
    ];

    for (const c of candidates) {
        if (c.paths.length === 0 || c.paths.some((p) => existsSync(p))) {
            return c.channel;
        }
    }
    return 'chromium';
}

const channel = detectBrowserChannel();
console.log(`[playwright] 检测到浏览器通道: ${channel}`);

export default defineConfig({
    testDir: './tests',
    timeout: 30000,
    retries: 0,
    use: {
        baseURL: 'http://localhost:5173',
        screenshot: 'only-on-failure',
        channel, // 自动探测，多电脑兼容
    },
});
