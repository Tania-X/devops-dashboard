import { Layout } from 'antd';
import { Outlet } from 'react-router-dom';
import AppHeader from './AppHeader';
import AppSidebar from './AppSidebar';
import { colors, spacing } from '../../theme/tokens';

const { Content } = Layout;

export default function AppLayout() {
  return (
    <Layout style={{ height: '100vh', overflow: 'hidden', background: colors.bg.page }}>
      <AppSidebar />
      <Layout style={{ background: colors.bg.page, height: '100vh', overflow: 'hidden' }}>
        <AppHeader />
        <Content style={{ padding: spacing.page, background: colors.bg.page, overflow: 'auto' }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
