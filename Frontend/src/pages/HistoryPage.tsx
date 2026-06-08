import { useState } from 'react';
import { Table, Typography, Tag, Space, Button, Select, Modal, message, DatePicker } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { ReloadOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { getHistoryList, getHistoryDetail, deleteHistory } from '../shared/api';
import { confirmAction } from '../shared/ui';
import { QUERY_STATUS_MAP, QUERY_STATUS_FILTER_OPTIONS } from '../shared/constants';
import type { HistoryRecord, HistoryDetailData } from '../shared/api/types';
import type { Dayjs } from 'dayjs';

const { Title } = Typography;
const { RangePicker } = DatePicker;

export default function HistoryPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [status, setStatus] = useState<string | undefined>();
  const [startTime, setStartTime] = useState<string | undefined>();
  const [endTime, setEndTime] = useState<string | undefined>();
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailData, setDetailData] = useState<HistoryDetailData | null>(null);

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['history', page, pageSize, status, startTime, endTime],
    queryFn: () =>
      getHistoryList({
        page,
        page_size: pageSize,
        status,
        start_time: startTime,
        end_time: endTime,
      }),
  });

  const records = data?.records ?? [];
  const total = data?.total ?? 0;

  const handleViewDetail = async (recordId: string) => {
    try {
      const detail = await getHistoryDetail(recordId);
      setDetailData(detail);
      setDetailVisible(true);
    } catch {
      message.error('获取详情失败');
    }
  };

  const handleContinue = async (record: HistoryRecord) => {
    try {
      const detail = await getHistoryDetail(record.id);
      navigate('/query', { state: { sessionId: detail.session_id, loadSession: true } });
    } catch {
      message.error('获取会话信息失败');
    }
  };

  // TODO: PRD 要求批量删除历史（最多50条），API 仅定义单条删除，待后端提供批量接口后实现

  const handleDelete = (recordId: string) => {
    confirmAction({
      title: '删除确认',
      content: '确定删除该条查询记录？此操作不可撤销。',
      onOk: async () => {
        try {
          await deleteHistory(recordId);
          message.success('已删除');
          queryClient.invalidateQueries({ queryKey: ['history'] });
        } catch {
          message.error('删除失败');
        }
      },
    });
  };

  const handleDateChange = (dates: [Dayjs | null, Dayjs | null] | null) => {
    if (dates) {
      setStartTime(dates[0]?.toISOString());
      setEndTime(dates[1]?.toISOString());
    } else {
      setStartTime(undefined);
      setEndTime(undefined);
    }
    setPage(1);
  };

  const columns: ColumnsType<HistoryRecord> = [
    {
      title: '查询内容',
      dataIndex: 'input',
      ellipsis: true,
      width: '40%',
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (s: string) => {
        const tag = QUERY_STATUS_MAP[s] ?? { color: 'default', label: s };
        return <Tag color={tag.color}>{tag.label}</Tag>;
      },
    },
    {
      title: '结果数',
      dataIndex: 'result_count',
      width: 100,
      render: (v: number | undefined) => v ?? '-',
    },
    {
      title: '耗时',
      dataIndex: 'execution_time',
      width: 100,
      render: (v: number | undefined) => (v != null ? `${v}ms` : '-'),
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      width: 180,
      render: (v: string) => new Date(v).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      width: 200,
      render: (_, record) => (
        <Space>
          <Button size="small" icon={<EyeOutlined />} onClick={() => handleViewDetail(record.id)}>
            详情
          </Button>
          <Button size="small" icon={<ReloadOutlined />} onClick={() => handleContinue(record)}>
            追问
          </Button>
          <Button size="small" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record.id)}>
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>历史记录</Title>
        <Space>
          <Select
            placeholder="状态筛选"
            value={status}
            onChange={(v) => { setStatus(v); setPage(1); }}
            allowClear
            style={{ width: 120 }}
            options={QUERY_STATUS_FILTER_OPTIONS}
          />
          <RangePicker
            showTime
            onChange={handleDateChange}
            placeholder={['开始时间', '结束时间']}
          />
          <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
            刷新
          </Button>
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={records}
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

      <Modal
        title="查询详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={640}
      >
        {detailData && (
          <div>
            <p><strong>查询内容：</strong>{detailData.input}</p>
            <p><strong>生成 SQL：</strong><code>{detailData.sql}</code></p>
            <p>
              <strong>状态：</strong>
              <Tag color={QUERY_STATUS_MAP[detailData.status]?.color ?? 'default'}>
                {QUERY_STATUS_MAP[detailData.status]?.label ?? detailData.status}
              </Tag>
            </p>
            {detailData.error_message && <p><strong>错误信息：</strong>{detailData.error_message}</p>}
            {detailData.result_count != null && <p><strong>结果数：</strong>{detailData.result_count}</p>}
            {detailData.execution_time != null && <p><strong>耗时：</strong>{detailData.execution_time}ms</p>}
            {detailData.result_data && (
              <div style={{ marginBottom: 12 }}>
                <strong>查询结果：</strong>
                <pre style={{ background: '#f5f5f5', padding: 12, borderRadius: 6, maxHeight: 200, overflow: 'auto', fontSize: 12, marginTop: 4 }}>
                  {detailData.result_data}
                </pre>
              </div>
            )}
            <p><strong>时间：</strong>{new Date(detailData.created_at).toLocaleString('zh-CN')}</p>
            <Button
              type="primary"
              icon={<ReloadOutlined />}
              onClick={() => {
                setDetailVisible(false);
                navigate('/query', { state: { sessionId: detailData.session_id, loadSession: true } });
              }}
            >
              续接追问
            </Button>
          </div>
        )}
      </Modal>
    </div>
  );
}
