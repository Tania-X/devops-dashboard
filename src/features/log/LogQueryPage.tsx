import { useEffect, useState } from 'react';
import { Table, Tag, Select, Space, Card, Input } from 'antd';
import type { LogItem } from '../../api/model';
import { LogItemLevel } from '../../api/model';
import { getDevOpsDashboardAPI } from '../../api/client';

const levelColorMap: Record<string, { color: string; bg: string; label: string }> = {
  INFO: { color: '#3274d9', bg: 'rgba(50, 116, 217, 0.2)', label: 'INFO' },
  WARN: { color: '#f2c94c', bg: 'rgba(242, 201, 76, 0.2)', label: 'WARN' },
  ERROR: { color: '#e02f44', bg: 'rgba(224, 47, 68, 0.2)', label: 'ERROR' },
};

function LevelTag({ level }: { level: string }) {
  const cfg = levelColorMap[level] || levelColorMap.INFO;
  return (
    <Tag
      style={{
        background: cfg.bg,
        color: cfg.color,
        border: 'none',
        borderRadius: 4,
        fontSize: 12,
        padding: '2px 8px',
        fontFamily: '"Roboto Mono", monospace',
      }}
    >
      {cfg.label}
    </Tag>
  );
}

export default function LogQueryPage() {
  const [data, setData] = useState<LogItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 10, total: 0 });
  const [levelFilter, setLevelFilter] = useState<string | undefined>(undefined);
  const [keyword, setKeyword] = useState('');

  const fetchList = (page: number, pageSize: number, level?: string, searchKeyword?: string) => {
    setLoading(true);
    getDevOpsDashboardAPI()
      .getLogList({ page, pageSize, level: level as any, keyword: searchKeyword || undefined })
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
    fetchList(newPagination.current, newPagination.pageSize, levelFilter, keyword);
  };

  const handleLevelChange = (value: string | undefined) => {
    setLevelFilter(value);
    fetchList(1, pagination.pageSize, value, keyword);
  };

  const handleSearch = (value: string) => {
    setKeyword(value);
    fetchList(1, pagination.pageSize, levelFilter, value);
  };

  const columns = [
    {
      title: '时间',
      dataIndex: 'time',
      key: 'time',
      width: 180,
      render: (text: string) => (
        <span style={{ color: '#aaaaaa', fontFamily: '"Roboto Mono", monospace' }}>{text}</span>
      ),
    },
    {
      title: '级别',
      dataIndex: 'level',
      key: 'level',
      width: 100,
      render: (level: string) => <LevelTag level={level} />,
    },
    {
      title: '服务',
      dataIndex: 'service',
      key: 'service',
      width: 180,
      render: (text: string) => (
        <span style={{ color: '#ffffff', fontFamily: '"Roboto Mono", monospace' }}>{text}</span>
      ),
    },
    {
      title: '内容',
      dataIndex: 'content',
      key: 'content',
      ellipsis: { showTitle: true },
      render: (text: string) => <span style={{ color: '#cccccc' }}>{text}</span>,
    },
  ];

  return (
    <div style={{ width: '100%' }}>
      <h1 style={{ color: '#ffffff', fontSize: 20, fontWeight: 600, marginBottom: 24 }}>
        日志查询
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
          <span style={{ color: '#aaaaaa' }}>级别筛选：</span>
          <Select
            allowClear
            placeholder="全部级别"
            value={levelFilter}
            onChange={handleLevelChange}
            style={{ width: 140 }}
            options={[
              { value: LogItemLevel.INFO, label: 'INFO' },
              { value: LogItemLevel.WARN, label: 'WARN' },
              { value: LogItemLevel.ERROR, label: 'ERROR' },
            ]}
          />
          <Input.Search
            placeholder="搜索日志内容关键词"
            allowClear
            onSearch={handleSearch}
            style={{ width: 280 }}
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
          size="middle"
        />
      </Card>
    </div>
  );
}
