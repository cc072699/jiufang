import { useState } from 'react';
import { Table, Typography, Tag, Space, Button, message, Input, Select } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, HistoryOutlined, SearchOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { getReportList, deleteReport, updateReport } from '../shared/api';
import { confirmAction } from '../shared/ui';
import { SCHEDULE_LABELS, ACTIVE_STATUS_FILTER_OPTIONS, STATUS_LABELS, STATUS_COLORS } from '../shared/constants';
import type { ReportRecord } from '../shared/api/types';

const { Title } = Typography;

export default function ReportsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [name, setName] = useState('');
  const [status, setStatus] = useState<string | undefined>();

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['reports', page, pageSize, name, status],
    queryFn: () => getReportList({ page, page_size: pageSize, name: name || undefined, status }),
  });

  const reports = data?.reports ?? [];
  const total = data?.total ?? 0;

  const handleToggleStatus = (record: ReportRecord) => {
    const newStatus = record.status === 'active' ? 'inactive' : 'active';
    const label = newStatus === 'active' ? '启用' : '停用';
    confirmAction({
      title: `${label}报告`,
      content: `确定${label}报告「${record.name}」？`,
      okType: 'primary',
      onOk: async () => {
        try {
          await updateReport(record.id, { status: newStatus });
          message.success(`已${label}`);
          queryClient.invalidateQueries({ queryKey: ['reports'] });
        } catch {
          message.error('操作失败');
        }
      },
    });
  };

  const handleDelete = (record: ReportRecord) => {
    confirmAction({
      title: '删除报告',
      content: `确定删除报告「${record.name}」？此操作不可撤销。`,
      onOk: async () => {
        try {
          await deleteReport(record.id);
          message.success('已删除');
          queryClient.invalidateQueries({ queryKey: ['reports'] });
        } catch {
          message.error('删除失败');
        }
      },
    });
  };

  const columns: ColumnsType<ReportRecord> = [
    { title: '报告名称', dataIndex: 'name', ellipsis: true },
    {
      title: '推送频率',
      dataIndex: 'schedule_type',
      width: 100,
      render: (v: string) => SCHEDULE_LABELS[v] || v,
    },
    {
      title: '推送时间',
      dataIndex: 'schedule_time',
      width: 160,
      render: (v: string) => new Date(v).toLocaleString('zh-CN'),
    },
    { title: '接收人数', dataIndex: 'recipients_count', width: 100 },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      render: (s: string) => <Tag color={STATUS_COLORS[s]}>{STATUS_LABELS[s] || s}</Tag>,
    },
    {
      title: '上次推送',
      dataIndex: 'last_run_at',
      width: 160,
      render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-',
    },
    {
      title: '操作',
      width: 280,
      render: (_, record) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => navigate(`/reports/${record.id}/edit`, { state: { record } })}>
            编辑
          </Button>
          <Button size="small" icon={<HistoryOutlined />} onClick={() => navigate('/push-records', { state: { reportName: record.name } })}>
            发送记录
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
        <Title level={4} style={{ margin: 0 }}>定时报告</Title>
        <Space>
          <Input
            placeholder="搜索报告名称"
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
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/reports/new')}>
            新建报告
          </Button>
          <Button icon={<ReloadOutlined />} onClick={() => refetch()}>刷新</Button>
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={reports}
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
