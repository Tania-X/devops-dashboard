import { useEffect, useState } from 'react';
import { Table, Card, Spin } from 'antd';
import type { DeploymentItem, DeploymentHistoryItem } from '../../api/model';
import { getDevOpsDashboardAPI } from '../../api/client';
import dayjs from 'dayjs';
import StatusTag, { type SemanticLevel } from '../../components/StatusTag';
import PageHeader from '../../components/PageHeader';
import { colors, fonts, radius, spacing } from '../../theme/tokens';

/** 环境 → 语义状态 */
const envLevelMap: Record<string, SemanticLevel> = {
  dev: 'info',
  test: 'warning',
  prod: 'critical',
};

const envLabelMap: Record<string, string> = {
  dev: '开发',
  test: '测试',
  prod: '生产',
};

/** 部署状态 → 语义状态 */
const statusLevelMap: Record<string, SemanticLevel> = {
  pending: 'unknown',
  deploying: 'info',
  success: 'success',
  failed: 'critical',
};

const statusLabelMap: Record<string, string> = {
  pending: '等待中',
  deploying: '部署中',
  success: '成功',
  failed: '失败',
};

function EnvTag({ env }: { env: string }) {
  return <StatusTag level={envLevelMap[env] || 'unknown'} label={envLabelMap[env] || env} />;
}

function StatusTagView({ status }: { status: string }) {
  return <StatusTag level={statusLevelMap[status] || 'unknown'} label={statusLabelMap[status] || status} />;
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
        <span style={{ color: colors.text.primary, fontFamily: fonts.mono, fontWeight: 500 }}>{text}</span>
      ),
    },
    {
      title: '当前版本',
      dataIndex: 'version',
      key: 'version',
      width: 120,
      render: (text: string) => (
        <span style={{ color: colors.text.secondary, fontFamily: fonts.mono }}>{text}</span>
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
      render: (status: string) => <StatusTagView status={status} />,
    },
    {
      title: '最后部署时间',
      dataIndex: 'lastDeployedAt',
      key: 'lastDeployedAt',
      width: 180,
      render: (text: string) => (
        <span style={{ color: colors.text.secondary, fontFamily: fonts.mono }}>{dayjs(text).format('YYYY-MM-DD HH:mm:ss')}</span>
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
        <span style={{ color: colors.text.primary, fontFamily: fonts.mono }}>{text}</span>
      ),
    },
    {
      title: '操作人',
      dataIndex: 'operator',
      key: 'operator',
      width: 140,
      render: (text: string) => <span style={{ color: colors.codeText }}>{text}</span>,
    },
    {
      title: '耗时',
      dataIndex: 'durationSec',
      key: 'durationSec',
      width: 100,
      render: (v: number) => (
        <span style={{ color: colors.text.secondary }}>{v >= 60 ? `${Math.floor(v / 60)}m${v % 60}s` : `${v}s`}</span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => <StatusTagView status={status} />,
    },
    {
      title: '部署时间',
      dataIndex: 'deployedAt',
      key: 'deployedAt',
      width: 180,
      render: (text: string) => (
        <span style={{ color: colors.text.secondary, fontFamily: fonts.mono }}>{dayjs(text).format('YYYY-MM-DD HH:mm:ss')}</span>
      ),
    },
  ];

  return (
    <div style={{ width: '100%' }}>
      <PageHeader title="部署状态" />

      <Card
        style={{
          background: colors.bg.panel,
          border: 'none',
          borderRadius: radius.panel,
        }}
        styles={{ body: { padding: spacing.panel } }}
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
                <div style={{ background: colors.bg.page, padding: '12px 24px' }}>
                  <h4
                    style={{
                      color: colors.text.primary,
                      fontSize: fonts.size.body,
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
