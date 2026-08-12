// 权限点中文标签映射 — 单一事实源（前端展示层）。
// 权限点定义以 spec/v1-api.yaml + 后端 authz 包为准，此处仅维护中文展示名。
// 新增权限点时：同步后端 authz 常量 + 本映射 + 相关按钮。
export const PERMISSION_LABELS: Record<string, string> = {
  'dashboard:view': '查看概览',
  'server:read': '查看服务器',
  'log:read': '查询日志',
  'deployment:read': '查看部署',
  'monitor:read': '查看监控',
  'agent:read': '查看 Agent',
  'agent:create': '新增 Agent',
  'agent:update': '编辑 Agent',
  'agent:delete': '删除 Agent',
  'agent:deploy': '部署 Agent',
  'agent:stop': '停止 Agent',
  'user:read': '查看用户',
  'user:create': '新增用户',
  'user:update': '编辑用户',
  'user:delete': '删除用户',
  'webhook:read': '查看 Webhook',
  'webhook:update': '配置 Webhook',
  'webhook:test': '测试推送',
  'settings:manage': '权限配置',
};

// 未知权限点兜底：显示原始字符串
export function permissionLabel(perm: string): string {
  return PERMISSION_LABELS[perm] || perm;
}
