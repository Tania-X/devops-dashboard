import { colors, fonts } from '../theme/tokens';

interface PageHeaderProps {
  title: string;
  /** 标题下方的说明文字（可选） */
  description?: string;
}

/**
 * 统一页面标题 —— 替代各页面重复的 <h1 style={{ color:'#fff', fontSize:20... }}>
 * 规格：spec 5 字体规范 H1 页面标题 20px/600
 */
export default function PageHeader({ title, description }: PageHeaderProps) {
  return (
    <div style={{ marginBottom: 24 }}>
      <h1
        style={{
          color: colors.text.primary,
          fontSize: fonts.size.h1,
          fontWeight: fonts.weight.h1,
          margin: 0,
        }}
      >
        {title}
      </h1>
      {description && (
        <p style={{ color: colors.text.secondary, fontSize: fonts.size.caption, margin: '6px 0 0' }}>
          {description}
        </p>
      )}
    </div>
  );
}
