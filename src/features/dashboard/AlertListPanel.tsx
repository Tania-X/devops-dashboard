import { useEffect, useState } from 'react';
import { Card, Spin, Tag, List } from 'antd';
import type { AlertListPanelConfig } from './dashboard-config';
import { getDevOpsDashboardAPI } from '../../api/client';
import type { AlertItem } from '../../api/model';

interface AlertListPanelProps {
  config: AlertListPanelConfig;
}

const levelColorMap: Record<string, string> = {
  info: '#3274d9',
  warning: '#f2c94c',
  critical: '#e02f44',
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
        background: '#1f1f1f',
        border: 'none',
        borderRadius: 4,
      }}
      headStyle={{
        color: '#ffffff',
        borderBottom: '1px solid #333333',
        fontSize: 16,
        fontWeight: 500,
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
                borderBottom: '1px solid #333333',
                padding: '12px 0',
              }}
            >
              <div style={{ width: '100%' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                  <Tag
                    color={levelColorMap[item.level] || '#aaaaaa'}
                    style={{
                      background: `${levelColorMap[item.level] || '#aaaaaa'}33`,
                      border: 'none',
                    }}
                  >
                    {item.level.toUpperCase()}
                  </Tag>
                  <span style={{ color: '#666666', fontSize: 12 }}>{item.time}</span>
                </div>
                <div style={{ color: '#ffffff', fontSize: 14 }}>{item.message}</div>
                <div style={{ color: '#aaaaaa', fontSize: 12, marginTop: 4 }}>来源: {item.source}</div>
              </div>
            </List.Item>
          )}
        />
      )}
    </Card>
  );
}
