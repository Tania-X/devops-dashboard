import { colors } from '../theme/tokens';

interface AppLogoProps {
  /** 是否显示文字部分（登录页大 logo 显示，侧边栏小 logo 也显示） */
  showText?: boolean;
  size?: number;
}

/**
 * DevOps Dashboard 品牌 Logo —— 签名元素。
 * 图形：三个上升的脉冲柱（信号/监控的抽象），下方一条扫描线，
 * 寓意"持续观测、心跳在线"。渐变蓝呼应品牌色。
 */
export default function AppLogo({ showText = true, size = 32 }: AppLogoProps) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
      <svg
        width={size}
        height={size}
        viewBox="0 0 48 48"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        aria-label="DevOps Dashboard logo"
      >
        <defs>
          <linearGradient id="ddg-brand" x1="0" y1="48" x2="48" y2="0">
            <stop offset="0%" stopColor={colors.brand.primary} />
            <stop offset="100%" stopColor={colors.brand.gradientEnd} />
          </linearGradient>
        </defs>
        {/* 三个脉冲柱（等宽数字风格：信号强度） */}
        <rect x="6" y="22" width="7" height="16" rx="2" fill="url(#ddg-brand)" opacity="0.55" />
        <rect x="20.5" y="12" width="7" height="26" rx="2" fill="url(#ddg-brand)" opacity="0.8" />
        <rect x="35" y="4" width="7" height="34" rx="2" fill="url(#ddg-brand)" />
        {/* 底部扫描线 */}
        <path d="M4 42 H44" stroke={colors.status.success} strokeWidth="2.5" strokeLinecap="round" opacity="0.9" />
      </svg>
      {showText && (
        <span
          style={{
            color: colors.text.primary,
            fontSize: 18,
            fontWeight: 600,
            letterSpacing: 0.5,
            whiteSpace: 'nowrap',
          }}
        >
          DevOps
        </span>
      )}
    </div>
  );
}
