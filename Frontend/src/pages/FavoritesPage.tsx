import { useState } from 'react';
import { Table, Typography, Space, Button, message, Tooltip, Input } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { DeleteOutlined, SearchOutlined, ReloadOutlined, StarFilled } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { getFavoriteList, deleteFavorite } from '../shared/api';
import { confirmAction } from '../shared/ui';
import type { FavoriteRecord } from '../shared/api/types';

const { Title, Text } = Typography;

export default function FavoritesPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [keyword, setKeyword] = useState('');

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['favorites', page, pageSize, keyword],
    queryFn: () => getFavoriteList({ page, page_size: pageSize, name: keyword || undefined }),
  });

  const favorites = data?.favorites ?? [];
  const total = data?.total ?? 0;

  const handleQuery = (record: FavoriteRecord) => {
    navigate('/query', { state: { input: record.input } });
  };

  const handleDelete = (favoriteId: string) => {
    confirmAction({
      title: '取消收藏',
      content: '确定取消该收藏？此操作不可撤销。',
      okText: '取消收藏',
      onOk: async () => {
        try {
          await deleteFavorite(favoriteId);
          message.success('已取消收藏');
          queryClient.invalidateQueries({ queryKey: ['favorites'] });
          queryClient.invalidateQueries({ queryKey: ['workbench'] });
        } catch {
          message.error('操作失败');
        }
      },
    });
  };

  const columns: ColumnsType<FavoriteRecord> = [
    {
      title: '',
      width: 32,
      render: () => <StarFilled style={{ color: '#faad14' }} />,
    },
    {
      title: '收藏名称',
      dataIndex: 'name',
      width: 180,
    },
    {
      title: '查询内容',
      dataIndex: 'input',
      ellipsis: true,
    },
    {
      title: '备注',
      dataIndex: 'description',
      width: 160,
      ellipsis: true,
      render: (v: string) => v || '-',
    },
    {
      title: '收藏时间',
      dataIndex: 'created_at',
      width: 180,
      render: (v: string) => new Date(v).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      width: 160,
      render: (_, record) => (
        <Space>
          <Tooltip title="用此查询条件发起查询">
            <Button size="small" type="primary" icon={<SearchOutlined />} onClick={() => handleQuery(record)}>
              查询
            </Button>
          </Tooltip>
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
        <Title level={4} style={{ margin: 0 }}>我的收藏</Title>
        <Space>
          <Input
            placeholder="搜索收藏"
            value={keyword}
            onChange={(e) => { setKeyword(e.target.value); setPage(1); }}
            allowClear
            style={{ width: 200 }}
            prefix={<SearchOutlined />}
          />
          <Text type="secondary">共 {total} 条</Text>
          <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
            刷新
          </Button>
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={favorites}
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
