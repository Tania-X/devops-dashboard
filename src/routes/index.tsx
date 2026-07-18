import { Routes, Route } from 'react-router-dom';
import AppLayout from '../components/layout/AppLayout';
import DashboardPage from '../features/dashboard/DashboardPage';
import ServerListPage from '../features/server/ServerListPage';
import LogQueryPage from '../features/log/LogQueryPage';
import DeploymentPage from '../features/deployment/DeploymentPage';

export default function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<AppLayout />}>
        <Route index element={<DashboardPage />} />
        <Route path="servers" element={<ServerListPage />} />
        <Route path="logs" element={<LogQueryPage />} />
        <Route path="deployments" element={<DeploymentPage />} />
      </Route>
    </Routes>
  );
}
