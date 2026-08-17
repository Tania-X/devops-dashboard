import { useEffect, useState } from 'react';
import { Card, Spin, List } from 'antd';
import type { AlertListPanelConfig } from './dashboard-config';
import { getDevOpsDashboardAPI } from '../../api/client';
import type { AlertItem } from '../../api/model';
import StatusTag from '../../components/StatusTag';
import { colors, fonts, radius } from '../../theme/tokens';

interface AlertListPanelProps {
  config: AlertListPanelConfig;
}

/** 告警级别 → 语义状态（颜色统一走 StatusTag / tokens） */
const levelMap: Record<string, 'info' | 'warning' | 'critical' | 'unknown'> = {
  info: 'info',
  warning: 'warning',
  critical: 'critical',
};

export default function AlertListPanel({ config }: AlertListPanelProps) {
  const [alerts, setAlerts] = useState<AlertItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const api = getDevOpsDashboardAPI();
    api
      .getDashboardAlerts({ limit: config.limit })
      .then((res) => setAlerts(res.data))
      .finally(() => setLoading(false));
  }, [config.limit]);

  return (
    <Card
      title={config.title}
      style={{
        background: colors.bg.panel,
        border: 'none',
        borderRadius: radius.panel,
      }}
      headStyle={{
        color: colors.text.primary,
        borderBottom: `1px solid ${colors.border}`,
        fontSize: fonts.size.h2,
        fontWeight: fonts.weight.h2,
      }}
    >
      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: 24 }}>
          <Spin />
        </div>
      ) : (
        <List
          dataSource={alerts}
          renderItem={(item) => (
            <List.Item
              style={{
                borderBottom: `1px solid ${colors.border}`,
                padding: '12px 0',
              }}
            >
              <div style={{ width: '100%' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                  <StatusTag level={levelMap[item.level] || 'unknown'} />
                  <span style={{ color: colors.text.muted, fontSize: fonts.size.caption }}>{item.time}</span>
                </div>
                <div style={{ color: colors.text.primary, fontSize: fonts.size.body }}>{item.message}</div>
                <div style={{ color: colors.text.secondary, fontSize: fonts.size.caption, marginTop: 4 }}>
                  来源: {item.source}
                </div>
              </div>
            </List.Item>
          )}
        />
      )}
    </Card>
  );
}
