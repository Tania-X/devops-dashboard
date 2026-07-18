import { useEffect, useState } from 'react';
import { Table, Tag, Card, Spin } from 'antd';
import type { DeploymentItem, DeploymentHistoryItem } from '../../api/model';
import { getDevOpsDashboardAPI } from '../../api/client';
import dayjs from 'dayjs';

const envColorMap: Record<string, { color: string; bg: string; label: string }> = {
  dev: { color: '#3274d9', bg: 'rgba(50, 116, 217, 0.2)', label: '开发' },
  test: { color: '#f2c94c', bg: 'rgba(242, 201, 76, 0.2)', label: '测试' },
  prod: { color: '#e02f44', bg: 'rgba(224, 47, 68, 0.2)', label: '生产' },
};

const statusColorMap: Record<string, { color: string; bg: string; label: string }> = {
  pending: { color: '#aaaaaa', bg: 'rgba(170, 170, 170, 0.2)', label: '等待中' },
  deploying: { color: '#3274d9', bg: 'rgba(50, 116, 217, 0.2)', label: '部署中' },
  success: { color: '#73bf69', bg: 'rgba(115, 191, 105, 0.2)', label: '成功' },
  failed: { color: '#e02f44', bg: 'rgba(224, 47, 68, 0.2)', label: '失败' },
};

function EnvTag({ env }: { env: string }) {
  const cfg = envColorMap[env] || envColorMap.dev;
  return (
    <Tag
      style={{
        background: cfg.bg,
        color: cfg.color,
        border: 'none',
        borderRadius: 4,
        fontSize: 12,
        padding: '2px 8px',
      }}
    >
      {cfg.label}
    </Tag>
  );
}

function StatusTag({ status }: { status: string }) {
  const cfg = statusColorMap[status] || statusColorMap.pending;
  return (
    <Tag
      style={{
        background: cfg.bg,
        color: cfg.color,
        border: 'none',
        borderRadius: 4,
        fontSize: 12,
        padding: '2px 8px',
      }}
    >
      {cfg.label}
    </Tag>
  );
}

export default function DeploymentPage() {
  const [data, setData] = useState<DeploymentItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [historyMap, setHistoryMap] = useState<Record<string, DeploymentHistoryItem[]>>({});
  const [historyLoading, setHistoryLoading] = useState<Record<string, boolean>>({});

  useEffect(() => {
    setLoading(true);
    getDevOpsDashboardAPI()
      .getDeploymentList()
      .then((res) => {
        setData(res.data);
      })
      .finally(() => setLoading(false));
  }, []);

  const loadHistory = (id: string) => {
    if (historyMap[id]) return;
    setHistoryLoading((prev) => ({ ...prev, [id]: true }));
    getDevOpsDashboardAPI()
      .getDeploymentHistory(id)
      .then((res) => {
        setHistoryMap((prev) => ({ ...prev, [id]: res.data }));
      })
      .finally(() => {
        setHistoryLoading((prev) => ({ ...prev, [id]: false }));
      });
  };

  const columns = [
    {
      title: '应用名称',
      dataIndex: 'appName',
      key: 'appName',
      render: (text: string) => (
        <span style={{ color: '#ffffff', fontFamily: '"Roboto Mono", monospace', fontWeight: 500 }}>{text}</span>
      ),
    },
    {
      title: '当前版本',
      dataIndex: 'version',
      key: 'version',
      width: 120,
      render: (text: string) => (
        <span style={{ color: '#aaaaaa', fontFamily: '"Roboto Mono", monospace' }}>{text}</span>
      ),
    },
    {
      title: '环境',
      dataIndex: 'env',
      key: 'env',
      width: 100,
      render: (env: string) => <EnvTag env={env} />,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => <StatusTag status={status} />,
    },
    {
      title: '最后部署时间',
      dataIndex: 'lastDeployedAt',
      key: 'lastDeployedAt',
      width: 180,
      render: (text: string) => (
        <span style={{ color: '#aaaaaa', fontFamily: '"Roboto Mono", monospace' }}>{dayjs(text).format('YYYY-MM-DD HH:mm:ss')}</span>
      ),
    },
  ];

  const historyColumns = [
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 120,
      render: (text: string) => (
        <span style={{ color: '#ffffff', fontFamily: '"Roboto Mono", monospace' }}>{text}</span>
      ),
    },
    {
      title: '操作人',
      dataIndex: 'operator',
      key: 'operator',
      width: 140,
      render: (text: string) => <span style={{ color: '#cccccc' }}>{text}</span>,
    },
    {
      title: '耗时',
      dataIndex: 'durationSec',
      key: 'durationSec',
      width: 100,
      render: (v: number) => (
        <span style={{ color: '#aaaaaa' }}>{v >= 60 ? `${Math.floor(v / 60)}m${v % 60}s` : `${v}s`}</span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => <StatusTag status={status} />,
    },
    {
      title: '部署时间',
      dataIndex: 'deployedAt',
      key: 'deployedAt',
      width: 180,
      render: (text: string) => (
        <span style={{ color: '#aaaaaa', fontFamily: '"Roboto Mono", monospace' }}>{dayjs(text).format('YYYY-MM-DD HH:mm:ss')}</span>
      ),
    },
  ];

  return (
    <div style={{ width: '100%' }}>
      <h1 style={{ color: '#ffffff', fontSize: 20, fontWeight: 600, marginBottom: 24 }}>
        部署状态
      </h1>

      <Card
        style={{
          background: '#1f1f1f',
          border: 'none',
          borderRadius: 4,
        }}
        bodyStyle={{ padding: 16 }}
      >
        <Table
          columns={columns as any}
          dataSource={data}
          rowKey="id"
          loading={loading}
          pagination={false}
          size="middle"
          expandable={{
            onExpand: (expanded, record) => {
              if (expanded) loadHistory(record.id);
            },
            expandedRowRender: (record: DeploymentItem) => {
              const history = historyMap[record.id];
              const hLoading = historyLoading[record.id];
              return (
                <div style={{ background: '#141414', padding: '12px 24px' }}>
                  <h4
                    style={{
                      color: '#ffffff',
                      fontSize: 14,
                      fontWeight: 500,
                      marginBottom: 12,
                    }}
                  >
                    {record.appName} — 部署历史
                  </h4>
                  {hLoading ? (
                    <div style={{ display: 'flex', justifyContent: 'center', padding: 24 }}>
                      <Spin />
                    </div>
                  ) : (
                    <Table
                      dataSource={history || []}
                      rowKey="version"
                      size="small"
                      pagination={false}
                      columns={historyColumns as any}
                    />
                  )}
                </div>
              );
            },
            rowExpandable: () => true,
          }}
        />
      </Card>
    </div>
  );
}
