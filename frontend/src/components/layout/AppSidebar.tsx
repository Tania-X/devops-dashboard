import { Layout, Menu } from 'antd';
import {
  DashboardOutlined,
  CloudServerOutlined,
  FileTextOutlined,
  RocketOutlined,
  FundOutlined,
  NodeIndexOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';

const { Sider } = Layout;

// 菜单项带权限点(perm):无权限的菜单不渲染(菜单级权限控制)
const menuItems = [
  {
    key: '/',
    icon: <DashboardOutlined />,
    label: '系统概览',
    perm: 'dashboard:view',
  },
  {
    key: '/servers',
    icon: <CloudServerOutlined />,
    label: '服务器管理',
    perm: 'server:read',
  },
  {
    key: '/logs',
    icon: <FileTextOutlined />,
    label: '日志查询',
    perm: 'log:read',
  },
  {
    key: '/deployments',
    icon: <RocketOutlined />,
    label: '部署状态',
    perm: 'deployment:read',
  },
  {
    key: '/monitor',
    icon: <FundOutlined />,
    label: '实时监控',
    perm: 'monitor:read',
  },
  {
    key: '/agents',
    icon: <NodeIndexOutlined />,
    label: 'Agent 管理',
    perm: 'agent:read',
  },
  {
    key: '/users',
    icon: <UserOutlined />,
    label: '用户管理',
    perm: 'user:read',
  },
];

export default function AppSidebar() {
  const location = useLocation();
  const navigate = useNavigate();
  const { permissions } = useAuth();

  // 按权限点过滤菜单:无权限的菜单项不显示(如 viewer 看不到"用户管理")
  const visibleItems = menuItems.filter((item) => permissions.includes(item.perm));

  return (
    <Sider
      width={200}
      style={{
        background: '#000000',
        overflow: 'auto',
        borderRight: '1px solid #333333',
      }}
    >
      <div
        style={{
          height: 56,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: '#ffffff',
          fontSize: 18,
          fontWeight: 600,
          borderBottom: '1px solid #333333',
        }}
      >
        DevOps
      </div>
      <Menu
        mode="inline"
        selectedKeys={[location.pathname]}
        items={visibleItems}
        onClick={({ key }) => navigate(key)}
        style={{
          background: '#000000',
          borderRight: 'none',
        }}
        theme="dark"
      />
    </Sider>
  );
}
