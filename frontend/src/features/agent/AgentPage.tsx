import { useEffect, useState, useCallback } from 'react';
import { Table, Button, Modal, Form, Input, InputNumber, Select, Tag, Space, Card, message, Popconfirm } from 'antd';
import { PlusOutlined, ReloadOutlined, CloudUploadOutlined, StopOutlined } from '@ant-design/icons';
import { getDevOpsDashboardAPI } from '../../api/client';
import type { AgentTarget } from '../../api/model';

const statusColorMap: Record<string, string> = {
  online: '#73bf69',
  offline: '#e02f44',
  unknown: '#aaaaaa',
};

const statusLabelMap: Record<string, string> = {
  online: '在线',
  offline: '离线',
  unknown: '未知',
};

const api = getDevOpsDashboardAPI(); // 模块级单例，避免每次渲染重建导致 useCallback 失效

export default function AgentPage() {
  const [data, setData] = useState<AgentTarget[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<AgentTarget | null>(null);
  const [form] = Form.useForm();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getAgentList();
      setData(res.data);
    } catch {
      message.error('获取 Agent 列表失败');
    } finally {
      setLoading(false);
    }
  }, [api]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ port: 22, deployDir: '/opt/devops-agent', agentPort: 9100 });
    setModalOpen(true);
  };

  const openEdit = (record: AgentTarget) => {
    setEditing(record);
    form.setFieldsValue(record);
    setModalOpen(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editing) {
        await api.updateAgent(editing.id, values);
        message.success('更新成功');
      } else {
        await api.createAgent(values);
        message.success('创建成功');
      }
      setModalOpen(false);
      fetchData();
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error(err?.response?.data?.error || '操作失败');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await api.deleteAgent(id);
      message.success('删除成功');
      fetchData();
    } catch {
      message.error('删除失败');
    }
  };

  const handleDeploy = async (id: string) => {
    try {
      await api.deployAgent(id);
      message.success('部署已触发');
      fetchData();
    } catch {
      message.error('部署失败');
    }
  };

  const handleStop = async (id: string) => {
    try {
      await api.stopAgent(id);
      message.success('已停止');
      fetchData();
    } catch {
      message.error('停止失败');
    }
  };

  const columns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '主机地址', dataIndex: 'host', key: 'host' },
    { title: 'SSH 端口', dataIndex: 'port', key: 'port' },
    { title: '用户名', dataIndex: 'username', key: 'username' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={statusColorMap[status] || '#aaaaaa'}>
          {statusLabelMap[status] || status}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: unknown, record: AgentTarget) => (
        <Space size="small">
          <Button size="small" onClick={() => openEdit(record)}>编辑</Button>
          <Button size="small" icon={<CloudUploadOutlined />} onClick={() => handleDeploy(record.id)}>部署</Button>
          <Popconfirm title="确定停止该 Agent?" onConfirm={() => handleStop(record.id)}>
            <Button size="small" danger icon={<StopOutlined />}>停止</Button>
          </Popconfirm>
          <Popconfirm title="确定删除?" onConfirm={() => handleDelete(record.id)}>
            <Button size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24, background: '#111217', minHeight: '100%' }}>
      <Card
        title={<span style={{ color: '#fff', fontSize: 16 }}>Agent 分发目标</span>}
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新增目标</Button>
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

      <Modal
        title={<span style={{ color: '#fff' }}>{editing ? '编辑目标' : '新增目标'}</span>}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => setModalOpen(false)}
        okText="确认"
        cancelText="取消"
        styles={{
          content: { background: '#1f1f1f' },
          header: { background: '#1f1f1f' },
        }}
      >
        <Form form={form} layout="vertical" autoComplete="off">
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input autoComplete="off" />
          </Form.Item>
          <Form.Item name="host" label="主机地址" rules={[{ required: true, message: '请输入主机地址' }]}>
            <Input autoComplete="off" />
          </Form.Item>
          <Form.Item name="port" label="SSH 端口" rules={[{ required: true, message: '请输入端口' }]}>
            <InputNumber min={1} max={65535} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="authType" label="认证方式" rules={[{ required: true, message: '请选择认证方式' }]}>
            <Select options={[{ label: 'Password', value: 'password' }, { label: 'Key', value: 'key' }]} />
          </Form.Item>
          <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input autoComplete="off" />
          </Form.Item>
          <Form.Item name="password" label="密码/密钥">
            <Input.Password autoComplete="new-password" />
          </Form.Item>
          <Form.Item name="deployDir" label="部署目录" rules={[{ required: true, message: '请输入部署目录' }]}>
            <Input autoComplete="off" />
          </Form.Item>
          <Form.Item name="agentPort" label="Agent 端口" rules={[{ required: true, message: '请输入端口' }]}>
            <InputNumber min={1} max={65535} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
