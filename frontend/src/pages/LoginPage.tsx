import { useState } from 'react';
import { Form, Input, Button, Card, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import AppLogo from '../components/AppLogo';
import { colors, radius } from '../theme/tokens';

export default function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      await login(values.username, values.password);
      message.success('登录成功');
      navigate('/', { replace: true });
    } catch (err: any) {
      message.error(err.message || '登录失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        background: `radial-gradient(ellipse 80% 60% at 50% -10%, rgba(23, 125, 220, 0.16), transparent),
                     radial-gradient(ellipse 60% 50% at 90% 110%, rgba(63, 182, 217, 0.10), transparent),
                     ${colors.bg.page}`,
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      {/* 顶部品牌区 */}
      <div style={{ marginBottom: 32 }}>
        <AppLogo size={44} />
      </div>

      <Card
        style={{
          width: 400,
          background: colors.bg.panel,
          border: `1px solid ${colors.border}`,
          borderRadius: radius.panel + 4,
          boxShadow: '0 8px 32px rgba(0, 0, 0, 0.45)',
        }}
        styles={{ body: { padding: 32 } }}
      >
        <h1
          style={{
            color: colors.text.primary,
            fontSize: 22,
            fontWeight: 600,
            textAlign: 'center',
            marginBottom: 8,
          }}
        >
          运维监控仪表盘
        </h1>
        <p
          style={{
            color: colors.text.muted,
            fontSize: 13,
            textAlign: 'center',
            marginBottom: 28,
          }}
        >
          服务器 · 日志 · 部署 · 告警 一站式管理
        </p>
        <Form onFinish={handleSubmit} size="large" autoComplete="off">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input prefix={<UserOutlined style={{ color: colors.text.muted }} />} placeholder="用户名" autoComplete="off" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password
              prefix={<LockOutlined style={{ color: colors.text.muted }} />}
              placeholder="密码"
              autoComplete="new-password"
            />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0 }}>
            <Button type="primary" htmlType="submit" loading={loading} block>
              登录
            </Button>
          </Form.Item>
        </Form>
      </Card>

      <div
        style={{
          position: 'absolute',
          bottom: 24,
          color: colors.text.muted,
          fontSize: 12,
        }}
      >
        DevOps Dashboard · Internal Ops Console
      </div>
    </div>
  );
}
