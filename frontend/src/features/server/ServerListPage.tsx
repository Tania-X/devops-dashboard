import { useEffect, useState } from 'react';
import { Table, Drawer, Select, Space, Card, Descriptions, Spin, Row, Col, Progress } from 'antd';
import type { ServerItem, ServerDetail } from '../../api/model';
import { ServerItemStatus } from '../../api/model';
import { getDevOpsDashboardAPI } from '../../api/client';
import StatusTag, { type SemanticLevel } from '../../components/StatusTag';
import PageHeader from '../../components/PageHeader';
import { colors, fonts, radius, spacing } from '../../theme/tokens';

const statusLevelMap: Record<string, SemanticLevel> = {
  running: 'success',
  stopped: 'unknown',
  maintenance: 'warning',
};

const statusLabelMap: Record<string, string> = {
  running: '运行中',
  stopped: '已停机',
  maintenance: '维护中',
};

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
        <span style={{ color: colors.text.primary, fontFamily: fonts.mono }}>{text}</span>
      ),
    },
    {
      title: 'IP 地址',
      dataIndex: 'ip',
      key: 'ip',
      render: (text: string) => (
        <span style={{ color: colors.text.secondary, fontFamily: fonts.mono }}>{text}</span>
      ),
    },
    {
      title: '操作系统',
      dataIndex: 'os',
      key: 'os',
      render: (text: string) => <span style={{ color: colors.text.secondary }}>{text}</span>,
    },
    {
      title: 'CPU',
      dataIndex: 'cpuCores',
      key: 'cpuCores',
      render: (v: number) => <span style={{ color: colors.text.secondary }}>{v} 核</span>,
    },
    {
      title: '内存',
      dataIndex: 'memoryGb',
      key: 'memoryGb',
      render: (v: number) => <span style={{ color: colors.text.secondary }}>{v} GB</span>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <StatusTag level={statusLevelMap[status] || 'unknown'} label={statusLabelMap[status] || status} />
      ),
    },
    {
      title: '运行时长',
      dataIndex: 'uptime',
      key: 'uptime',
      render: (text: string) => <span style={{ color: colors.text.secondary }}>{text}</span>,
    },
  ];

  return (
    <div style={{ width: '100%' }}>
      <PageHeader title="服务器管理" />

      <Card
        style={{
          background: colors.bg.panel,
          border: 'none',
          borderRadius: radius.panel,
        }}
        styles={{ body: { padding: spacing.panel } }}
      >
        <Space style={{ marginBottom: 16 }}>
          <span style={{ color: colors.text.secondary }}>状态筛选：</span>
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
        title={<span style={{ color: colors.text.primary }}>{selectedServer?.hostname || '服务器详情'}</span>}
        open={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        width={560}
        styles={{
          body: { background: colors.bg.page, padding: 24 },
          header: { background: colors.bg.panel, borderBottom: `1px solid ${colors.border}` },
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
              title={<span style={{ color: colors.text.primary, fontSize: fonts.size.h2, fontWeight: fonts.weight.h2 }}>基本信息</span>}
              column={2}
              styles={{
                label: { color: colors.text.secondary },
                content: { color: colors.text.primary },
              }}
              items={[
                { key: '1', label: '主机名', children: selectedServer.hostname },
                { key: '2', label: 'IP', children: selectedServer.ip },
                { key: '3', label: '操作系统', children: selectedServer.os },
                { key: '4', label: 'CPU', children: `${selectedServer.cpuCores} 核` },
                { key: '5', label: '内存', children: `${selectedServer.memoryGb} GB` },
                {
                  key: '6',
                  label: '状态',
                  children: (
                    <StatusTag
                      level={statusLevelMap[selectedServer.status] || 'unknown'}
                      label={statusLabelMap[selectedServer.status] || selectedServer.status}
                    />
                  ),
                },
                { key: '7', label: '运行时长', children: selectedServer.uptime },
              ]}
            />

            <h3
              style={{
                color: colors.text.primary,
                fontSize: fonts.size.h2,
                fontWeight: fonts.weight.h2,
                marginTop: 32,
                marginBottom: 16,
                borderBottom: `1px solid ${colors.border}`,
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
                    <span style={{ color: colors.text.secondary, fontFamily: fonts.mono }}>{disk.mount}</span>
                  </Col>
                  <Col span={12}>
                    <Progress
                      percent={percent}
                      size="small"
                      strokeColor={percent > 90 ? colors.status.critical : percent > 75 ? colors.status.warning : colors.status.success}
                      railColor={colors.border}
                      showInfo={false}
                    />
                  </Col>
                  <Col span={6} style={{ textAlign: 'right' }}>
                    <span style={{ color: colors.text.secondary, fontSize: fonts.size.caption }}>
                      {disk.usedGb} / {disk.totalGb} GB
                    </span>
                  </Col>
                </Row>
              );
            })}

            <h3
              style={{
                color: colors.text.primary,
                fontSize: fonts.size.h2,
                fontWeight: fonts.weight.h2,
                marginTop: 32,
                marginBottom: 16,
                borderBottom: `1px solid ${colors.border}`,
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
                  title: <span style={{ color: colors.text.secondary }}>名称</span>,
                  dataIndex: 'name',
                  key: 'name',
                  render: (v: string) => (
                    <span style={{ color: colors.text.primary, fontFamily: fonts.mono }}>{v}</span>
                  ),
                },
                {
                  title: <span style={{ color: colors.text.secondary }}>IP 地址</span>,
                  dataIndex: 'ip',
                  key: 'ip',
                  render: (v: string) => (
                    <span style={{ color: colors.text.primary, fontFamily: fonts.mono }}>{v}</span>
                  ),
                },
                {
                  title: <span style={{ color: colors.text.secondary }}>MAC 地址</span>,
                  dataIndex: 'mac',
                  key: 'mac',
                  render: (v: string) => (
                    <span style={{ color: colors.text.primary, fontFamily: fonts.mono }}>{v}</span>
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
