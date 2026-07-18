import { useEffect, useState } from 'react';
import { Table, Tag, Drawer, Select, Space, Card, Descriptions, Spin, Row, Col, Progress } from 'antd';
import type { ServerItem, ServerDetail } from '../../api/model';
import { ServerItemStatus } from '../../api/model';
import { getDevOpsDashboardAPI } from '../../api/client';

const statusColorMap: Record<string, { color: string; bg: string; label: string }> = {
  running: { color: '#73bf69', bg: 'rgba(115, 191, 105, 0.2)', label: '运行中' },
  stopped: { color: '#aaaaaa', bg: 'rgba(170, 170, 170, 0.2)', label: '已停机' },
  maintenance: { color: '#f2c94c', bg: 'rgba(242, 201, 76, 0.2)', label: '维护中' },
};

function StatusTag({ status }: { status: string }) {
  const cfg = statusColorMap[status] || statusColorMap.stopped;
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

export default function ServerListPage() {
  const [data, setData] = useState<ServerItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 10, total: 0 });
  const [statusFilter, setStatusFilter] = useState<string | undefined>(undefined);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [selectedServer, setSelectedServer] = useState<ServerDetail | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);

  const fetchList = (page: number, pageSize: number, status?: string) => {
    setLoading(true);
    getDevOpsDashboardAPI()
      .getServerList({ page, pageSize, status: status as any })
      .then((res) => {
        setData(res.data.list);
        setPagination({
          current: res.data.page,
          pageSize: res.data.pageSize,
          total: res.data.total,
        });
      })
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchList(1, 10);
  }, []);

  const handleTableChange = (newPagination: any) => {
    fetchList(newPagination.current, newPagination.pageSize, statusFilter);
  };

  const handleStatusChange = (value: string | undefined) => {
    setStatusFilter(value);
    fetchList(1, pagination.pageSize, value);
  };

  const handleRowClick = (record: ServerItem) => {
    setDrawerVisible(true);
    setDetailLoading(true);
    getDevOpsDashboardAPI()
      .getServerDetail(record.id)
      .then((res) => {
        setSelectedServer(res.data);
      })
      .finally(() => setDetailLoading(false));
  };

  const columns = [
    {
      title: '主机名',
      dataIndex: 'hostname',
      key: 'hostname',
      render: (text: string) => (
        <span style={{ color: '#ffffff', fontFamily: '"Roboto Mono", monospace' }}>{text}</span>
      ),
    },
    {
      title: 'IP 地址',
      dataIndex: 'ip',
      key: 'ip',
      render: (text: string) => (
        <span style={{ color: '#aaaaaa', fontFamily: '"Roboto Mono", monospace' }}>{text}</span>
      ),
    },
    {
      title: '操作系统',
      dataIndex: 'os',
      key: 'os',
      render: (text: string) => <span style={{ color: '#aaaaaa' }}>{text}</span>,
    },
    {
      title: 'CPU',
      dataIndex: 'cpuCores',
      key: 'cpuCores',
      render: (v: number) => <span style={{ color: '#aaaaaa' }}>{v} 核</span>,
    },
    {
      title: '内存',
      dataIndex: 'memoryGb',
      key: 'memoryGb',
      render: (v: number) => <span style={{ color: '#aaaaaa' }}>{v} GB</span>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <StatusTag status={status} />,
    },
    {
      title: '运行时长',
      dataIndex: 'uptime',
      key: 'uptime',
      render: (text: string) => <span style={{ color: '#aaaaaa' }}>{text}</span>,
    },
  ];

  return (
    <div style={{ width: '100%' }}>
      <h1 style={{ color: '#ffffff', fontSize: 20, fontWeight: 600, marginBottom: 24 }}>
        服务器管理
      </h1>

      <Card
        style={{
          background: '#1f1f1f',
          border: 'none',
          borderRadius: 4,
        }}
        bodyStyle={{ padding: 16 }}
      >
        <Space style={{ marginBottom: 16 }}>
          <span style={{ color: '#aaaaaa' }}>状态筛选：</span>
          <Select
            allowClear
            placeholder="全部状态"
            value={statusFilter}
            onChange={handleStatusChange}
            style={{ width: 140 }}
            options={[
              { value: ServerItemStatus.running, label: '运行中' },
              { value: ServerItemStatus.stopped, label: '已停机' },
              { value: ServerItemStatus.maintenance, label: '维护中' },
            ]}
          />
        </Space>

        <Table
          columns={columns as any}
          dataSource={data}
          rowKey="id"
          loading={loading}
          pagination={{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total: pagination.total,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
          onChange={handleTableChange}
          onRow={(record) => ({
            onClick: () => handleRowClick(record),
            style: { cursor: 'pointer' },
          })}
          size="middle"
        />
      </Card>

      <Drawer
        title={
          <span style={{ color: '#ffffff' }}>
            {selectedServer?.hostname || '服务器详情'}
          </span>
        }
        open={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        width={560}
        styles={{
          body: { background: '#141414', padding: 24 },
          header: { background: '#1f1f1f', borderBottom: '1px solid #333333' },
          mask: { background: 'rgba(0,0,0,0.6)' },
        }}
      >
        {detailLoading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 40 }}>
            <Spin />
          </div>
        ) : selectedServer ? (
          <div>
            <Descriptions
              title={<span style={{ color: '#ffffff', fontSize: 16, fontWeight: 500 }}>基本信息</span>}
              column={2}
              labelStyle={{ color: '#aaaaaa' }}
              contentStyle={{ color: '#ffffff' }}
              items={[
                { key: '1', label: '主机名', children: selectedServer.hostname },
                { key: '2', label: 'IP', children: selectedServer.ip },
                { key: '3', label: '操作系统', children: selectedServer.os },
                { key: '4', label: 'CPU', children: `${selectedServer.cpuCores} 核` },
                { key: '5', label: '内存', children: `${selectedServer.memoryGb} GB` },
                {
                  key: '6',
                  label: '状态',
                  children: <StatusTag status={selectedServer.status} />,
                },
                { key: '7', label: '运行时长', children: selectedServer.uptime },
              ]}
            />

            <h3
              style={{
                color: '#ffffff',
                fontSize: 16,
                fontWeight: 500,
                marginTop: 32,
                marginBottom: 16,
                borderBottom: '1px solid #333333',
                paddingBottom: 8,
              }}
            >
              磁盘分区
            </h3>
            {selectedServer.diskPartitions?.map((disk) => {
              const percent = Math.round((disk.usedGb / disk.totalGb) * 100);
              return (
                <Row key={disk.mount} style={{ marginBottom: 12 }} align="middle">
                  <Col span={6}>
                    <span style={{ color: '#aaaaaa', fontFamily: '"Roboto Mono", monospace' }}>
                      {disk.mount}
                    </span>
                  </Col>
                  <Col span={12}>
                    <Progress
                      percent={percent}
                      size="small"
                      strokeColor={percent > 90 ? '#e02f44' : percent > 75 ? '#f2c94c' : '#73bf69'}
                      trailColor="#333333"
                      showInfo={false}
                    />
                  </Col>
                  <Col span={6} style={{ textAlign: 'right' }}>
                    <span style={{ color: '#aaaaaa', fontSize: 12 }}>
                      {disk.usedGb} / {disk.totalGb} GB
                    </span>
                  </Col>
                </Row>
              );
            })}

            <h3
              style={{
                color: '#ffffff',
                fontSize: 16,
                fontWeight: 500,
                marginTop: 32,
                marginBottom: 16,
                borderBottom: '1px solid #333333',
                paddingBottom: 8,
              }}
            >
              网络接口
            </h3>
            <Table
              dataSource={selectedServer.networkInterfaces}
              rowKey="name"
              size="small"
              pagination={false}
              columns={[
                {
                  title: <span style={{ color: '#aaaaaa' }}>名称</span>,
                  dataIndex: 'name',
                  key: 'name',
                  render: (v: string) => (
                    <span style={{ color: '#ffffff', fontFamily: '"Roboto Mono", monospace' }}>{v}</span>
                  ),
                },
                {
                  title: <span style={{ color: '#aaaaaa' }}>IP 地址</span>,
                  dataIndex: 'ip',
                  key: 'ip',
                  render: (v: string) => (
                    <span style={{ color: '#ffffff', fontFamily: '"Roboto Mono", monospace' }}>{v}</span>
                  ),
                },
                {
                  title: <span style={{ color: '#aaaaaa' }}>MAC 地址</span>,
                  dataIndex: 'mac',
                  key: 'mac',
                  render: (v: string) => (
                    <span style={{ color: '#ffffff', fontFamily: '"Roboto Mono", monospace' }}>{v}</span>
                  ),
                },
              ]}
            />
          </div>
        ) : null}
      </Drawer>
    </div>
  );
}
