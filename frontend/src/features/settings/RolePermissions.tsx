import { useEffect, useState, useCallback } from 'react';
import { Card, Button, message, Space, Tag, Select, Checkbox, Alert, Empty, Modal, Form, Input, Popconfirm } from 'antd';
import { SaveOutlined, ReloadOutlined, LockOutlined, PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { getDevOpsDashboardAPI } from '../../api/client';
import type { RoleInfo, PermissionGroup, CreateRoleRequest } from '../../api/model';
import { permissionLabel } from './permissionLabels';

const api = getDevOpsDashboardAPI(); // 模块级单例

// RolePermissions 角色权限配置面板 — RBAC 二期
// 角色 × 权限点矩阵：选择角色 → 按分组勾选权限 → 保存热生效（无需重新登录）
export default function RolePermissions() {
  const [roles, setRoles] = useState<RoleInfo[]>([]);
  const [groups, setGroups] = useState<PermissionGroup[]>([]);
  const [selectedRole, setSelectedRole] = useState<string | undefined>(undefined);
  const [checkedPerms, setCheckedPerms] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [rolesRes, groupsRes] = await Promise.all([
        api.listRoles(),
        api.listPermissionGroups(),
      ]);
      const roleList = rolesRes.data.roles;
      setRoles(roleList);
      setGroups(groupsRes.data.groups);
      // 默认选中第一个非锁定角色（admin 锁定，选一个可编辑的）
      const editable = roleList.find((r) => !r.locked);
      const target = editable ?? roleList[0];
      setSelectedRole(target?.name);
      setCheckedPerms(new Set(target?.permissions ?? []));
    } catch {
      message.error('加载角色权限失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  // 切换角色时载入该角色权限点
  const handleRoleChange = (roleName: string) => {
    setSelectedRole(roleName);
    const role = roles.find((r) => r.name === roleName);
    setCheckedPerms(new Set(role?.permissions ?? []));
  };

  const currentRole = roles.find((r) => r.name === selectedRole);
  const locked = currentRole?.locked ?? false;

  // 切换单个权限点（含前置依赖联动）：
  //   勾选依赖点(如 webhook:update) → 自动补上被依赖点(webhook:read)
  //   取消被依赖点(webhook:read) → 自动取消所有依赖它的点(webhook:update)
  const togglePerm = (perm: string, group: PermissionGroup) => {
    const requires = group.requires ?? {};
    setCheckedPerms((prev) => {
      const next = new Set(prev);
      if (next.has(perm)) {
        next.delete(perm);
        // 若取消的是被依赖点，级联取消依赖它的权限点
        Object.entries(requires).forEach(([dependent, base]) => {
          if (base === perm && next.has(dependent)) {
            next.delete(dependent);
          }
        });
      } else {
        next.add(perm);
        // 勾选依赖点时自动补上被依赖点（后端也会兜底，此处为交互即时反馈）
        if (requires[perm]) {
          next.add(requires[perm]);
        }
      }
      return next;
    });
  };

  const toggleGroup = (group: PermissionGroup) => {
    const requires = group.requires ?? {};
    setCheckedPerms((prev) => {
      const next = new Set(prev);
      const allChecked = group.permissions.every((p) => next.has(p));
      if (allChecked) {
        // 整组取消：先删组内所有，再级联删依赖这些点的权限
        group.permissions.forEach((p) => {
          next.delete(p);
          Object.entries(requires).forEach(([dependent, base]) => {
            if (base === p && next.has(dependent)) {
              next.delete(dependent);
            }
          });
        });
      } else {
        // 整组勾选：勾上全部 + 自动补全组内依赖链
        group.permissions.forEach((p) => {
          next.add(p);
          if (requires[p]) {
            next.add(requires[p]);
          }
        });
      }
      return next;
    });
  };

  const handleSave = async () => {
    if (!selectedRole) return;
    setSaving(true);
    try {
      await api.updateRolePermissions(selectedRole, {
        permissions: Array.from(checkedPerms).sort(),
      });
      message.success('权限已更新，即时生效');
      await load(); // 刷新角色列表（回读最新权限）
    } catch (err: any) {
      message.error(err?.response?.data?.error || '保存失败');
    } finally {
      setSaving(false);
    }
  };

  // ---- 角色管理（新建/编辑/删除自定义角色）----
  const [roleModalOpen, setRoleModalOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<RoleInfo | null>(null);
  const [roleForm] = Form.useForm<CreateRoleRequest>();

  const openCreateRole = () => {
    setEditingRole(null);
    roleForm.resetFields();
    setRoleModalOpen(true);
  };

  const openEditRole = () => {
    if (!currentRole || currentRole.builtin) return;
    setEditingRole(currentRole);
    roleForm.setFieldsValue({
      name: currentRole.name,
      label: currentRole.label,
      description: currentRole.description,
    });
    setRoleModalOpen(true);
  };

  const handleRoleModalOk = async () => {
    try {
      const values = await roleForm.validateFields();
      if (editingRole) {
        await api.updateRoleMeta(editingRole.name, {
          label: values.label,
          description: values.description,
        });
        message.success('角色已更新');
      } else {
        await api.createRole({
          name: values.name,
          label: values.label,
          description: values.description,
        });
        message.success('角色已创建');
      }
      setRoleModalOpen(false);
      await load();
    } catch (err: any) {
      if (err?.errorFields) return; // 表单校验失败
      message.error(err?.response?.data?.error || '操作失败');
    }
  };

  const handleDeleteRole = async () => {
    if (!currentRole) return;
    try {
      await api.deleteRole(currentRole.name);
      message.success('角色已删除');
      await load();
    } catch (err: any) {
      message.error(err?.response?.data?.error || '删除失败');
    }
  };

  return (
    <Card
      title={<span style={{ color: '#fff', fontSize: 16 }}>角色权限配置</span>}
      style={{ background: '#1f1f1f', border: '1px solid #333', maxWidth: 860 }}
      loading={loading}
      extra={
        <Button icon={<ReloadOutlined />} onClick={load} size="small">
          刷新
        </Button>
      }
    >
      {/* 角色选择 */}
      <Space style={{ marginBottom: 16 }}>
        <span style={{ color: '#bbb' }}>角色</span>
        <Select
          value={selectedRole}
          style={{ width: 160 }}
          onChange={handleRoleChange}
          options={roles.map((r) => ({
            value: r.name,
            label: (
              <Space size={6}>
                {r.label}
                <span style={{ color: '#888', fontSize: 12 }}>({r.name})</span>
                {r.locked && <LockOutlined style={{ color: '#d89614' }} />}
              </Space>
            ),
          }))}
        />
        {currentRole && (
          <Tag color={currentRole.locked ? 'gold' : 'blue'}>{currentRole.description}</Tag>
        )}
        <Button
          type="dashed"
          size="small"
          icon={<PlusOutlined />}
          onClick={openCreateRole}
        >
          新建角色
        </Button>
        <Button
          size="small"
          icon={<EditOutlined />}
          disabled={!currentRole || currentRole.builtin}
          onClick={openEditRole}
        >
          编辑
        </Button>
        <Popconfirm
          title="删除角色"
          description={currentRole ? `确定删除角色「${currentRole.label}」？其下用户需先转移。` : ''}
          disabled={!currentRole || currentRole.builtin}
          onConfirm={handleDeleteRole}
        >
          <Button
            size="small"
            danger
            icon={<DeleteOutlined />}
            disabled={!currentRole || currentRole.builtin}
          >
            删除
          </Button>
        </Popconfirm>
      </Space>

      {/* 角色新建/编辑弹窗 */}
      <Modal
        title={editingRole ? '编辑角色' : '新建角色'}
        open={roleModalOpen}
        onOk={handleRoleModalOk}
        onCancel={() => setRoleModalOpen(false)}
        okText={editingRole ? '保存' : '创建'}
        cancelText="取消"
      >
        <Form form={roleForm} layout="vertical">
          <Form.Item
            label="角色标识"
            name="name"
            rules={[
              { required: true, message: '请输入角色标识' },
              { pattern: /^[a-z0-9]+(-[a-z0-9]+)*$/, message: '小写字母/数字/连字符，2-32 位' },
            ]}
          >
            <Input placeholder="如 auditor" disabled={!!editingRole} />
          </Form.Item>
          <Form.Item label="显示名" name="label" rules={[{ required: true, message: '请输入显示名' }]}>
            <Input placeholder="如 审计员" />
          </Form.Item>
          <Form.Item label="说明" name="description">
            <Input placeholder="角色用途说明（可选）" />
          </Form.Item>
        </Form>
      </Modal>

      {locked ? (
        <Alert
          type="warning"
          showIcon
          message="admin 为通配策略（拥有全部权限），不可修改。如确需调整，请联系开发者修改后端 seed 配置。"
          style={{ marginBottom: 16, background: '#2a2018', border: '1px solid #4a3a28' }}
        />
      ) : (
        <Alert
          type="info"
          showIcon
          message="勾选保存后立即生效，无需重新登录。被移除权限的角色，其用户访问相应接口将返回 403。"
          style={{ marginBottom: 16, background: '#1a1a2a', border: '1px solid #333' }}
        />
      )}

      {/* 权限点分组矩阵 */}
      {groups.length === 0 ? (
        <Empty description="暂无权限点" />
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {groups.map((group) => {
            const checkedCount = group.permissions.filter((p) => checkedPerms.has(p)).length;
            const allChecked = checkedCount === group.permissions.length && group.permissions.length > 0;
            const someChecked = checkedCount > 0 && !allChecked;
            return (
              <div
                key={group.obj}
                style={{
                  border: '1px solid #333',
                  borderRadius: 6,
                  padding: '10px 14px',
                  background: '#1a1a22',
                  opacity: locked ? 0.6 : 1,
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', marginBottom: 8 }}>
                  <Checkbox
                    checked={allChecked}
                    indeterminate={someChecked}
                    disabled={locked}
                    onChange={() => toggleGroup(group)}
                  >
                    <span style={{ color: '#fff', fontWeight: 600 }}>{group.label}</span>
                  </Checkbox>
                </div>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                  {group.permissions.map((perm) => (
                    <Checkbox
                      key={perm}
                      checked={checkedPerms.has(perm)}
                      disabled={locked}
                      onChange={() => togglePerm(perm, group)}
                      style={{ marginRight: 12, color: '#ccc' }}
                    >
                      {permissionLabel(perm)}
                      <span style={{ color: '#666', fontSize: 12, marginLeft: 4 }}>({perm})</span>
                    </Checkbox>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      )}

      <div style={{ marginTop: 20 }}>
        <Button
          type="primary"
          icon={<SaveOutlined />}
          loading={saving}
          disabled={locked || !selectedRole}
          onClick={handleSave}
        >
          保存权限配置
        </Button>
      </div>
    </Card>
  );
}
