import { useEffect, useState } from 'react';
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

export default function StatPanel({ config }: StatPanelProps) {
  const [value, setValue] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);

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
          setValue(typeof raw === 'number' ? raw : Number(raw) || 0);
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

  return (
    <Card
      style={{
        background: colors.bg.panel,
        border: 'none',
        borderRadius: radius.panel,
        height: 120,
        boxShadow: '0 1px 2px rgba(0,0,0,0.3)',
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
            }}
          >
            {displayValue}
          </div>
        </div>
      )}
    </Card>
  );
}
