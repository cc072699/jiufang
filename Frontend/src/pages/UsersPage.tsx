import { useState } from 'react';
import { Table, Typography, Tag, Space, Button, Input, Select, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { PlusOutlined, EditOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { getUserList, updateUser, deleteUser } from '../shared/api';
import { confirmAction } from '../shared/ui';
import type { UserRecord } from '../shared/api/types';

const { Title } = Typography;

const ROLE_LABELS: Record<string, string> = {
  admin: '管理员',
  manager: '部门经理',
  executive: '高管',
};

const ROLE_COLORS: Record<string, string> = {
  admin: 'red',
  manager: 'blue',
  executive: 'green',
};

export default function UsersPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [keyword, setKeyword] = useState('');
  const [role, setRole] = useState<string | undefined>();
  const [status, setStatus] = useState<number | undefined>();

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['users', page, pageSize, keyword, role, status],
    queryFn: () => getUserList({ page, page_size: pageSize, username: keyword || undefined, role, status }),
  });

  const users = data?.users ?? [];
  const total = data?.total ?? 0;

  const handleToggleStatus = (userId: string, username: string, currentStatus: number) => {
    const newStatus = currentStatus === 1 ? 0 : 1;
    const actionLabel = newStatus === 0 ? '停用' : '启用';
    confirmAction({
      title: `${actionLabel}用户`,
      content: `确定${actionLabel}用户「${username}」？`,
      onOk: async () => {
        try {
          await updateUser(userId, { status: newStatus });
          message.success(`已${actionLabel}`);
          queryClient.invalidateQueries({ queryKey: ['users'] });
        } catch {
          message.error(`${actionLabel}失败`);
        }
      },
    });
  };

  const handleDelete = (userId: string, username: string) => {
    confirmAction({
      title: '删除用户',
      content: `确定删除用户「${username}」？此操作不可恢复。`,
      onOk: async () => {
        try {
          await deleteUser(userId);
          message.success('已删除');
          queryClient.invalidateQueries({ queryKey: ['users'] });
        } catch {
          message.error('删除失败');
        }
      },
    });
  };

  const columns: ColumnsType<UserRecord> = [
    { title: '用户名', dataIndex: 'username', width: 150 },
    { title: '邮箱', dataIndex: 'email', ellipsis: true },
    {
      title: '角色',
      dataIndex: 'role',
      width: 120,
      render: (r: string) => <Tag color={ROLE_COLORS[r]}>{ROLE_LABELS[r] || r}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      render: (s: number) => <Tag color={s === 1 ? 'green' : 'default'}>{s === 1 ? '正常' : '停用'}</Tag>,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 180,
      render: (v: string) => new Date(v).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      width: 220,
      render: (_, record) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => navigate(`/users/${record.id}/edit`, { state: { record } })}>
            编辑
          </Button>
          {record.status === 1 ? (
            <Button size="small" danger onClick={() => handleToggleStatus(record.id, record.username, record.status)}>
              停用
            </Button>
          ) : (
            <Button size="small" type="primary" onClick={() => handleToggleStatus(record.id, record.username, record.status)}>
              启用
            </Button>
          )}
          <Button size="small" danger onClick={() => handleDelete(record.id, record.username)}>
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>人员管理</Title>
        <Space>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/users/new')}>
            新增人员
          </Button>
        </Space>
      </div>

      <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
        <Input
          placeholder="搜索用户名"
          prefix={<SearchOutlined />}
          value={keyword}
          onChange={(e) => { setKeyword(e.target.value); setPage(1); }}
          style={{ width: 200 }}
          allowClear
        />
        <Select
          placeholder="角色筛选"
          value={role}
          onChange={(v) => { setRole(v); setPage(1); }}
          allowClear
          style={{ width: 140 }}
          options={[
            { label: '全部角色', value: undefined },
            { label: '管理员', value: 'admin' },
            { label: '部门经理', value: 'manager' },
            { label: '高管', value: 'executive' },
          ]}
        />
        <Select
          placeholder="状态筛选"
          value={status}
          onChange={(v) => { setStatus(v); setPage(1); }}
          allowClear
          style={{ width: 120 }}
          options={[
            { label: '全部状态', value: undefined },
            { label: '正常', value: 1 },
            { label: '停用', value: 0 },
          ]}
        />
        <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
          刷新
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={users}
        rowKey="id"
        loading={isLoading}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          pageSizeOptions: ['20', '50', '100'],
          onChange: (p, ps) => { if (ps !== pageSize) { setPage(1); setPageSize(ps); } else { setPage(p); } },
        }}
      />
    </div>
  );
}
