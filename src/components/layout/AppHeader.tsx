import { Layout, Breadcrumb } from 'antd';
import { useLocation } from 'react-router-dom';

const { Header } = Layout;

const breadcrumbMap: Record<string, string> = {
  '/': '系统概览',
  '/servers': '服务器管理',
  '/logs': '日志查询',
  '/deployments': '部署状态',
};

export default function AppHeader() {
  const location = useLocation();
  const title = breadcrumbMap[location.pathname] || '系统概览';

  return (
    <Header
      style={{
        height: 56,
        padding: '0 24px',
        background: '#111217',
        borderBottom: '1px solid #333333',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        flexShrink: 0,
      }}
    >
      <Breadcrumb
        items={[{ title: 'DevOps' }, { title: title }]}
        style={{ color: '#aaaaaa' }}
      />
      <span style={{ color: '#aaaaaa', fontSize: 14 }}>Admin</span>
    </Header>
  );
}
