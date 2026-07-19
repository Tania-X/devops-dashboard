import { Layout } from 'antd';
import { Outlet } from 'react-router-dom';
import AppHeader from './AppHeader';
import AppSidebar from './AppSidebar';

const { Content } = Layout;

export default function AppLayout() {
  return (
    <Layout style={{ height: '100vh', overflow: 'hidden', background: '#141414' }}>
      <AppSidebar />
      <Layout style={{ background: '#141414', height: '100vh', overflow: 'hidden' }}>
        <AppHeader />
        <Content style={{ padding: 24, background: '#141414', overflow: 'auto' }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
