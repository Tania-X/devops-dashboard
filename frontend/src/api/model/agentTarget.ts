export interface AgentTarget {
  id: string;
  name: string;
  host: string;
  port: number;
  username: string;
  authType: string;
  deployDir: string;
  agentPort: number;
  status: 'unknown' | 'online' | 'offline';
  createdAt: string;
  updatedAt: string;
}
