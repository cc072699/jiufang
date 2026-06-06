import { useState, useEffect } from 'react';
import { Typography, List, Card, Tag, Space, Button, Table, Checkbox, message, Empty, Input, Divider, Modal, Form } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { ReloadOutlined, SaveOutlined, SearchOutlined, UserAddOutlined, PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  getGroupList, configurePermission, getGroupMembers, addGroupMember,
  removeGroupMember, getUserList, createGroup, updateGroup, deleteGroup, getPermissions,
} from '../shared/api';
import { confirmAction } from '../shared/ui';
import type { GroupRecord, PermissionItem, GroupMemberItem } from '../shared/api/types';

const { Title, Text } = Typography;

// Mock table schema for permission config (simulating ERP tables)
const ERP_TABLES = [
  { name: 'sales', label: '销售表', fields: ['id', 'product_name', 'amount', 'date', 'region'] },
  { name: 'inventory', label: '库存表', fields: ['id', 'product_name', 'stock', 'warehouse', 'safety_stock'] },
  { name: 'customers', label: '客户表', fields: ['id', 'name', 'contact', 'region', 'level'] },
  { name: 'products', label: '产品表', fields: ['id', 'name', 'category', 'price', 'status'] },
];

export default function PermissionPage() {
  const queryClient = useQueryClient();
  const [selectedGroup, setSelectedGroup] = useState<GroupRecord | null>(null);
  const [tablePerms, setTablePerms] = useState<Record<string, boolean>>({});
  const [fieldPerms, setFieldPerms] = useState<Record<string, boolean>>({});
  const [filterConditions, setFilterConditions] = useState<Record<string, string>>({});
  const [saving, setSaving] = useState(false);
  const [permissionsLoaded, setPermissionsLoaded] = useState(false);

  // Add member modal state
  const [addModalOpen, setAddModalOpen] = useState(false);
  const [userSearch, setUserSearch] = useState('');
  const [addingMember, setAddingMember] = useState(false);

  // Create group modal state
  const [createGroupOpen, setCreateGroupOpen] = useState(false);
  const [creatingGroup, setCreatingGroup] = useState(false);
  const [groupForm] = Form.useForm();

  // Edit group modal state
  const [editGroupOpen, setEditGroupOpen] = useState(false);
  const [editingGroup, setEditingGroup] = useState(false);
  const [editForm] = Form.useForm();

  // Group name search
  const [groupSearch, setGroupSearch] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['groups'],
    queryFn: () => getGroupList({ page: 1, page_size: 100 }),
  });

  const groups = (data?.groups ?? []).filter(
    (g) => !groupSearch || g.name.includes(groupSearch),
  );

  // Fetch group members when a group is selected
  const {
    data: memberData,
    isLoading: membersLoading,
  } = useQuery({
    queryKey: ['group-members', selectedGroup?.id],
    queryFn: () => getGroupMembers(selectedGroup!.id),
    enabled: !!selectedGroup,
  });

  const members = memberData?.members ?? [];

  const initEmptyPerms = () => {
    const tPerms: Record<string, boolean> = {};
    const fPerms: Record<string, boolean> = {};
    const filters: Record<string, string> = {};
    ERP_TABLES.forEach((t) => {
      tPerms[t.name] = false;
      filters[t.name] = '';
      t.fields.forEach((f) => { fPerms[`${t.name}.${f}`] = false; });
    });
    setTablePerms(tPerms);
    setFieldPerms(fPerms);
    setFilterConditions(filters);
    setPermissionsLoaded(false);
  };

  const handleSelectGroup = (group: GroupRecord) => {
    setSelectedGroup(group);
    initEmptyPerms();
  };

  // Load existing permissions from backend when group is selected
  const loadPermissions = async () => {
    if (!selectedGroup || permissionsLoaded) return;
    try {
      const data = await getPermissions(selectedGroup.id);
      if (data?.permissions?.length) {
        const tPerms: Record<string, boolean> = {};
        const fPerms: Record<string, boolean> = {};
        const filters: Record<string, string> = {};

        // Init all to false
        ERP_TABLES.forEach((t) => {
          tPerms[t.name] = false;
          filters[t.name] = '';
          t.fields.forEach((f) => { fPerms[`${t.name}.${f}`] = false; });
        });

        // Apply loaded permissions
        for (const perm of data.permissions) {
          const tableName = perm.table_name;
          if (tableName in tPerms) {
            tPerms[tableName] = true;
            if (perm.filter_condition) {
              filters[tableName] = perm.filter_condition;
            }
            try {
              const fields: string[] = JSON.parse(perm.allowed_fields || '[]');
              for (const f of fields) {
                const key = `${tableName}.${f}`;
                if (key in fPerms) fPerms[key] = true;
              }
            } catch { /* ignore parse error */ }
          }
        }

        setTablePerms(tPerms);
        setFieldPerms(fPerms);
        setFilterConditions(filters);
      }
    } catch {
      // 权限加载失败不阻断 — 保持空白状态
    }
    setPermissionsLoaded(true);
  };

  const handleSave = () => {
    if (!selectedGroup) return;

    confirmAction({
      title: '权限变更确认',
      content: `权限变更将立即对 ${selectedGroup.member_count} 名成员生效，是否确认？`,
      okText: '确认保存',
      onOk: async () => {
        setSaving(true);
        try {
          const permissions: PermissionItem[] = ERP_TABLES
            .filter((t) => tablePerms[t.name])
            .map((t) => ({
              table_name: t.name,
              allowed_fields: JSON.stringify(
                t.fields.filter((f) => fieldPerms[`${t.name}.${f}`])
              ),
              filter_condition: filterConditions[t.name] || undefined,
            }));
          await configurePermission(selectedGroup.id, { permissions });
          message.success('权限已更新');
          queryClient.invalidateQueries({ queryKey: ['groups'] });
        } catch {
          message.error('保存失败');
        } finally {
          setSaving(false);
        }
      },
    });
  };

  const handleAddMember = async (userId: string) => {
    if (!selectedGroup) return;
    setAddingMember(true);
    try {
      await addGroupMember(selectedGroup.id, userId);
      message.success('成员已添加');
      setAddModalOpen(false);
      setUserSearch('');
      queryClient.invalidateQueries({ queryKey: ['group-members'] });
      queryClient.invalidateQueries({ queryKey: ['groups'] });
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : '添加失败';
      message.error(msg);
    } finally {
      setAddingMember(false);
    }
  };

  const handleCreateGroup = async () => {
    try {
      const values = await groupForm.validateFields();
      setCreatingGroup(true);
      await createGroup({ name: values.name, description: values.description });
      message.success('用户组已创建');
      setCreateGroupOpen(false);
      groupForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['groups'] });
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'errorFields' in err) return;
      message.error(err instanceof Error ? err.message : '创建失败');
    } finally {
      setCreatingGroup(false);
    }
  };

  const handleEditGroup = (group: GroupRecord) => {
    editForm.setFieldsValue({ name: group.name, description: group.description });
    setEditGroupOpen(true);
  };

  const handleEditGroupSubmit = async () => {
    if (!selectedGroup) return;
    try {
      const values = await editForm.validateFields();
      setEditingGroup(true);
      await updateGroup(selectedGroup.id, {
        name: values.name,
        description: values.description,
      });
      message.success('用户组已更新');
      setEditGroupOpen(false);
      queryClient.invalidateQueries({ queryKey: ['groups'] });
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'errorFields' in err) return;
      message.error(err instanceof Error ? err.message : '更新失败');
    } finally {
      setEditingGroup(false);
    }
  };

  const handleDeleteGroup = (group: GroupRecord) => {
    confirmAction({
      title: '删除用户组',
      content: `确定删除用户组「${group.name}」？此操作不可恢复。`,
      onOk: async () => {
        try {
          await deleteGroup(group.id);
          message.success('已删除');
          if (selectedGroup?.id === group.id) {
            setSelectedGroup(null);
            initEmptyPerms();
          }
          queryClient.invalidateQueries({ queryKey: ['groups'] });
        } catch {
          message.error('删除失败');
        }
      },
    });
  };

  const handleRemoveMember = (member: GroupMemberItem) => {
    if (!selectedGroup) return;
    Modal.confirm({
      title: '确认移除',
      content: `确定将成员「${member.username}」移出该组吗？`,
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        try {
          await removeGroupMember(selectedGroup.id, member.id);
          message.success('成员已移除');
          queryClient.invalidateQueries({ queryKey: ['group-members'] });
          queryClient.invalidateQueries({ queryKey: ['groups'] });
        } catch {
          message.error('移除失败');
        }
      },
    });
  };

  const permColumns: ColumnsType<{ key: string; table: string; tableLabel: string; field: string }> = [
    { title: '数据表', dataIndex: 'tableLabel', width: 120 },
    { title: '字段', dataIndex: 'field', width: 150 },
    {
      title: '可见',
      dataIndex: 'key',
      width: 80,
      render: (_: unknown, record: { key: string }) => (
        <Checkbox
          checked={fieldPerms[record.key]}
          onChange={(e) => setFieldPerms((prev) => ({ ...prev, [record.key]: e.target.checked }))}
        />
      ),
    },
  ];

  const fieldRows = ERP_TABLES.flatMap((t) =>
    t.fields.map((f) => ({
      key: `${t.name}.${f}`,
      table: t.name,
      tableLabel: t.label,
      field: f,
    }))
  );

  // User search query for add member modal
  const { data: userListData, isLoading: userListLoading } = useQuery({
    queryKey: ['users-for-member', userSearch],
    queryFn: () => getUserList({ page: 1, page_size: 100, username: userSearch || undefined }),
    enabled: addModalOpen,
  });

  const availableUsers = (userListData?.users ?? []).filter(
    (u) => !members.some((m) => m.id === u.id)
  );

  // Auto-load permissions when group is selected and permissions not yet loaded
  useEffect(() => {
    if (selectedGroup && !permissionsLoaded) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      loadPermissions();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedGroup, permissionsLoaded]);

  return (
    <div style={{ display: 'flex', gap: 16, height: 'calc(100vh - 160px)' }}>
      {/* Left: group list */}
      <Card
        title="用户组"
        size="small"
        style={{ width: 240, flexShrink: 0 }}
        extra={
          <Space>
            <Button size="small" icon={<PlusOutlined />} onClick={() => setCreateGroupOpen(true)} />
            <Button size="small" icon={<ReloadOutlined />} onClick={() => queryClient.invalidateQueries({ queryKey: ['groups'] })} />
          </Space>
        }
      >
        <Input
          placeholder="搜索组名"
          prefix={<SearchOutlined />}
          value={groupSearch}
          onChange={(e) => setGroupSearch(e.target.value)}
          allowClear
          size="small"
          style={{ marginBottom: 8 }}
        />
        <List
          size="small"
          loading={isLoading}
          dataSource={groups}
          renderItem={(group) => (
            <List.Item
              style={{
                cursor: 'pointer',
                background: selectedGroup?.id === group.id ? '#e6f4ff' : undefined,
                borderRadius: 6,
                padding: '8px 12px',
              }}
              onClick={() => handleSelectGroup(group)}
            >
              <div style={{ flex: 1 }}>
                <Text strong>{group.name}</Text>
                <br />
                <Text type="secondary" style={{ fontSize: 12 }}>
                  {group.member_count} 人 · {group.description || '无描述'}
                </Text>
              </div>
              <Space size={0}>
                <Button
                  type="text"
                  size="small"
                  icon={<EditOutlined />}
                  onClick={(e) => { e.stopPropagation(); handleEditGroup(group); }}
                />
                <Button
                  type="text"
                  size="small"
                  danger
                  icon={<DeleteOutlined />}
                  onClick={(e) => { e.stopPropagation(); handleDeleteGroup(group); }}
                />
              </Space>
            </List.Item>
          )}
        />
      </Card>

      {/* Right: permission config */}
      <Card style={{ flex: 1 }} size="small">
        {!selectedGroup ? (
          <Empty description="请先选择用户组" style={{ marginTop: 80 }} />
        ) : (
          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
              <Space>
                <Title level={5} style={{ margin: 0 }}>{selectedGroup.name}</Title>
                <Tag>{selectedGroup.member_count} 名成员</Tag>
              </Space>
              <Button type="primary" icon={<SaveOutlined />} onClick={handleSave} loading={saving}>
                保存权限
              </Button>
            </div>

            {/* Table-level permissions */}
            <div style={{ marginBottom: 24 }}>
              <Text strong>表级权限</Text>
              <div style={{ marginTop: 8 }}>
                <Space wrap>
                  {ERP_TABLES.map((t) => (
                    <Checkbox
                      key={t.name}
                      checked={tablePerms[t.name]}
                      onChange={(e) => {
                        const checked = e.target.checked;
                        setTablePerms((prev) => ({ ...prev, [t.name]: checked }));
                        const updates: Record<string, boolean> = {};
                        t.fields.forEach((f) => { updates[`${t.name}.${f}`] = checked; });
                        setFieldPerms((prev) => ({ ...prev, ...updates }));
                      }}
                    >
                      {t.label}
                    </Checkbox>
                  ))}
                </Space>
              </div>
            </div>

            {/* Field-level permissions */}
            <div>
              <Text strong>字段级权限</Text>
              <Table
                columns={permColumns}
                dataSource={fieldRows}
                size="small"
                pagination={false}
                style={{ marginTop: 8 }}
                scroll={{ y: 300 }}
              />
            </div>

            {/* Data-level filter conditions */}
            <div style={{ marginTop: 24 }}>
              <Text strong>数据级过滤条件</Text>
              <div style={{ marginTop: 8 }}>
                {ERP_TABLES.filter((t) => tablePerms[t.name]).map((t) => (
                  <div key={t.name} style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                    <Tag style={{ minWidth: 72, textAlign: 'center' }}>{t.label}</Tag>
                    <Input
                      placeholder={`SQL WHERE 条件，如: region = '华东'`}
                      value={filterConditions[t.name] || ''}
                      onChange={(e) => setFilterConditions((prev) => ({ ...prev, [t.name]: e.target.value }))}
                      style={{ flex: 1 }}
                      maxLength={500}
                    />
                  </div>
                ))}
                {ERP_TABLES.every((t) => !tablePerms[t.name]) && (
                  <Text type="secondary" style={{ fontSize: 12 }}>请先启用表级权限</Text>
                )}
              </div>
            </div>

            <Divider />

            {/* Member management UI */}
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                <Text strong>成员列表</Text>
                <Button
                  size="small"
                  type="primary"
                  ghost
                  icon={<UserAddOutlined />}
                  onClick={() => setAddModalOpen(true)}
                >
                  添加成员
                </Button>
              </div>
              <List
                size="small"
                bordered
                loading={membersLoading}
                dataSource={members}
                locale={{ emptyText: '暂无成员' }}
                renderItem={(member) => (
                  <List.Item
                    actions={[
                      <Button
                        key="remove"
                        type="link"
                        size="small"
                        danger
                        onClick={() => handleRemoveMember(member)}
                      >
                        移除
                      </Button>,
                    ]}
                  >
                    <List.Item.Meta
                      title={member.username}
                      description={`${member.email} · ${member.role === 'admin' ? '管理员' : member.role === 'manager' ? '部门经理' : '高管'}`}
                    />
                  </List.Item>
                )}
              />
            </div>
          </div>
        )}
      </Card>

      {/* Add member modal */}
      <Modal
        title="添加成员"
        open={addModalOpen}
        onCancel={() => { setAddModalOpen(false); setUserSearch(''); }}
        footer={null}
        destroyOnClose
      >
        <Input
          placeholder="搜索用户名"
          prefix={<SearchOutlined />}
          value={userSearch}
          onChange={(e) => setUserSearch(e.target.value)}
          allowClear
          style={{ marginBottom: 12 }}
        />
        <List
          size="small"
          loading={userListLoading}
          dataSource={availableUsers}
          locale={{ emptyText: '未找到可添加的用户' }}
          renderItem={(user) => (
            <List.Item
              actions={[
                <Button
                  key="add"
                  type="link"
                  size="small"
                  loading={addingMember}
                  onClick={() => handleAddMember(user.id)}
                >
                  添加
                </Button>,
              ]}
            >
              <List.Item.Meta
                title={user.username}
                description={`${user.email} · ${user.role === 'admin' ? '管理员' : user.role === 'manager' ? '部门经理' : '高管'}`}
              />
            </List.Item>
          )}
        />
      </Modal>

      {/* Create group modal */}
      <Modal
        title="新建用户组"
        open={createGroupOpen}
        onOk={handleCreateGroup}
        onCancel={() => { setCreateGroupOpen(false); groupForm.resetFields(); }}
        confirmLoading={creatingGroup}
        okText="创建"
        cancelText="取消"
        destroyOnClose
      >
        <Form form={groupForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="name"
            label="组名称"
            rules={[
              { required: true, message: '请输入组名称' },
              { min: 3, max: 50, message: '组名称长度 3~50 个字符' },
            ]}
          >
            <Input placeholder="例：销售组" />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
            rules={[{ max: 200, message: '描述不超过 200 个字符' }]}
          >
            <Input placeholder="可选描述" />
          </Form.Item>
        </Form>
      </Modal>

      {/* Edit group modal */}
      <Modal
        title="编辑用户组"
        open={editGroupOpen}
        onOk={handleEditGroupSubmit}
        onCancel={() => setEditGroupOpen(false)}
        confirmLoading={editingGroup}
        okText="保存"
        cancelText="取消"
        destroyOnClose
      >
        <Form form={editForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="name"
            label="组名称"
            rules={[
              { required: true, message: '请输入组名称' },
              { min: 3, max: 50, message: '组名称长度 3~50 个字符' },
            ]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
            rules={[{ max: 200, message: '描述不超过 200 个字符' }]}
          >
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
