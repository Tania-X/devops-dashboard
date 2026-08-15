import { useEffect, useState, useCallback } from 'react';
import { Card, Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { getDevOpsDashboardAPI } from '../../api/client';
import type { AuditLog } from '../../api/model';

const api = getDevOpsDashboardAPI(); // 模块级单例

// 动作 → 标签颜色映射（审计日志展示）
const ACTION_META: Record<string, { label: string; color: string }> = {
  'role.create': { label: '创建角色', color: 'green' },
  'role.update': { label: '更新角色', color: 'blue' },
  'role.delete': { label: '删除角色', color: 'red' },
  'permission.update': { label: '更新权限', color: 'purple' },
  'user.create': { label: '新增用户', color: 'green' },
  'user.update': { label: '编辑用户', color: 'blue' },
  'user.delete': { label: '删除用户', color: 'red' },
};

function actionLabel(action: string) {
  return ACTION_META[action]?.label ?? action;
}

function actionColor(action: string) {
  return ACTION_META[action]?.color ?? 'default';
}

// AuditLogs 审计日志面板 — 记录敏感管理操作（角色/权限/用户增删改）
export default function AuditLogs() {
  const [items, setItems] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const PAGE_SIZE = 20;

  const load = useCallback(async (p: number) => {
    setLoading(true);
    try {
      const { data } = await api.listAuditLogs({ page: p, size: PAGE_SIZE });
      setItems(data.items);
      setTotal(data.total);
    } catch {
      setItems([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load(page);
  }, [page, load]);

  const columns: ColumnsType<AuditLog> = [
    {
      title: '时间',
      dataIndex: 'createdAt',
      width: 180,
      render: (v: string) => new Date(v).toLocaleString(),
    },
    {
      title: '操作人',
      dataIndex: 'actor',
      width: 120,
    },
    {
      title: '动作',
      dataIndex: 'action',
      width: 110,
      render: (v: string) => <Tag color={actionColor(v)}>{actionLabel(v)}</Tag>,
    },
    {
      title: '对象',
      dataIndex: 'target',
      width: 160,
    },
    {
      title: '详情',
      dataIndex: 'detail',
      ellipsis: true,
      render: (v: string) => <span style={{ color: '#bbb', fontSize: 12 }}>{v}</span>,
    },
  ];

  return (
    <Card
      title={<span style={{ color: '#fff', fontSize: 16 }}>审计日志</span>}
      style={{ background: '#1f1f1f', border: '1px solid #333', maxWidth: 860 }}
    >
      <Table<AuditLog>
        rowKey="id"
        columns={columns}
        dataSource={items}
        loading={loading}
        pagination={{
          current: page,
          pageSize: PAGE_SIZE,
          total,
          showSizeChanger: false,
          onChange: setPage,
        }}
        size="small"
      />
    </Card>
  );
}
