import { useEffect, useRef, useState } from 'react';
import {
  Table,
  Drawer,
  Select,
  Space,
  Card,
  Descriptions,
  Spin,
  Row,
  Col,
  Progress,
  Input,
} from 'antd';
import type { ProcessDetail, ProcessItem } from '../../api/model';
import { getDevOpsDashboardAPI } from '../../api/client';
import StatusTag, { type SemanticLevel } from '../../components/StatusTag';
import PageHeader from '../../components/PageHeader';
import { colors, fonts, radius, spacing } from '../../theme/tokens';

const statusLevelMap: Record<string, SemanticLevel> = {
  running: 'success',
  sleep: 'info',
  idle: 'unknown',
  stop: 'critical',
  zombie: 'warning',
  disk: 'warning',
};

const statusLabelMap: Record<string, string> = {
  running: '运行中',
  sleep: '睡眠',
  idle: '空闲',
  stop: '已停止',
  zombie: '僵尸',
  disk: '磁盘等待',
};

function ProcessStatusTag({ status }: { status: string }) {
  const key = status.toLowerCase();
  return (
    <StatusTag
      level={statusLevelMap[key] || 'unknown'}
      label={statusLabelMap[key] || status}
    />
  );
}

/** 阈值 → 状态色（spec 2.4） */
function thresholdColor(v: number, warn: number, crit: number): string {
  if (v > crit) return colors.status.critical;
  if (v > warn) return colors.status.warning;
  return colors.status.success;
}

