import { useState } from 'react';
import { Table, Typography, Tag, Space, Button, Select, Input, DatePicker, Descriptions } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { getLogList } from '../shared/api';
import type { LogRecord } from '../shared/api/types';

const { Title, Text } = Typography;

const OP_TYPE_LABELS: Record<string, string> = {
  login: '登录',
  logout: '登出',
  query: '查询',
  create_user: '创建用户',
  update_user: '更新用户',
  delete_user: '删除用户',
  config_permission: '更新权限',
  create_report: '创建报告',
  update_report: '更新报告',
  delete_report: '删除报告',
  create_alert: '创建预警',
  update_alert: '更新预警',
  delete_alert: '删除预警',
};

// 筛选下拉: 设计文档 API-028 定义的 6 种筛选参数值
const OP_FILTER_OPTIONS = [
  { label: '登录', value: 'login' },
  { label: '登出', value: 'logout' },
  { label: '查询', value: 'query' },
  { label: '创建', value: 'create' },
  { label: '更新', value: 'update' },
  { label: '删除', value: 'delete' },
];

const OP_TYPE_COLORS: Record<string, string> = {
  login: 'green',
  logout: 'default',
  query: 'blue',
  create_user: 'orange',
  update_user: 'orange',
  delete_user: 'red',
  create_group: 'purple',
  update_permission: 'purple',
  create_report: 'cyan',
  update_report: 'cyan',
  delete_report: 'red',
  create_alert: 'magenta',
  update_alert: 'magenta',
  delete_alert: 'red',
};

export default function LogsPage() {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [opType, setOpType] = useState<string | undefined>();
  const [userId, setUserId] = useState<string>('');
  const [startTime, setStartTime] = useState<string | undefined>();
  const [endTime, setEndTime] = useState<string | undefined>();

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['logs', page, pageSize, opType, userId, startTime, endTime],
    queryFn: () => getLogList({
      page,
      page_size: pageSize,
      operation_type: opType,
      user_id: userId || undefined,
      start_time: startTime,
      end_time: endTime,
    }),
  });

  const logs = data?.logs ?? [];
  const total = data?.total ?? 0;

  const columns: ColumnsType<LogRecord> = [
    { title: '用户名', dataIndex: 'username', width: 120 },
    {
      title: '操作类型',
      dataIndex: 'operation_type',
      width: 140,
      render: (v: string) => (
        <Tag color={OP_TYPE_COLORS[v] || 'default'}>
          {OP_TYPE_LABELS[v] || v}
        </Tag>
      ),
    },
    {
      title: '操作详情',
      dataIndex: 'operation_detail',
      ellipsis: true,
      render: (v: string) => {
        if (!v) return '-';
        try {
          const obj = JSON.parse(v);
          return <Text code style={{ fontSize: 12 }}>{JSON.stringify(obj)}</Text>;
        } catch {
          return v;
        }
      },
    },
    { title: 'IP 地址', dataIndex: 'ip_address', width: 140 },
    {
      title: '时间',
      dataIndex: 'created_at',
      width: 180,
      render: (v: string) => new Date(v).toLocaleString('zh-CN'),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>操作日志</Title>
        <Space wrap>
          <Input
            placeholder="用户 ID"
            value={userId}
            onChange={(e) => { setUserId(e.target.value); setPage(1); }}
            allowClear
            style={{ width: 140 }}
            prefix={<SearchOutlined />}
          />
          <DatePicker.RangePicker
            showTime
            placeholder={['开始时间', '结束时间']}
            onChange={(dates) => {
              setStartTime(dates?.[0]?.toISOString());
              setEndTime(dates?.[1]?.toISOString());
              setPage(1);
            }}
          />
          <Select
            placeholder="操作类型"
            value={opType}
            onChange={(v) => { setOpType(v); setPage(1); }}
            allowClear
            style={{ width: 160 }}
            options={[
              { label: '全部类型', value: undefined },
              ...OP_FILTER_OPTIONS,
            ]}
          />
          <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
            刷新
          </Button>
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={logs}
        rowKey="id"
        loading={isLoading}
        expandable={{
          expandedRowRender: (record) => {
            const detail = record.operation_detail;
            if (!detail) return <Text type="secondary">无详情</Text>;
            try {
              const obj = JSON.parse(detail);
              return (
                <Descriptions size="small" column={2} bordered>
                  {Object.entries(obj).map(([key, val]) => (
                    <Descriptions.Item key={key} label={key}>
                      {typeof val === 'object' ? JSON.stringify(val) : String(val)}
                    </Descriptions.Item>
                  ))}
                </Descriptions>
              );
            } catch {
              return <Text>{detail}</Text>;
            }
          },
        }}
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
