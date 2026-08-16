import { useEffect, useState } from 'react';
import { Table, Tag, Select, Space } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { getDevOpsDashboardAPI } from '../../api/client';
import type { Alert, GetAlertHistoryParams, GetAlertHistory200 } from '../../api/model';
import { AlertLevel } from '../../api/model';

const api = getDevOpsDashboardAPI(); // 模块级单例，避免每次渲染重建

// 级别 → 颜色(与状态标签色值一致)
const LEVEL_COLORS: Record<string, string> = {
  info: '#73bf69',
  warning: '#f2c94c',
  critical: '#e02f44',
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
        <Tag color={LEVEL_COLORS[lv]} style={{ color: '#111' }}>
          {LEVEL_LABELS[lv] ?? lv}
        </Tag>
      ),
    },
    { title: '消息', dataIndex: 'message', ellipsis: true },
    { title: '来源', dataIndex: 'source', width: 140 },
    { title: '时间', dataIndex: 'time', width: 120 },
    { title: '落库时间', dataIndex: 'createdAt', width: 180, render: (v: string) => (v ? new Date(v).toLocaleString() : '-') },
  ];

  return (
    <div style={{ padding: 24, background: '#111217', minHeight: '100%' }}>
      <h1 style={{ color: '#ffffff', fontSize: 20, fontWeight: 600, marginBottom: 12 }}>告警历史</h1>
      <Space style={{ marginBottom: 16 }}>
        <Select
          allowClear
          placeholder="级别筛选"
          style={{ width: 160, background: '#1f1f1f' }}
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
        style={{ background: '#1f1f1f' }}
      />
    </div>
  );
}
