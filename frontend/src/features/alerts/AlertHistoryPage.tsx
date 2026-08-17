import { useEffect, useState } from 'react';
import { Table, Select, Space } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { getDevOpsDashboardAPI } from '../../api/client';
import type { Alert, GetAlertHistoryParams, GetAlertHistory200 } from '../../api/model';
import { AlertLevel } from '../../api/model';
import StatusTag, { type SemanticLevel } from '../../components/StatusTag';
import PageHeader from '../../components/PageHeader';
import { colors } from '../../theme/tokens';

const api = getDevOpsDashboardAPI(); // 模块级单例，避免每次渲染重建

// 级别 → 语义状态（颜色统一走 StatusTag / tokens）
const LEVEL_MAP: Record<string, SemanticLevel> = {
  info: 'success',
  warning: 'warning',
  critical: 'critical',
};

const LEVEL_LABELS: Record<string, string> = {
  info: '恢复/信息',
  warning: '警告',
  critical: '严重',
};

// AlertHistoryPage 告警历史页 — 落库告警记录,支持级别筛选 + 分页
export default function AlertHistoryPage() {
  const [data, setData] = useState<GetAlertHistory200['list']>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [level, setLevel] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(false);

  const load = async (p: number, ps: number, lv?: string) => {
    setLoading(true);
    try {
      const params: GetAlertHistoryParams = { page: p, pageSize: ps };
      if (lv) params.level = lv;
      const { data: res } = await api.getAlertHistory(params);
      setData(res.list);
      setTotal(res.total);
    } catch {
      setData([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(page, pageSize, level);
  }, [page, pageSize, level]);

  const columns: ColumnsType<Alert> = [
    {
      title: '级别',
      dataIndex: 'level',
      width: 110,
      render: (lv: AlertLevel) => (
        <StatusTag level={LEVEL_MAP[lv] || 'unknown'} label={LEVEL_LABELS[lv] ?? lv} />
      ),
    },
    { title: '消息', dataIndex: 'message', ellipsis: true },
    { title: '来源', dataIndex: 'source', width: 140 },
    { title: '时间', dataIndex: 'time', width: 120 },
    { title: '落库时间', dataIndex: 'createdAt', width: 180, render: (v: string) => (v ? new Date(v).toLocaleString() : '-') },
  ];

  return (
    <div style={{ width: '100%' }}>
      <PageHeader title="告警历史" />
      <Space style={{ marginBottom: 16 }}>
        <Select
          allowClear
          placeholder="级别筛选"
          style={{ width: 160, background: colors.bg.panel }}
          value={level}
          onChange={(v) => {
            setLevel(v);
            setPage(1);
          }}
          options={[
            { value: 'info', label: '恢复/信息' },
            { value: 'warning', label: '警告' },
            { value: 'critical', label: '严重' },
          ]}
        />
      </Space>
      <Table
        rowKey="id"
        columns={columns}
        dataSource={data}
        loading={loading}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (t) => `共 ${t} 条`,
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
        }}
        style={{ background: colors.bg.panel }}
      />
    </div>
  );
}
