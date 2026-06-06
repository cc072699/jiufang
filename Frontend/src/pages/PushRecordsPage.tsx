import { useState, useEffect } from 'react';
import { Table, Typography, Tag, Space, Button, Select, Input, DatePicker } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons';
import { useLocation } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getPushRecordList } from '../shared/api';
import type { PushRecord } from '../shared/api/types';
import { PUSH_TYPE_LABELS, PUSH_TYPE_COLORS, PUSH_TYPE_FILTER_OPTIONS, PUSH_STATUS_LABELS, PUSH_STATUS_COLORS, PUSH_STATUS_FILTER_OPTIONS } from '../shared/constants';

const { Title } = Typography;
const { RangePicker } = DatePicker;

export default function PushRecordsPage() {
  const location = useLocation();
  const navState = location.state as { reportName?: string } | null;

  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [pushType, setPushType] = useState<string | undefined>();
  const [pushStatus, setPushStatus] = useState<string | undefined>();
  const [startTime, setStartTime] = useState<string | undefined>();
  const [endTime, setEndTime] = useState<string | undefined>();
  const [sourceName, setSourceName] = useState(navState?.reportName || '');

  useEffect(() => {
    if (navState?.reportName) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setSourceName(navState.reportName);
      window.history.replaceState({}, '');
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['push-records', page, pageSize, pushType, pushStatus, startTime, endTime],
    queryFn: () => getPushRecordList({
      page, page_size: pageSize, push_type: pushType, status: pushStatus,
      start_time: startTime, end_time: endTime,
    }),
  });

  const records = data?.records ?? [];
  const total = data?.total ?? 0;

  // Client-side filter by source_name (until backend supports API param)
  const filteredRecords = sourceName
    ? records.filter((r) => r.source_name.includes(sourceName))
    : records;

  const columns: ColumnsType<PushRecord> = [
    {
      title: '类型',
      dataIndex: 'push_type',
      width: 100,
      render: (v: string) => <Tag color={PUSH_TYPE_COLORS[v]}>{PUSH_TYPE_LABELS[v]}</Tag>,
    },
    { title: '来源名称', dataIndex: 'source_name', ellipsis: true },
    { title: '接收人', dataIndex: 'recipient', width: 120 },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (s: string) => <Tag color={PUSH_STATUS_COLORS[s]}>{PUSH_STATUS_LABELS[s]}</Tag>,
    },
    {
      title: '错误信息',
      dataIndex: 'error_message',
      ellipsis: true,
      render: (v: string) => v || '-',
    },
    {
      title: '推送时间',
      dataIndex: 'pushed_at',
      width: 180,
      render: (v: string) => new Date(v).toLocaleString('zh-CN'),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>推送记录</Title>
        <Space>
          <Input
            placeholder="来源名称"
            prefix={<SearchOutlined />}
            value={sourceName}
            onChange={(e) => { setSourceName(e.target.value); setPage(1); }}
            allowClear
            style={{ width: 160 }}
          />
          <Select
            placeholder="推送类型"
            value={pushType}
            onChange={(v) => { setPushType(v); setPage(1); }}
            allowClear
            style={{ width: 120 }}
            options={PUSH_TYPE_FILTER_OPTIONS}
          />
          <Select
            placeholder="推送状态"
            value={pushStatus}
            onChange={(v) => { setPushStatus(v); setPage(1); }}
            allowClear
            style={{ width: 120 }}
            options={PUSH_STATUS_FILTER_OPTIONS}
          />
          <RangePicker
            showTime
            onChange={(dates) => {
              if (dates) {
                setStartTime(dates[0]?.toISOString());
                setEndTime(dates[1]?.toISOString());
              } else {
                setStartTime(undefined);
                setEndTime(undefined);
              }
              setPage(1);
            }}
            placeholder={['开始时间', '结束时间']}
          />
          <Button icon={<ReloadOutlined />} onClick={() => refetch()}>刷新</Button>
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={filteredRecords}
        rowKey="id"
        loading={isLoading}
        pagination={{
          current: page, pageSize, total: sourceName ? filteredRecords.length : total,
          showSizeChanger: true, pageSizeOptions: ['20', '50', '100'],
          onChange: (p, ps) => { if (ps !== pageSize) { setPage(1); setPageSize(ps); } else { setPage(p); } },
        }}
      />
    </div>
  );
}
