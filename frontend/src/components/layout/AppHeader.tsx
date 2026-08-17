import { Layout, Breadcrumb, Button, Space } from 'antd';
import { UserOutlined, LogoutOutlined } from '@ant-design/icons';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { getDevOpsDashboardAPI } from '../../api/client';
import { colors, layout } from '../../theme/tokens';

const { Header } = Layout;

const breadcrumbMap: Record<string, string> = {
  '/': '系统概览',
  '/servers': '服务器管理',
  '/logs': '日志查询',
  '/deployments': '部署状态',
  '/monitor': '实时监控',
  '/agents': 'Agent 管理',
  '/users': '用户管理',
  '/settings': '系统设置',
  '/alerts': '告警历史',
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
        height: layout.headerHeight,
        padding: '0 24px',
        background: colors.bg.header,
        borderBottom: `1px solid ${colors.border}`,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        flexShrink: 0,
      }}
    >
      <Breadcrumb
        items={[{ title: 'DevOps' }, { title: title }]}
        style={{ color: colors.text.secondary }}
      />
      <Space size={8}>
        <UserOutlined style={{ color: colors.text.secondary }} />
        <span style={{ color: colors.text.secondary, fontSize: 14 }}>{user?.username || 'Admin'}</span>
        <Button
          type="text"
          icon={<LogoutOutlined />}
          onClick={handleLogout}
          style={{ color: colors.text.secondary }}
        >
          登出
        </Button>
      </Space>
    </Header>
  );
}
