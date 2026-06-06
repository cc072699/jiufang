import { useState } from 'react';
import { Table, Typography, Tag, Space, Button, message, Input, Select } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, SearchOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { getAlertList, deleteAlert, updateAlert } from '../shared/api';
import { confirmAction } from '../shared/ui';
import type { AlertRecord } from '../shared/api/types';
import { ACTIVE_STATUS_FILTER_OPTIONS, STATUS_COLORS, STATUS_LABELS } from '../shared/constants';

const { Title } = Typography;

export default function AlertsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [name, setName] = useState('');
  const [status, setStatus] = useState<string | undefined>();

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['alerts', page, pageSize, name, status],
    queryFn: () => getAlertList({ page, page_size: pageSize, name: name || undefined, status }),
  });

  const alerts = data?.alerts ?? [];
  const total = data?.total ?? 0;

  const handleToggleStatus = (record: AlertRecord) => {
    const newStatus = record.status === 'active' ? 'inactive' : 'active';
    const label = newStatus === 'active' ? '启用' : '停用';
    confirmAction({
      title: `${label}预警`,
      content: `确定${label}预警规则「${record.name}」？`,
      okType: 'primary',
      onOk: async () => {
        try {
          await updateAlert(record.id, { status: newStatus });
          message.success(`已${label}`);
          queryClient.invalidateQueries({ queryKey: ['alerts'] });
        } catch {
          message.error('操作失败');
        }
      },
    });
  };

  const handleDelete = (record: AlertRecord) => {
    confirmAction({
      title: '删除预警规则',
      content: `确定删除「${record.name}」？此操作不可撤销。`,
      onOk: async () => {
        try {
          await deleteAlert(record.id);
          message.success('已删除');
          queryClient.invalidateQueries({ queryKey: ['alerts'] });
        } catch {
          message.error('删除失败');
        }
      },
    });
  };

  const columns: ColumnsType<AlertRecord> = [
    { title: '规则名称', dataIndex: 'name', ellipsis: true },
    { title: '触发条件', dataIndex: 'condition', ellipsis: true },
    { title: '接收人数', dataIndex: 'recipients_count', width: 100 },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      render: (s: string) => <Tag color={STATUS_COLORS[s]}>{STATUS_LABELS[s]}</Tag>,
    },
    {
      title: '上次触发',
      dataIndex: 'last_triggered_at',
      width: 160,
      render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 160,
      render: (v: string) => new Date(v).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      width: 220,
      render: (_, record) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => navigate(`/alerts/${record.id}/edit`, { state: { record } })}>
            编辑
          </Button>
          <Button size="small" onClick={() => handleToggleStatus(record)}>
            {record.status === 'active' ? '停用' : '启用'}
          </Button>
          <Button size="small" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record)}>
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>预警规则</Title>
        <Space>
          <Input
            placeholder="搜索规则名称"
            prefix={<SearchOutlined />}
            value={name}
            onChange={(e) => { setName(e.target.value); setPage(1); }}
            style={{ width: 180 }}
            allowClear
          />
          <Select
            placeholder="状态筛选"
            value={status}
            onChange={(v) => { setStatus(v); setPage(1); }}
            allowClear
            style={{ width: 120 }}
            options={ACTIVE_STATUS_FILTER_OPTIONS}
          />
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/alerts/new')}>
            新建规则
          </Button>
          <Button icon={<ReloadOutlined />} onClick={() => refetch()}>刷新</Button>
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={alerts}
        rowKey="id"
        loading={isLoading}
        pagination={{
          current: page, pageSize, total,
          showSizeChanger: true, pageSizeOptions: ['20', '50', '100'],
          onChange: (p, ps) => { if (ps !== pageSize) { setPage(1); setPageSize(ps); } else { setPage(p); } },
        }}
      />
    </div>
  );
}
