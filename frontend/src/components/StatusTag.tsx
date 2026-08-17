import { Tag } from 'antd';
import { semanticStatus, radius } from '../theme/tokens';

export type SemanticLevel = keyof typeof semanticStatus;

interface StatusTagProps {
  /** 语义级别：success / warning / critical / info / unknown */
  level?: SemanticLevel;
  /** 展示文字（默认用级别的英文大写） */
  label?: string;
}

/**
 * 统一状态标签 —— 替代各页面重复定义的 statusColorMap / levelColorMap。
 * 颜色来自 theme/tokens.ts 的 semanticStatus（spec 6.1），保证全系统一致。
 */
export default function StatusTag({ level = 'unknown', label }: StatusTagProps) {
  const cfg = semanticStatus[level] || semanticStatus.unknown;
  return (
    <Tag
      style={{
        background: cfg.bg,
        color: cfg.color,
        border: 'none',
        borderRadius: radius.panel,
        fontSize: 12,
        padding: '2px 8px',
        marginInlineEnd: 0,
      }}
    >
      {label ?? level.toUpperCase()}
    </Tag>
  );
}
