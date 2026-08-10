import { Button, type ButtonProps } from 'antd';
import { usePermission } from '../hooks/usePermission';

interface AuthButtonProps extends ButtonProps {
  /** 所需权限点，如 "user:create"。无权限时按钮不渲染。 */
  perm: string;
}

// AuthButton 权限按钮：无权限时不渲染（按钮级权限控制，UX 层）。
// 注意：这只是前端体验控制，接口安全始终由后端 RequirePermission 强制校验。
export default function AuthButton({ perm, children, ...props }: AuthButtonProps) {
  const can = usePermission(perm);
  if (!can) return null;
  return <Button {...props}>{children}</Button>;
}
