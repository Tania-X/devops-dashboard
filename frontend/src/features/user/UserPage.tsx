import { useEffect, useState, useCallback } from 'react';
import { Table, Button, Modal, Form, Input, Select, Tag, Space, Card, message, Popconfirm } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { getDevOpsDashboardAPI } from '../../api/client';
import type { UserItem } from '../../api/model';

const roleColorMap: Record<string, string> = {
  admin: '#534AB7',
  viewer: '#888780',
};

const roleLabelMap: Record<string, string> = {
  admin: '管理员',
  viewer: '观察者',
};

interface UserFormModalProps {
  open: boolean;
  editingUser: UserItem | null;
  onClose: () => void;
  onSuccess: () => void;
}

function UserFormModal({ open, editingUser, onClose, onSuccess }: UserFormModalProps) {
  const [form] = Form.useForm();
  const [submitting, setSubmitting] = useState(false);
  const editingId = editingUser?.id;
  const editingRole = editingUser?.role;
  const api = getDevOpsDashboardAPI();

  useEffect(() => {
    if (open) {
      if (editingUser) {
        form.setFieldsValue({ role: editingUser.role, password: '' });
      } else {
        form.resetFields();
      }
    }
  }, [open, editingUser, editingRole, editingId, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);
      if (editingUser) {
        await api.updateUser(editingUser.id, values);
        message.success('更新成功');
      } else {
        await api.createUser(values);
        message.success('创建成功');
      }
      onSuccess();
      onClose();
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error(err?.response?.data?.error || '操作失败');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      title={<span style={{ color: '#fff' }}>{editingUser ? '编辑用户' : '新增用户'}</span>}
      open={open}
      onOk={handleSubmit}
      onCancel={onClose}
      confirmLoading={submitting}
      okText="确认"
      cancelText="取消"
      styles={{
        content: { background: '#1f1f1f' },
        header: { background: '#1f1f1f' },
      }}
    >
      <Form form={form} layout="vertical" autoComplete="off" key={editingId || 'create'}>
        <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
          <Input autoComplete="off" disabled={!!editingUser} />
        </Form.Item>
        <Form.Item name="password" label="密码" rules={[{ required: !editingUser, message: '请输入密码' }]}>
          <Input.Password autoComplete="new-password" placeholder={editingUser ? '留空则不修改' : ''} />
        </Form.Item>
        <Form.Item name="role" label="角色" rules={[{ required: true, message: '请选择角色' }]}>
          <Select options={[{ label: '管理员', value: 'admin' }, { label: '观察者', value: 'viewer' }]} />
        </Form.Item>
      </Form>
    </Modal>
  );
}

export default function UserPage() {
  const [data, setData] = useState<UserItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<UserItem | null>(null);
  const api = getDevOpsDashboardAPI();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getUserList();
      setData(res.data);
    } catch {
      message.error('获取用户列表失败');
    } finally {
      setLoading(false);
    }
  }, [api]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const openCreate = () => {
    setEditingUser(null);
    setModalOpen(true);
  };

  const openEdit = (record: UserItem) => {
    setEditingUser(record);
    setModalOpen(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await api.deleteUser(id);
      message.success('删除成功');
      fetchData();
    } catch {
      message.error('删除失败');
    }
  };

  const columns = [
    { title: '用户名', dataIndex: 'username', key: 'username' },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => (
        <Tag color={roleColorMap[role] || '#888780'}>
          {roleLabelMap[role] || role}
        </Tag>
      ),
    },
    { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt' },
    {
      title: '操作',
      key: 'actions',
      render: (_: unknown, record: UserItem) => (
        <Space size="small">
          <Button size="small" onClick={() => openEdit(record)}>编辑</Button>
          <Popconfirm title="确定删除该用户?" onConfirm={() => handleDelete(record.id)}>
            <Button size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24, background: '#111217', minHeight: '100%' }}>
      <Card
        title={<span style={{ color: '#fff', fontSize: 16 }}>用户管理</span>}
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新增用户</Button>
          </Space>
        }
        style={{ background: '#1f1f1f', border: '1px solid #333' }}
        styles={{ header: { borderBottom: '1px solid #333' }, body: { padding: 0 } }}
      >
        <Table
          dataSource={data}
          columns={columns}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 10 }}
          style={{ background: '#1f1f1f' }}
        />
      </Card>

      <UserFormModal
        open={modalOpen}
        editingUser={editingUser}
        onClose={() => setModalOpen(false)}
        onSuccess={fetchData}
      />
    </div>
  );
}
