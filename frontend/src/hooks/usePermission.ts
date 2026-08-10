import { useAuth } from '../contexts/AuthContext';

// usePermission 检查当前登录用户是否拥有指定权限点(如 "user:delete")。
// 供按钮级权限控制使用(UX 层);接口安全仍由后端 RequirePermission 校验。
export function usePermission(perm: string): boolean {
  const { permissions } = useAuth();
  return permissions.includes(perm);
}
