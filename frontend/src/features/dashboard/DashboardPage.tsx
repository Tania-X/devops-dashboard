import { Row, Col } from 'antd';
import { dashboardConfig } from './dashboard-config';
import StatPanel from './StatPanel';
import ChartPanel from './ChartPanel';
import AlertListPanel from './AlertListPanel';
import QuickLinksPanel from './QuickLinksPanel';
import type { PanelConfig } from './dashboard-config';

function renderPanel(config: PanelConfig) {
  switch (config.type) {
    case 'stat':
      return <StatPanel config={config} />;
    case 'chart':
      return <ChartPanel config={config} />;
    case 'alert-list':
      return <AlertListPanel config={config} />;
    case 'quick-links':
      return <QuickLinksPanel config={config} />;
    default:
      return null;
  }
}

export default function DashboardPage() {
  return (
    <div style={{ width: '100%' }}>
      <h1 style={{ color: '#ffffff', fontSize: 20, fontWeight: 600, marginBottom: 24 }}>
        {dashboardConfig.title}
      </h1>
      {dashboardConfig.rows.map((row, rowIndex) => (
        <Row key={rowIndex} gutter={[16, 16]} style={{ marginBottom: 16 }}>
          {row.panels.map((panel, panelIndex) => (
            <Col key={panelIndex} span={panel.colSpan}>
              {renderPanel(panel.config)}
            </Col>
          ))}
        </Row>
      ))}
    </div>
  );
}
