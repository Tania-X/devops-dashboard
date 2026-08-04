import { Layout, Breadcrumb, Button, Space } from 'antd';
import { UserOutlined, LogoutOutlined } from '@ant-design/icons';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { getDevOpsDashboardAPI } from '../../api/client';

const { Header } = Layout;

const breadcrumbMap: Record<string, string> = {
  '/': '系统概览',
  '/servers': '服务器管理',
  '/logs': '日志查询',
  '/deployments': '部署状态',
  '/monitor': '实时监控',
  '/agents': 'Agent 管理',
  '/users': '用户管理',
};

export default function AppHeader() {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const title = breadcrumbMap[location.pathname] || '系统概览';

  const handleLogout = async () => {
    try {
      const api = getDevOpsDashboardAPI();
      await api.logout();
    } catch {
      // ignore logout API errors
    }
    logout();
    navigate('/login');
  };

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
      <Space>
        <UserOutlined style={{ color: '#aaaaaa' }} />
        <span style={{ color: '#aaaaaa', fontSize: 14 }}>{user?.username || 'Admin'}</span>
        <Button
          type="text"
          icon={<LogoutOutlined />}
          onClick={handleLogout}
          style={{ color: '#aaaaaa' }}
        >
          登出
        </Button>
      </Space>
    </Header>
  );
}
