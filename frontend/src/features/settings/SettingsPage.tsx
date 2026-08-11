import { useEffect, useState } from 'react';
import { Card, Switch, Radio, Input, Button, Form, message, Space, Alert, Typography } from 'antd';
import { SendOutlined, SaveOutlined } from '@ant-design/icons';
import { getDevOpsDashboardAPI } from '../../api/client';
import type { WebhookConfigUpdate } from '../../api/model';

const { Text } = Typography;

const api = getDevOpsDashboardAPI(); // 模块级单例，避免每次渲染重建

// SettingsPage 系统设置页 — 告警 Webhook 推送配置（企业微信/钉钉机器人）
export default function SettingsPage() {
  const [form] = Form.useForm<WebhookConfigUpdate>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [enabled, setEnabled] = useState(false);
  // enabled 双向绑定：Switch 变化时同步 state 与表单（保存时统一从表单取，避免竞态）
  const syncEnabled = (checked: boolean) => {
    setEnabled(checked);
    form.setFieldValue('enabled', checked);
  };

  // 加载当前配置并回填表单
  const loadConfig = async () => {
    setLoading(true);
    try {
      const { data } = await api.getWebhookConfig();
      form.setFieldsValue({
        enabled: data.enabled,
        kind: data.kind,
        url: data.url,
      });
      setEnabled(data.enabled);
    } catch {
      message.error('加载配置失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConfig();
  }, []);

  // 保存配置（PUT，热生效）
  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      setSaving(true);
      // secret 留空表示不修改（后端处理）；enabled 从表单读取（state 仅作 Switch 受控显示）
      await api.updateWebhookConfig({
        enabled: values.enabled ?? false,
        kind: values.kind,
        url: values.url,
        secret: values.secret || undefined,
      });
      message.success('配置已保存');
    } catch {
      // validateFields 失败或请求失败
      message.error('保存失败，请检查配置');
    } finally {
      setSaving(false);
    }
  };

  // 测试推送（POST /test）
  const handleTest = async () => {
    setTesting(true);
    try {
      const { data } = await api.testWebhookConfig();
      message.success(data.detail || '测试消息已发送，请检查群聊');
    } catch {
      message.error('测试推送失败，请检查 Webhook 地址');
    } finally {
      setTesting(false);
    }
  };

  return (
    <div style={{ padding: 24, background: '#111217', minHeight: '100%' }}>
      <Card
        title={<span style={{ color: '#fff', fontSize: 16 }}>告警通知设置</span>}
        style={{ background: '#1f1f1f', border: '1px solid #333', maxWidth: 720 }}
        loading={loading}
      >
        <Form
          form={form}
          layout="vertical"
          autoComplete="off"
          style={{ color: '#fff' }}
        >
          <Form.Item label={<Text style={{ color: '#fff' }}>启用推送通知</Text>} name="enabled">
            <Switch
              checked={enabled}
              onChange={syncEnabled}
              checkedChildren="开"
              unCheckedChildren="关"
            />
          </Form.Item>

          <Form.Item
            label={<Text style={{ color: '#fff' }}>渠道类型</Text>}
            name="kind"
            rules={[{ required: true, message: '请选择渠道类型' }]}
          >
            <Radio.Group
              onChange={(e) => form.setFieldValue('kind', e.target.value)}
            >
              <Radio value="wecom" style={{ color: '#fff' }}>
                企业微信 WeCom
              </Radio>
              <Radio value="dingtalk" style={{ color: '#fff' }}>
                钉钉 DingTalk
              </Radio>
            </Radio.Group>
          </Form.Item>

          <Form.Item
            label={<Text style={{ color: '#fff' }}>Webhook 地址</Text>}
            name="url"
            rules={[
              { required: true, message: '请输入 Webhook 地址' },
              {
                pattern: /^https?:\/\//,
                message: '必须以 http:// 或 https:// 开头',
              },
            ]}
          >
            <Input
              placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"
              style={{ background: '#111217', color: '#fff' }}
            />
          </Form.Item>

          <Form.Item
            label={<Text style={{ color: '#fff' }}>加签密钥（可选）</Text>}
            name="secret"
            tooltip="钉钉机器人安全设置中的加签密钥；企业微信机器人无需填写"
          >
            <Input.Password
              placeholder="仅钉钉渠道需要，留空表示不修改"
              style={{ background: '#111217', color: '#fff' }}
            />
          </Form.Item>

          {enabled && (
            <Alert
              type="info"
              showIcon
              message="保存后立即生效，无需重启服务。告警产生时（CPU/内存/磁盘超阈值）会自动推送到该地址。"
              style={{ marginBottom: 24, background: '#1a1a2a', border: '1px solid #333' }}
            />
          )}

          <Space>
            <Button
              type="primary"
              icon={<SaveOutlined />}
              loading={saving}
              onClick={handleSave}
            >
              保存
            </Button>
            <Button
              icon={<SendOutlined />}
              loading={testing}
              disabled={!enabled}
              onClick={handleTest}
            >
              测试推送
            </Button>
          </Space>
        </Form>
      </Card>
    </div>
  );
}
