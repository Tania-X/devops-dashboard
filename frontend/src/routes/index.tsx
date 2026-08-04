import { Routes, Route } from 'react-router-dom';
import AppLayout from '../components/layout/AppLayout';
import ProtectedRoute from '../components/ProtectedRoute';
import LoginPage from '../pages/LoginPage';
import DashboardPage from '../features/dashboard/DashboardPage';
import ServerListPage from '../features/server/ServerListPage';
import LogQueryPage from '../features/log/LogQueryPage';
import DeploymentPage from '../features/deployment/DeploymentPage';
import MonitorPage from '../features/monitor/MonitorPage';
import AgentPage from '../features/agent/AgentPage';
import UserPage from '../features/user/UserPage';

export default function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<ProtectedRoute />}>
        <Route path="/" element={<AppLayout />}>
          <Route index element={<DashboardPage />} />
          <Route path="servers" element={<ServerListPage />} />
          <Route path="logs" element={<LogQueryPage />} />
          <Route path="deployments" element={<DeploymentPage />} />
          <Route path="monitor" element={<MonitorPage />} />
          <Route path="agents" element={<AgentPage />} />
          <Route path="users" element={<UserPage />} />
        </Route>
      </Route>
    </Routes>
  );
}
