import { Layout, Menu } from 'antd';
import {
  DashboardOutlined,
  CloudServerOutlined,
  FileTextOutlined,
  RocketOutlined,
} from '@ant-design/icons';
import { useLocation, useNavigate } from 'react-router-dom';

const { Sider } = Layout;

const menuItems = [
  {
    key: '/',
    icon: <DashboardOutlined />,
    label: '系统概览',
  },
  {
    key: '/servers',
    icon: <CloudServerOutlined />,
    label: '服务器管理',
  },
  {
    key: '/logs',
    icon: <FileTextOutlined />,
    label: '日志查询',
  },
  {
    key: '/deployments',
    icon: <RocketOutlined />,
    label: '部署状态',
  },
];

export default function AppSidebar() {
  const location = useLocation();
  const navigate = useNavigate();

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
        items={menuItems}
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
