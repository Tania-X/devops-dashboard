import { useState } from 'react';
import { Tabs } from 'antd';
import { FundOutlined, InfoCircleOutlined } from '@ant-design/icons';
import ProcessListPage from './ProcessListPage';
import HostInfoPage from './HostInfoPage';
import PageHeader from '../../components/PageHeader';
import { colors } from '../../theme/tokens';

export default function MonitorPage() {
  const [activeKey, setActiveKey] = useState('processes');

  return (
    <div style={{ width: '100%' }}>
      <PageHeader title="实时监控" />

      <Tabs
        activeKey={activeKey}
        onChange={setActiveKey}
        style={{ color: colors.text.primary }}
        items={[
          {
            key: 'processes',
            label: (
              <span>
                <FundOutlined style={{ marginRight: 6 }} />
                进程列表
              </span>
            ),
            children: <ProcessListPage />,
          },
          {
            key: 'host',
            label: (
              <span>
                <InfoCircleOutlined style={{ marginRight: 6 }} />
                主机信息
              </span>
            ),
            children: <HostInfoPage />,
          },
        ]}
      />
    </div>
  );
}
