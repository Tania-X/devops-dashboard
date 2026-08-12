import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { ConfigProvider, theme } from 'antd'
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
            colorBgBase: '#141414',
            colorBgContainer: '#1f1f1f',
            colorTextBase: '#ffffff',
            colorPrimary: '#177ddc',
            colorSuccess: '#73bf69',
            colorWarning: '#f2c94c',
            colorError: '#e02f44',
            borderRadius: 4,
            fontSize: 14,
            fontFamily: '"Inter", "PingFang SC", "Microsoft YaHei", sans-serif',
          },
        }}
      >
        <App />
      </ConfigProvider>
    </StrictMode>,
  )
}

bootstrap()
