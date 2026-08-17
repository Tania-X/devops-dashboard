import { useEffect, useRef, useState } from 'react';
import React from 'react';
import { Card, Spin } from 'antd';
import type { StatPanelConfig } from './dashboard-config';
import { getDevOpsDashboardAPI } from '../../api/client';
import * as icons from '@ant-design/icons';
import { colors, fonts, radius, spacing } from '../../theme/tokens';

interface StatPanelProps {
  config: StatPanelConfig;
}

function getValueByPath(obj: unknown, path: string): number | string | undefined {
  return path.split('.').reduce((acc: unknown, key) => {
    if (acc && typeof acc === 'object') {
      return (acc as Record<string, unknown>)[key];
    }
    return undefined;
  }, obj) as number | string | undefined;
}

function getStatusColor(value: number, thresholds?: { warning: number; critical: number }): string {
  if (!thresholds) return colors.text.primary;
  if (value >= thresholds.critical) return colors.status.critical;
  if (value >= thresholds.warning) return colors.status.warning;
  return colors.status.success;
}

/**
 * Stat Panel 大数字卡片
 * - 15s 轮询；对比上次采样显示趋势行（↑/↓/→，监控语义：数值升高用警告色）
 * - hover 上浮 + 阴影（spec 7 动画规范）
 */
export default function StatPanel({ config }: StatPanelProps) {
  const [value, setValue] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  /** 上次采样的值，用于计算采样间变化 */
  const prevValueRef = useRef<number | null>(null);
  const [delta, setDelta] = useState<number | null>(null);

  useEffect(() => {
    const api = getDevOpsDashboardAPI();
    let cancelled = false;
    let retryCount = 0;

    const fetchData = () => {
      api.getDashboardMetrics()
        .then((res) => {
          if (cancelled) return;
          retryCount = 0;
          const raw = getValueByPath(res.data, config.dataKey);
          const next = typeof raw === 'number' ? raw : Number(raw) || 0;
          if (prevValueRef.current !== null) {
            setDelta(Number((next - prevValueRef.current).toFixed(1)));
          }
          prevValueRef.current = next;
          setValue(next);
          if (!cancelled) setLoading(false);
        })
        .catch((err) => {
          console.error('[StatPanel] API 请求失败:', config.title, err);
          retryCount++;
          if (retryCount < 3 && !cancelled) {
            setTimeout(fetchData, 2000);
          } else if (!cancelled) {
            setLoading(false);
          }
        });
    };

    fetchData();

    const interval = setInterval(fetchData, 15000);

    return () => {
      cancelled = true;
      clearInterval(interval);
    };
  }, [config.api, config.dataKey]);

  const IconComponent = (icons as unknown as Record<string, React.ComponentType<any>>)[config.icon];

  const displayValue = value !== null ? `${value.toFixed(1)}${config.unit === 'percent' ? '%' : ''}` : '--';
  const statusColor = value !== null ? getStatusColor(value, config.thresholds) : colors.text.primary;

  // 趋势行：符号 + 数值。监控语义下数值升高=负载上升（警告色），下降=恢复（成功色）
  let trendText = '—';
  let trendColor: string = colors.text.muted;
  if (delta !== null) {
    if (Math.abs(delta) < 0.05) {
      trendText = '→ 持平';
    } else {
      trendText = `${delta > 0 ? '↑' : '↓'} ${Math.abs(delta).toFixed(1)}`;
      trendColor = delta > 0 ? colors.status.warning : colors.status.success;
    }
  }

  return (
    <Card
      className="stat-panel"
      style={{
        background: colors.bg.panel,
        border: 'none',
        borderRadius: radius.panel,
        height: 120,
        // boxShadow / transition 在 index.css 的 .stat-panel 类中定义，
        // 避免内联样式优先级覆盖 hover 规则（AI review round 1）
      }}
      styles={{ body: { padding: spacing.panel } }}
    >
      {loading ? (
        <Spin />
      ) : (
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
            {IconComponent && React.createElement(IconComponent, { style: { color: colors.text.secondary, fontSize: 14 } })}
            <span style={{ color: colors.text.secondary, fontSize: fonts.size.body }}>{config.title}</span>
          </div>
          <div
            style={{
              fontSize: fonts.size.number,
              fontWeight: fonts.weight.number,
              color: statusColor,
              fontFamily: fonts.mono,
              lineHeight: 1.1,
            }}
          >
            {displayValue}
          </div>
          {/* 趋势行：对比上次采样（15s 前） */}
          <div
            style={{
              color: trendColor,
              fontSize: fonts.size.caption,
              fontFamily: fonts.mono,
              marginTop: 4,
            }}
          >
            {trendText}
          </div>
        </div>
      )}
    </Card>
  );
}
