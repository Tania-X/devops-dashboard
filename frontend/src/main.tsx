import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { ConfigProvider, theme } from 'antd'
import { colors, fonts, radius } from './theme/tokens'
import './index.css'
import './api/auth-interceptor' // 注册 axios 请求拦截器（自动附加 JWT，独立于 orval 生成文件）
import App from './App.tsx'

async function bootstrap() {
  if (import.meta.env.VITE_USE_MSW === 'true') {
    const { worker } = await import('./mocks/browser')
    await worker.start({
      onUnhandledRequest: 'error',
    })
  }

  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <ConfigProvider
        theme={{
          algorithm: theme.darkAlgorithm,
          token: {
            // 与 spec/ui-theme.md 的映射关系（见规范第 9 节）
            colorBgBase: colors.bg.page,           // --bg-page
            colorBgContainer: colors.bg.panel,     // --bg-panel
            colorBgElevated: colors.bg.panelHover, // 下拉/弹层背景
            colorTextBase: colors.text.primary,    // --text-primary
            colorTextSecondary: colors.text.secondary,
            colorTextTertiary: colors.text.muted,
            colorPrimary: colors.brand.primary,    // --accent-primary
            colorSuccess: colors.status.success,   // 正常态
            colorWarning: colors.status.warning,   // 警告态
            colorError: colors.status.critical,    // 严重态
            colorInfo: colors.status.info,         // 信息态
            colorBorder: colors.border,
            borderRadius: radius.panel,            // Panel 圆角
            fontSize: fonts.size.body,             // Body 字号
            fontFamily: fonts.family,
          },
        }}
      >
        <App />
      </ConfigProvider>
    </StrictMode>,
  )
}

bootstrap()
