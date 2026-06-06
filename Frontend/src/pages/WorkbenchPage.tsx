import { Typography, Card, Space, Row, Col, List, Tag, Spin } from 'antd';
import { SearchOutlined, HistoryOutlined, StarOutlined, ClockCircleOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getHistoryList, getFavoriteList } from '../shared/api';

const { Title, Text } = Typography;

export default function WorkbenchPage() {
  const navigate = useNavigate();

  const cards = [
    { icon: <SearchOutlined style={{ fontSize: 32 }} />, title: '对话查询', desc: '用自然语言查询 ERP 数据', path: '/query' },
    { icon: <HistoryOutlined style={{ fontSize: 32 }} />, title: '历史记录', desc: '查看历史查询', path: '/history' },
    { icon: <StarOutlined style={{ fontSize: 32 }} />, title: '我的收藏', desc: '常用查询收藏', path: '/favorites' },
  ];

  const { data: historyData, isLoading: historyLoading } = useQuery({
    queryKey: ['workbench', 'recent-history'],
    queryFn: () => getHistoryList({ page: 1, page_size: 3 }),
  });

  const { data: favoriteData, isLoading: favoriteLoading } = useQuery({
    queryKey: ['workbench', 'recent-favorites'],
    queryFn: () => getFavoriteList({ page: 1, page_size: 3 }),
  });

  const recentRecords = historyData?.records ?? [];
  const recentFavorites = favoriteData?.favorites ?? [];

  const statusTagMap: Record<string, { color: string; label: string }> = {
    success: { color: 'green', label: '成功' },
    failed: { color: 'red', label: '失败' },
  };

  return (
    <div>
      <Title level={4}>工作台</Title>
      <Text type="secondary">欢迎使用久方查询助手，请选择功能入口</Text>

      <Space size="large" style={{ marginTop: 24 }} wrap>
        {cards.map((c) => (
          <Card
            key={c.path}
            hoverable
            style={{ width: 220, textAlign: 'center' }}
            onClick={() => navigate(c.path)}
          >
            <div style={{ marginBottom: 12 }}>{c.icon}</div>
            <Title level={5}>{c.title}</Title>
            <Text type="secondary">{c.desc}</Text>
          </Card>
        ))}
      </Space>

      <Row gutter={24} style={{ marginTop: 32 }}>
        <Col xs={24} lg={12}>
          <Card
            title={<><ClockCircleOutlined /> 最近查询</>}
            extra={<a onClick={() => navigate('/history')}>查看全部</a>}
          >
            <Spin spinning={historyLoading}>
              {recentRecords.length === 0 && !historyLoading ? (
                <Text type="secondary">暂无查询记录</Text>
              ) : (
                <List
                  dataSource={recentRecords}
                  renderItem={(item) => {
                    const tag = statusTagMap[item.status] ?? { color: 'default', label: item.status };
                    return (
                      <List.Item>
                        <List.Item.Meta
                          title={
                            <span style={{ cursor: 'pointer' }} onClick={() => navigate('/history')}>
                              {item.input}
                            </span>
                          }
                          description={new Date(item.created_at).toLocaleString('zh-CN')}
                        />
                        <Tag color={tag.color}>{tag.label}</Tag>
                      </List.Item>
                    );
                  }}
                />
              )}
            </Spin>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card
            title={<><StarOutlined /> 我的收藏</>}
            extra={<a onClick={() => navigate('/favorites')}>查看全部</a>}
          >
            <Spin spinning={favoriteLoading}>
              {recentFavorites.length === 0 && !favoriteLoading ? (
                <Text type="secondary">暂无收藏</Text>
              ) : (
                <List
                  dataSource={recentFavorites}
                  renderItem={(item) => (
                    <List.Item>
                      <List.Item.Meta
                        title={
                          <span style={{ cursor: 'pointer' }} onClick={() => navigate('/query', { state: { input: item.input } })}>
                            {item.name}
                          </span>
                        }
                        description={new Date(item.created_at).toLocaleString('zh-CN')}
                      />
                    </List.Item>
                  )}
                />
              )}
            </Spin>
          </Card>
        </Col>
      </Row>
    </div>
  );
}
