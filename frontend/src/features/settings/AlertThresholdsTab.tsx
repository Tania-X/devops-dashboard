import { useEffect, useState } from 'react';
import { Card, InputNumber, Button, Space, message, Typography } from 'antd';
import { SaveOutlined } from '@ant-design/icons';
import { getDevOpsDashboardAPI } from '../../api/client';
import type { AlertThreshold, AlertThresholdMetric } from '../../api/model';
import { usePermission } from '../../hooks/usePermission';
import { colors, fonts, radius } from '../../theme/tokens';

const { Text } = Typography;

const api = getDevOpsDashboardAPI(); // 模块级单例，避免每次渲染重建

// 指标中文名与默认顺序
const METRIC_LABELS: Record<AlertThresholdMetric, string> = {
  cpu: 'CPU 使用率',
  memory: '内存使用率',
  disk: '磁盘使用率',
};

// AlertThresholdsTab 告警阈值配置（CPU/内存/磁盘 的 warning/critical 阈值）
// 权限:settings:manage=可编辑保存;无权限仅展示
export default function AlertThresholdsTab() {
  const [data, setData] = useState<AlertThreshold[]>([]);
  const [loading, setLoading] = useState(false);
  const [savingKey, setSavingKey] = useState<string | null>(null);
  const canManage = usePermission('settings:manage');

  const load = async () => {
    setLoading(true);
    try {
      const { data: list } = await api.getAlertThresholds();
      setData(list);
    } catch {
      message.error('加载阈值配置失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  // 单指标行内更新(后端按 metric 单点保存,热生效)
  const updateField = (metric: AlertThresholdMetric, field: 'warnThreshold' | 'critThreshold', value: number | null) => {
    setData((prev) =>
      prev.map((t) => (t.metric === metric ? { ...t, [field]: value ?? 0 } : t))
    );
  };

  const save = async (metric: AlertThresholdMetric) => {
    const t = data.find((d) => d.metric === metric);
    if (!t) return;
    if (t.warnThreshold <= 0 || t.critThreshold <= 0) {
      message.error('阈值必须大于 0');
      return;
    }
    if (t.warnThreshold >= t.critThreshold) {
      message.error('warning 阈值必须小于 critical 阈值');
      return;
    }
    setSavingKey(metric);
    try {
      await api.updateAlertThreshold({
        metric,
        warnThreshold: t.warnThreshold,
        critThreshold: t.critThreshold,
      });
      message.success('阈值已保存，立即生效');
    } catch {
      message.error('保存失败');
    } finally {
      setSavingKey(null);
    }
  };

  return (
    <Card
      title={<span style={{ color: colors.text.primary, fontSize: 16 }}>告警阈值设置</span>}
      style={{ background: colors.bg.panel, border: `1px solid ${colors.border}`, borderRadius: radius.panel, maxWidth: 720 }}
      loading={loading}
    >
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        {data.map((t) => (
          <Space key={t.metric} size={12} align="center" style={{ justifyContent: 'space-between', width: '100%' }}>
            <Text style={{ color: colors.text.primary, width: 110, display: 'inline-block' }}>
              {METRIC_LABELS[t.metric]}
            </Text>
            <Text style={{ color: colors.text.muted }}>warning(黄) 超过</Text>
            <InputNumber
              min={1}
              max={100}
              value={t.warnThreshold}
              disabled={!canManage}
              onChange={(v) => updateField(t.metric, 'warnThreshold', v)}
              style={{ width: 90, background: colors.bg.header, color: colors.text.primary }}
            />
            <Text style={{ color: colors.text.muted }}>%</Text>
            <Text style={{ color: colors.text.muted }}>critical(红) 超过</Text>
            <InputNumber
              min={1}
              max={100}
              value={t.critThreshold}
              disabled={!canManage}
              onChange={(v) => updateField(t.metric, 'critThreshold', v)}
              style={{ width: 90, background: colors.bg.header, color: colors.text.primary }}
            />
            <Text style={{ color: colors.text.muted }}>%</Text>
            {canManage && (
              <Button
                type="primary"
                size="small"
                icon={<SaveOutlined />}
                loading={savingKey === t.metric}
                onClick={() => save(t.metric)}
              >
                保存
              </Button>
            )}
          </Space>
        ))}
        <Text style={{ color: colors.text.muted, fontSize: fonts.size.caption }}>
          保存后立即生效，无需重启。告警产生时（使用率超阈值）会推送到「告警通知」配置的 Webhook。
        </Text>
      </Space>
    </Card>
  );
}
