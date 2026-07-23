import { useEffect, useState } from 'react';
import { Table, Tag, Drawer, Select, Space, Card, Descriptions, Spin, Input, Progress, Row, Col } from 'antd';
import type { ProcessItem, ProcessDetail } from '../../api/model';
import { getDevOpsDashboardAPI } from '../../api/client';

const statusColorMap: Record<string, { color: string; bg: string; label: string }> = {
  running: { color: '#73bf69', bg: 'rgba(115, 191, 105, 0.2)', label: '运行中' },
  sleep: { color: '#73bf69', bg: 'rgba(115, 191, 105, 0.1)', label: '睡眠' },
  idle: { color: '#aaaaaa', bg: 'rgba(170, 170, 170, 0.2)', label: '空闲' },
  stop: { color: '#e02f44', bg: 'rgba(224, 47, 68, 0.2)', label: '已停止' },
  zombie: { color: '#f2c94c', bg: 'rgba(242, 201, 76, 0.2)', label: '僵尸' },
  disk: { color: '#f2c94c', bg: 'rgba(242, 201, 76, 0.2)', label: '磁盘等待' },
};

function ProcessStatusTag({ status }: { status: string }) {
  const key = status.toLowerCase();
  const cfg = statusColorMap[key] || { color: '#aaaaaa', bg: 'rgba(170, 170, 170, 0.2)', label: status };
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

export default function ProcessListPage() {
  const [data, setData] = useState<ProcessItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [sortBy, setSortBy] = useState<string>('cpu');
  const [order, setOrder] = useState<string>('desc');
  const [limit, setLimit] = useState<number>(50);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [selectedProcess, setSelectedProcess] = useState<ProcessDetail | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);

  const fetchList = () => {
    setLoading(true);
    getDevOpsDashboardAPI()
      .getProcessList({ sortBy: sortBy as any, order: order as any, keyword: keyword || undefined, limit })
      .then((res) => {
        setData(res.data);
      })
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchList();
    // 每 10 秒自动刷新一次
    const interval = setInterval(fetchList, 10000);
    return () => clearInterval(interval);
  }, [sortBy, order, keyword, limit]);

  const handleRowClick = (record: ProcessItem) => {
    setDrawerVisible(true);
    setDetailLoading(true);
    getDevOpsDashboardAPI()
      .getProcessDetail(record.pid)
      .then((res) => {
        setSelectedProcess(res.data);
      })
      .finally(() => setDetailLoading(false));
  };

  const columns = [
    {
      title: 'PID',
      dataIndex: 'pid',
      key: 'pid',
      width: 80,
      render: (v: number) => (
        <span style={{ color: '#ffffff', fontFamily: '"Roboto Mono", monospace' }}>{v}</span>
      ),
    },
    {
      title: '进程名',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => <span style={{ color: '#ffffff', fontWeight: 500 }}>{text}</span>,
    },
    {
      title: 'CPU %',
      dataIndex: 'cpuPercent',
      key: 'cpuPercent',
      width: 120,
      render: (v: number) => {
        const color = v > 50 ? '#e02f44' : v > 20 ? '#f2c94c' : '#73bf69';
        return (
          <Space>
            <Progress
              percent={Math.min(v, 100)}
              size="small"
              strokeColor={color}
              trailColor="#333333"
              showInfo={false}
              style={{ width: 60 }}
            />
            <span style={{ color, fontFamily: '"Roboto Mono", monospace', fontSize: 12 }}>
              {v.toFixed(1)}
            </span>
          </Space>
        );
      },
    },
    {
      title: '内存 %',
      dataIndex: 'memoryPercent',
      key: 'memoryPercent',
      width: 120,
      render: (v: number) => {
        const color = v > 30 ? '#e02f44' : v > 10 ? '#f2c94c' : '#73bf69';
        return (
          <Space>
            <Progress
              percent={Math.min(v, 100)}
              size="small"
              strokeColor={color}
              trailColor="#333333"
              showInfo={false}
              style={{ width: 60 }}
            />
            <span style={{ color, fontFamily: '"Roboto Mono", monospace', fontSize: 12 }}>
              {v.toFixed(1)}
            </span>
          </Space>
        );
      },
    },
    {
      title: '内存 (MB)',
      dataIndex: 'memoryMb',
      key: 'memoryMb',
      width: 110,
      render: (v: number) => (
        <span style={{ color: '#aaaaaa', fontFamily: '"Roboto Mono", monospace', fontSize: 12 }}>
          {v.toFixed(1)}
        </span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (status: string) => <ProcessStatusTag status={status} />,
    },
  ];

  return (
    <div style={{ width: '100%' }}>
      <h1 style={{ color: '#ffffff', fontSize: 20, fontWeight: 600, marginBottom: 24 }}>
        进程列表
      </h1>

      <Card
        style={{
          background: '#1f1f1f',
          border: 'none',
          borderRadius: 4,
        }}
        bodyStyle={{ padding: 16 }}
      >
        <Space style={{ marginBottom: 16 }} wrap>
          <Input.Search
            placeholder="搜索进程名..."
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onSearch={() => fetchList()}
            allowClear
            style={{ width: 200 }}
          />
          <span style={{ color: '#aaaaaa' }}>排序：</span>
          <Select
            value={sortBy}
            onChange={(v) => setSortBy(v)}
            style={{ width: 100 }}
            options={[
              { value: 'cpu', label: 'CPU' },
              { value: 'memory', label: '内存' },
              { value: 'pid', label: 'PID' },
              { value: 'name', label: '名称' },
            ]}
          />
          <Select
            value={order}
            onChange={(v) => setOrder(v)}
            style={{ width: 90 }}
            options={[
              { value: 'desc', label: '降序' },
              { value: 'asc', label: '升序' },
            ]}
          />
          <span style={{ color: '#aaaaaa' }}>条数：</span>
          <Select
            value={limit}
            onChange={(v) => setLimit(v)}
            style={{ width: 80 }}
            options={[
              { value: 20, label: '20' },
              { value: 50, label: '50' },
              { value: 100, label: '100' },
            ]}
          />
        </Space>

        <Table
          columns={columns as any}
          dataSource={data}
          rowKey="pid"
          loading={loading}
          pagination={{ pageSize: limit, showTotal: (total) => `共 ${total} 个进程` }}
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
            {selectedProcess?.name || '进程详情'} (PID: {selectedProcess?.pid})
          </span>
        }
        open={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        width={600}
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
        ) : selectedProcess ? (
          <div>
            <Descriptions
              title={<span style={{ color: '#ffffff', fontSize: 16, fontWeight: 500 }}>基本信息</span>}
              column={2}
              labelStyle={{ color: '#aaaaaa' }}
              contentStyle={{ color: '#ffffff' }}
              items={[
                { key: '1', label: '进程名', children: selectedProcess.name },
                { key: '2', label: 'PID', children: selectedProcess.pid },
                { key: '3', label: 'PPID', children: selectedProcess.ppid },
                { key: '4', label: '用户名', children: selectedProcess.username },
                { key: '5', label: '状态', children: <ProcessStatusTag status={selectedProcess.status} /> },
                { key: '6', label: '创建时间', children: selectedProcess.createTime },
                { key: '7', label: '线程数', children: selectedProcess.numThreads },
                { key: '8', label: '打开文件数', children: selectedProcess.numOpenFiles },
                { key: '9', label: '网络连接', children: selectedProcess.numConnections },
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
              资源占用
            </h3>
            <Row gutter={24} style={{ marginBottom: 16 }}>
              <Col span={8}>
                <div style={{ color: '#aaaaaa', fontSize: 12, marginBottom: 4 }}>CPU 使用率</div>
                <Space>
                  <Progress
                    type="circle"
                    percent={Math.min(Math.round(selectedProcess.cpuPercent * 10) / 10, 100)}
                    size={60}
                    strokeColor={selectedProcess.cpuPercent > 50 ? '#e02f44' : '#177ddc'}
                    trailColor="#333333"
                    format={(pct) => `${pct?.toFixed(1)}%`}
                  />
                </Space>
              </Col>
              <Col span={8}>
                <div style={{ color: '#aaaaaa', fontSize: 12, marginBottom: 4 }}>内存使用率</div>
                <Space>
                  <Progress
                    type="circle"
                    percent={Math.min(Math.round(selectedProcess.memoryPercent * 10) / 10, 100)}
                    size={60}
                    strokeColor={selectedProcess.memoryPercent > 30 ? '#f2c94c' : '#73bf69'}
                    trailColor="#333333"
                    format={(pct) => `${pct?.toFixed(1)}%`}
                  />
                </Space>
              </Col>
              <Col span={8}>
                <div style={{ color: '#aaaaaa', fontSize: 12, marginBottom: 4 }}>内存占用</div>
                <span style={{ color: '#ffffff', fontFamily: '"Roboto Mono", monospace', fontSize: 16 }}>
                  {selectedProcess.memoryMb.toFixed(1)} MB
                </span>
              </Col>
            </Row>

            {selectedProcess.cmdline && (
              <>
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
                  命令行
                </h3>
                <div
                  style={{
                    background: '#1a1a1a',
                    border: '1px solid #333333',
                    borderRadius: 4,
                    padding: 12,
                    fontFamily: '"Roboto Mono", monospace',
                    fontSize: 12,
                    color: '#cccccc',
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-all',
                    maxHeight: 200,
                    overflow: 'auto',
                  }}
                >
                  {selectedProcess.cmdline}
                </div>
              </>
            )}

            {selectedProcess.workingDir && (
              <>
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
                  工作目录
                </h3>
                <div
                  style={{
                    color: '#cccccc',
                    fontFamily: '"Roboto Mono", monospace',
                    fontSize: 12,
                  }}
                >
                  {selectedProcess.workingDir}
                </div>
              </>
            )}

            {selectedProcess.env && selectedProcess.env.length > 0 && (
              <>
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
                  环境变量
                </h3>
                <div
                  style={{
                    background: '#1a1a1a',
                    border: '1px solid #333333',
                    borderRadius: 4,
                    padding: 12,
                    fontFamily: '"Roboto Mono", monospace',
                    fontSize: 11,
                    color: '#888888',
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-all',
                    maxHeight: 300,
                    overflow: 'auto',
                  }}
                >
                  {selectedProcess.env.map((e) => (
                    <div key={e} style={{ marginBottom: 2 }}>
                      {e}
                    </div>
                  ))}
                </div>
              </>
            )}
          </div>
        ) : null}
      </Drawer>
    </div>
  );
}
