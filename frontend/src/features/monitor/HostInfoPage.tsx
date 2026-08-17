import { useEffect, useState } from 'react';
import { Card, Row, Col, Descriptions, Spin } from 'antd';
import {
  CloudServerOutlined,
  HddOutlined,
  ApartmentOutlined,
  CalendarOutlined,
  FieldTimeOutlined,
  TagOutlined,
  BarcodeOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import type { HostInfo } from '../../api/model';
import { getDevOpsDashboardAPI } from '../../api/client';
import PageHeader from '../../components/PageHeader';
import { colors, fonts, radius } from '../../theme/tokens';

function InfoCard({
  icon,
  title,
  children,
}: {
  icon: React.ReactNode;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <Card
      style={{
        background: colors.bg.panel,
        border: 'none',
        borderRadius: radius.panel,
        height: '100%',
      }}
      styles={{ body: { padding: 20 } }}
    >
      <div style={{ display: 'flex', alignItems: 'center', marginBottom: 16 }}>
        <span style={{ color: colors.brand.primary, fontSize: 20, marginRight: 10 }}>{icon}</span>
        <span style={{ color: colors.text.primary, fontSize: 15, fontWeight: 500 }}>{title}</span>
      </div>
      {children}
    </Card>
  );
}

export default function HostInfoPage() {
  const [hostInfo, setHostInfo] = useState<HostInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  const fetchInfo = () => {
    setLoading(true);
    setError(false);
    getDevOpsDashboardAPI()
      .getHostInfo()
      .then((res) => {
        setHostInfo(res.data);
      })
      .catch(() => setError(true))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchInfo();
  }, []);

  if (loading) {
    return (
      <div style={{ width: '100%' }}>
        <PageHeader title="主机信息" />
        <div style={{ display: 'flex', justifyContent: 'center', padding: 60 }}>
          <Spin />
        </div>
      </div>
    );
  }

  if (error || !hostInfo) {
    return (
      <div style={{ width: '100%' }}>
        <PageHeader title="主机信息" />
        <Card
          style={{
            background: colors.bg.panel,
            border: 'none',
            borderRadius: radius.panel,
            textAlign: 'center',
            padding: 40,
          }}
        >
          <div style={{ color: colors.status.critical, fontSize: fonts.size.h2, marginBottom: 16 }}>
            获取主机信息失败
          </div>
          <div style={{ color: colors.brand.primary, cursor: 'pointer' }} onClick={fetchInfo}>
            点击重试
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div style={{ width: '100%' }}>
      <PageHeader title="主机信息" />

      <Row gutter={[16, 16]}>
        <Col span={12}>
          <InfoCard icon={<CloudServerOutlined />} title="系统信息">
            <Descriptions
              column={1}
              styles={{
                label: { color: colors.text.secondary, paddingBottom: 8 },
                content: { color: colors.text.primary },
              }}
              items={[
                { key: '1', label: (
                  <span><TagOutlined style={{ marginRight: 4 }} />主机名</span>
                ), children: hostInfo.hostname },
                { key: '2', label: (
                  <span><InfoCircleOutlined style={{ marginRight: 4 }} />操作系统</span>
                ), children: `${hostInfo.os} / ${hostInfo.platform}` },
                { key: '3', label: (
                  <span><BarcodeOutlined style={{ marginRight: 4 }} />平台版本</span>
                ), children: hostInfo.platformVersion },
                { key: '4', label: (
                  <span><BarcodeOutlined style={{ marginRight: 4 }} />内核版本</span>
                ), children: hostInfo.kernelVersion },
                { key: '5', label: (
                  <span><ApartmentOutlined style={{ marginRight: 4 }} />架构</span>
                ), children: hostInfo.arch },
              ]}
            />
          </InfoCard>
        </Col>

        <Col span={12}>
          <InfoCard icon={<HddOutlined />} title="硬件信息">
            <Descriptions
              column={1}
              styles={{
                label: { color: colors.text.secondary, paddingBottom: 8 },
                content: { color: colors.text.primary },
              }}
              items={[
                { key: '1', label: (
                  <span><InfoCircleOutlined style={{ marginRight: 4 }} />CPU 型号</span>
                ), children: hostInfo.cpuModel },
                { key: '2', label: (
                  <span><InfoCircleOutlined style={{ marginRight: 4 }} />CPU 核心</span>
                ), children: `${hostInfo.cpuCores} 物理 / ${hostInfo.cpuLogicalCores ?? '-'} 逻辑` },
                { key: '3', label: (
                  <span><InfoCircleOutlined style={{ marginRight: 4 }} />总内存</span>
                ), children: `${hostInfo.totalMemoryGb.toFixed(1)} GB` },
                { key: '4', label: (
                  <span><InfoCircleOutlined style={{ marginRight: 4 }} />交换分区</span>
                ), children: hostInfo.virtualMemoryGb != null ? `${hostInfo.virtualMemoryGb.toFixed(1)} GB` : '-' },
              ]}
            />
          </InfoCard>
        </Col>

        <Col span={12}>
          <InfoCard icon={<FieldTimeOutlined />} title="运行状态">
            <Descriptions
              column={1}
              styles={{
                label: { color: colors.text.secondary, paddingBottom: 8 },
                content: { color: colors.text.primary },
              }}
              items={[
                { key: '1', label: (
                  <span><FieldTimeOutlined style={{ marginRight: 4 }} />运行时长</span>
                ), children: hostInfo.uptime },
                { key: '2', label: (
                  <span><CalendarOutlined style={{ marginRight: 4 }} />启动时间</span>
                ), children: hostInfo.bootTime },
              ]}
            />
          </InfoCard>
        </Col>
      </Row>
    </div>
  );
}
