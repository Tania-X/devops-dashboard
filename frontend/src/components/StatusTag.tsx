import { Tag } from 'antd';
import { semanticStatus, radius } from '../theme/tokens';

export type SemanticLevel = keyof typeof semanticStatus;

interface StatusTagProps {
  /** 语义级别：success / warning / critical / info / unknown */
  level?: SemanticLevel;
  /** 展示文字（默认用级别的英文大写） */
  label?: string;
  /** 自定义文字色（覆盖 level 默认色；颜色应来自 theme/tokens.ts） */
  color?: string;
  /** 自定义背景色（覆盖 level 默认背景） */
  bg?: string;
}

/**
 * 统一状态标签 —— 替代各页面重复定义的 statusColorMap / levelColorMap。
 * 颜色来自 theme/tokens.ts 的 semanticStatus（spec 6.1），保证全系统一致。
 * 需要语义级别之外的专用色（如角色色）时，可用 color/bg 覆盖，但色值仍需取 tokens。
 */
export default function StatusTag({ level = 'unknown', label, color, bg }: StatusTagProps) {
  const cfg = semanticStatus[level] || semanticStatus.unknown;
  return (
    <Tag
      style={{
        background: bg ?? cfg.bg,
        color: color ?? cfg.color,
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
