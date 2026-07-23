import { useEffect, useState } from 'react';
import React from 'react';
import { Card, Spin } from 'antd';
import type { StatPanelConfig } from './dashboard-config';
import { getDevOpsDashboardAPI } from '../../api/client';
import * as icons from '@ant-design/icons';

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
  if (!thresholds) return '#ffffff';
  if (value >= thresholds.critical) return '#e02f44';
  if (value >= thresholds.warning) return '#f2c94c';
  return '#73bf69';
}

export default function StatPanel({ config }: StatPanelProps) {
  const [value, setValue] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const api = getDevOpsDashboardAPI();
    let cancelled = false;

    const fetchData = () => {
      api.getDashboardMetrics()
        .then((res) => {
          if (cancelled) return;
          const raw = getValueByPath(res.data, config.dataKey);
          setValue(typeof raw === 'number' ? raw : Number(raw) || 0);
        })
        .catch((err) => {
          console.error('[StatPanel] API 请求失败:', err);
        })
        .finally(() => {
          if (!cancelled) setLoading(false);
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
  const statusColor = value !== null ? getStatusColor(value, config.thresholds) : '#ffffff';

  return (
    <Card
      style={{
        background: '#1f1f1f',
        border: 'none',
        borderRadius: 4,
        height: 120,
      }}
      bodyStyle={{ padding: 16 }}
    >
      {loading ? (
        <Spin />
      ) : (
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
            {IconComponent && React.createElement(IconComponent, { style: { color: '#aaaaaa', fontSize: 14 } })}
            <span style={{ color: '#aaaaaa', fontSize: 14 }}>{config.title}</span>
          </div>
          <div
            style={{
              fontSize: 32,
              fontWeight: 700,
              color: statusColor,
              fontFamily: '"Roboto Mono", monospace',
            }}
          >
            {displayValue}
          </div>
        </div>
      )}
    </Card>
  );
}