export default function ProcessListPage() {
  const [data, setData] = useState<ProcessItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [sortBy, setSortBy] = useState('cpu');
  const [order, setOrder] = useState('desc');
  const [limit, setLimit] = useState<number>(200);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
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
      .catch((err) => {
        console.error('[ProcessList] API 调用失败:', err);
      })
      .finally(() => setLoading(false));
  };

  // 持有最新 fetchList 的引用，避免 setInterval 闭包捕获旧值
  const fetchListRef = useRef(fetchList);
  fetchListRef.current = fetchList;

  useEffect(() => {
    fetchList();
  }, [sortBy, order, keyword, limit]);

  // limit 变化时重置页码
  useEffect(() => {
    setCurrentPage(1);
  }, [limit]);

  // 自动刷新——通过 ref 始终拿到最新的 fetchList
  useEffect(() => {
    const interval = setInterval(() => fetchListRef.current(), 10000);
    return () => clearInterval(interval);
  }, []);

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
      title: '#',
      key: 'index',
      width: 60,
      render: (_: unknown, __: unknown, index: number) => (
        <span style={{ color: colors.text.muted, fontFamily: fonts.mono }}>{index + 1}</span>
      ),
    },
    {
      title: 'PID',
      dataIndex: 'pid',
      key: 'pid',
      width: 80,
      render: (v: number) => (
        <span style={{ color: colors.text.primary, fontFamily: fonts.mono }}>{v}</span>
      ),
    },
    {
      title: '进程名',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => <span style={{ color: colors.text.primary, fontWeight: 500 }}>{text}</span>,
    },
    {
      title: 'CPU %',
      dataIndex: 'cpuPercent',
      key: 'cpuPercent',
      width: 120,
      render: (v: number) => {
        const color = thresholdColor(v, 20, 50);
        return (
          <Space>
            <Progress
              percent={Math.min(v, 100)}
              size="small"
              strokeColor={color}
              railColor={colors.border}
              showInfo={false}
              style={{ width: 60 }}
            />
            <span style={{ color, fontFamily: fonts.mono, fontSize: 12 }}>{v.toFixed(1)}</span>
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
        const color = thresholdColor(v, 10, 30);
        return (
          <Space>
            <Progress
              percent={Math.min(v, 100)}
              size="small"
              strokeColor={color}
              railColor={colors.border}
              showInfo={false}
              style={{ width: 60 }}
            />
            <span style={{ color, fontFamily: fonts.mono, fontSize: 12 }}>{v.toFixed(1)}</span>
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
        <span style={{ color: colors.text.secondary, fontFamily: fonts.mono, fontSize: 12 }}>
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
      <PageHeader title="进程列表" />

      <Card
        style={{
          background: colors.bg.panel,
          border: 'none',
          borderRadius: radius.panel,
        }}
        styles={{ body: { padding: spacing.panel } }}
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
          <span style={{ color: colors.text.secondary }}>排序：</span>
          <Select
            value={sortBy}
            onChange={(v) => setSortBy(v)}
            style={{ width: 100 }}
            options={[
              { value: 'cpu', label: 'CPU' },
              { value: 'memory', label: '内存' },
              { value: 'pid', label: 'PID' },
              { value: 'name', label: '进程名' },
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
          <span style={{ color: colors.text.secondary }}>最多显示：</span>
          <Select
            value={limit}
            onChange={(v) => setLimit(v)}
            style={{ width: 80 }}
            options={[
              { value: 100, label: '100' },
              { value: 200, label: '200' },
              { value: 500, label: '500' },
            ]}
          />
        </Space>

        <Table
          columns={columns as any}
          dataSource={data}
          rowKey="pid"
          loading={loading}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 个进程`,
            pageSizeOptions: ['10', '20', '50', '100'],
            onChange: (page, size) => {
              setCurrentPage(page);
              setPageSize(size);
            },
          }}
          onRow={(record) => ({
            onClick: () => handleRowClick(record),
            style: { cursor: 'pointer' },
          })}
          size="middle"
        />
      </Card>

      <Drawer
        title={
          <span style={{ color: colors.text.primary }}>
            {selectedProcess?.name || '进程详情'} (PID: {selectedProcess?.pid})
          </span>
        }
        open={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        size="large"
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
        ) : selectedProcess ? (
          <div>
            <Descriptions
              title={<span style={{ color: colors.text.primary, fontSize: fonts.size.h2, fontWeight: fonts.weight.h2 }}>基本信息</span>}
              column={2}
              styles={{
                label: { color: colors.text.secondary },
                content: { color: colors.text.primary },
              }}
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
                color: colors.text.primary,
                fontSize: fonts.size.h2,
                fontWeight: fonts.weight.h2,
                marginTop: 32,
                marginBottom: 16,
                borderBottom: `1px solid ${colors.border}`,
                paddingBottom: 8,
              }}
            >
              资源占用
            </h3>
            <Row gutter={24} style={{ marginBottom: 16 }}>
              <Col span={8}>
                <div style={{ color: colors.text.secondary, fontSize: fonts.size.caption, marginBottom: 4 }}>
                  CPU 使用率
                </div>
                <Space>
                  <Progress
                    type="circle"
                    percent={Math.min(Math.round(selectedProcess.cpuPercent * 10) / 10, 100)}
                    size={60}
                    strokeColor={selectedProcess.cpuPercent > 50 ? colors.status.critical : colors.brand.primary}
                    railColor={colors.border}
                    format={(pct) => `${pct?.toFixed(1)}%`}
                  />
                </Space>
              </Col>
              <Col span={8}>
                <div style={{ color: colors.text.secondary, fontSize: fonts.size.caption, marginBottom: 4 }}>
                  内存使用率
                </div>
                <Space>
                  <Progress
                    type="circle"
                    percent={Math.min(Math.round(selectedProcess.memoryPercent * 10) / 10, 100)}
                    size={60}
                    strokeColor={selectedProcess.memoryPercent > 30 ? colors.status.warning : colors.status.success}
                    railColor={colors.border}
                    format={(pct) => `${pct?.toFixed(1)}%`}
                  />
                </Space>
              </Col>
              <Col span={8}>
                <div style={{ color: colors.text.secondary, fontSize: fonts.size.caption, marginBottom: 4 }}>
                  内存占用
                </div>
                <span style={{ color: colors.text.primary, fontFamily: fonts.mono, fontSize: fonts.size.h2 }}>
                  {selectedProcess.memoryMb.toFixed(1)} MB
                </span>
              </Col>
            </Row>

            {selectedProcess.cmdline && (
              <>
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
                  命令行
                </h3>
                <div
                  style={{
                    background: colors.codeBg,
                    border: `1px solid ${colors.border}`,
                    borderRadius: radius.panel,
                    padding: 12,
                    fontFamily: fonts.mono,
                    fontSize: fonts.size.caption,
                    color: colors.codeText,
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
                    color: colors.text.primary,
                    fontSize: fonts.size.h2,
                    fontWeight: fonts.weight.h2,
                    marginTop: 32,
                    marginBottom: 16,
                    borderBottom: `1px solid ${colors.border}`,
                    paddingBottom: 8,
                  }}
                >
                  工作目录
                </h3>
                <div style={{ color: colors.codeText, fontFamily: fonts.mono, fontSize: fonts.size.caption }}>
                  {selectedProcess.workingDir}
                </div>
              </>
            )}

            {selectedProcess.env && selectedProcess.env.length > 0 && (
              <>
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
                  环境变量
                </h3>
                <div
                  style={{
                    background: colors.codeBg,
                    border: `1px solid ${colors.border}`,
                    borderRadius: radius.panel,
                    padding: 12,
                    fontFamily: fonts.mono,
                    fontSize: 11,
                    color: colors.text.muted,
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
