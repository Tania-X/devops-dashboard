import { useEffect, useState, useMemo } from 'react';
import { Table, Tag, Select, Space, Card, Input, DatePicker } from 'antd';
import type { LogItem } from '../../api/model';
import { LogItemLevel } from '../../api/model';
import { getDevOpsDashboardAPI } from '../../api/client';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;

const SERVICE_LIST = [
  'api-gateway', 'user-service', 'order-service', 'payment-service',
  'notification-service', 'auth-service', 'log-collector', 'monitor-agent',
];

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

type LogItemExt = LogItem & {
  sourceHost?: string;
  logPath?: string;
  traceId?: string;
};

export default function LogQueryPage() {
  const [data, setData] = useState<LogItemExt[]>([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 10, total: 0 });
  const [levelFilter, setLevelFilter] = useState<string | undefined>(undefined);
  const [serviceFilter, setServiceFilter] = useState<string | undefined>(undefined);
  const [timeRange, setTimeRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);
  const [keyword, setKeyword] = useState('');

  const fetchList = (
    page: number,
    pageSize: number,
    level?: string,
    searchKeyword?: string,
    service?: string,
    range?: [dayjs.Dayjs, dayjs.Dayjs] | null
  ) => {
    setLoading(true);
    const params: any = { page, pageSize };
    if (level) params.level = level;
    if (searchKeyword) params.keyword = searchKeyword;
    if (service) params.service = service;
    if (range) {
      params.startTime = range[0].startOf('day').format('YYYY-MM-DD HH:mm:ss');
      params.endTime = range[1].endOf('day').format('YYYY-MM-DD HH:mm:ss');
    }

    getDevOpsDashboardAPI()
      .getLogList(params)
      .then((res) => {
        setData(res.data.list as LogItemExt[]);
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
    fetchList(newPagination.current, newPagination.pageSize, levelFilter, keyword, serviceFilter, timeRange);
  };

  const handleLevelChange = (value: string | undefined) => {
    setLevelFilter(value);
    fetchList(1, pagination.pageSize, value, keyword, serviceFilter, timeRange);
  };

  const handleServiceChange = (value: string | undefined) => {
    setServiceFilter(value);
    fetchList(1, pagination.pageSize, levelFilter, keyword, value, timeRange);
  };

  const handleTimeRangeChange = (dates: any) => {
    const range = dates ? ([dates[0], dates[1]] as [dayjs.Dayjs, dayjs.Dayjs]) : null;
    setTimeRange(range);
    fetchList(1, pagination.pageSize, levelFilter, keyword, serviceFilter, range);
  };

  const handleSearch = (value: string) => {
    setKeyword(value);
    fetchList(1, pagination.pageSize, levelFilter, value, serviceFilter, timeRange);
  };

  const serviceOptions = useMemo(
    () => SERVICE_LIST.map((s) => ({ value: s, label: s })),
    []
  );

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
      title: '来源主机',
      dataIndex: 'sourceHost',
      key: 'sourceHost',
      width: 140,
      render: (text: string) => (
        <span style={{ color: '#aaaaaa', fontFamily: '"Roboto Mono", monospace' }}>{text || '-'}</span>
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
        <Space style={{ marginBottom: 16 }} wrap>
          <span style={{ color: '#aaaaaa' }}>级别：</span>
          <Select
            allowClear
            placeholder="全部级别"
            value={levelFilter}
            onChange={handleLevelChange}
            style={{ width: 120 }}
            options={[
              { value: LogItemLevel.INFO, label: 'INFO' },
              { value: LogItemLevel.WARN, label: 'WARN' },
              { value: LogItemLevel.ERROR, label: 'ERROR' },
            ]}
          />
          <span style={{ color: '#aaaaaa' }}>服务：</span>
          <Select
            allowClear
            placeholder="全部服务"
            value={serviceFilter}
            onChange={handleServiceChange}
            style={{ width: 180 }}
            options={serviceOptions}
          />
          <RangePicker
            placeholder={['开始日期', '结束日期']}
            value={timeRange}
            onChange={handleTimeRangeChange}
            format="YYYY-MM-DD"
          />
          <Input.Search
            placeholder="搜索日志内容关键词"
            allowClear
            onSearch={handleSearch}
            style={{ width: 260 }}
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
          expandable={{
            expandedRowRender: (record: LogItemExt) => (
              <div style={{ background: '#141414', padding: '12px 24px' }}>
                <div style={{ marginBottom: 8 }}>
                  <span style={{ color: '#aaaaaa' }}>来源主机：</span>
                  <span style={{ color: '#ffffff', fontFamily: '"Roboto Mono", monospace' }}>
                    {record.sourceHost || '-'}
                  </span>
                </div>
                <div style={{ marginBottom: 8 }}>
                  <span style={{ color: '#aaaaaa' }}>日志路径：</span>
                  <span style={{ color: '#ffffff', fontFamily: '"Roboto Mono", monospace' }}>
                    {record.logPath || '-'}
                  </span>
                </div>
                <div style={{ marginBottom: 8 }}>
                  <span style={{ color: '#aaaaaa' }}>Trace ID：</span>
                  <span style={{ color: '#ffffff', fontFamily: '"Roboto Mono", monospace' }}>
                    {record.traceId || '-'}
                  </span>
                </div>
                <div>
                  <span style={{ color: '#aaaaaa' }}>完整内容：</span>
                  <span style={{ color: '#cccccc' }}>{record.content}</span>
                </div>
              </div>
            ),
            rowExpandable: () => true,
          }}
        />
      </Card>
    </div>
  );
}
