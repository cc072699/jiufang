import { useState, useRef, useEffect } from 'react';
import { Input, Button, Card, Typography, Space, Alert, Skeleton, List, Tooltip, message, Modal, Form } from 'antd';
import { SendOutlined, ThunderboltOutlined, PlusOutlined, MessageOutlined, ExportOutlined, SearchOutlined, StarOutlined } from '@ant-design/icons';
import { useLocation } from 'react-router-dom';
import { useQueryClient, useQuery } from '@tanstack/react-query';
import { queryNaturalLanguage, submitFeedback, createFavorite, getHistoryList, getHistoryDetail } from '../shared/api';
import { DataTable, ChartRenderer, EmptyState } from '../shared/ui';
import type { QueryResultData } from '../shared/api/types';

const { Text } = Typography;

interface Message {
  role: 'user' | 'assistant';
  content: string;
  result?: QueryResultData;
  error?: string;
  inputDraft?: string;
}

interface Session {
  id: string;
  title: string;
  createdAt: number;
}

export default function QueryPage() {
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [sessionId, setSessionId] = useState<string | undefined>();
  const [sessions, setSessions] = useState<Session[]>([]);
  const [sessionSearch, setSessionSearch] = useState('');
  const endRef = useRef<HTMLDivElement>(null);
  const location = useLocation();
  const queryClient = useQueryClient();
  const [feedbackModal, setFeedbackModal] = useState<{ visible: boolean; reason: string; sessionId: string; query: string }>({
    visible: false, reason: '', sessionId: '', query: '',
  });
  const [favoriteModal, setFavoriteModal] = useState<{ visible: boolean; input: string; sql: string }>({
    visible: false, input: '', sql: '',
  });
  const [favoriteForm] = Form.useForm();

  // Fetch recent history for sidebar
  const { data: historyData } = useQuery({
    queryKey: ['query-sessions'],
    queryFn: () => getHistoryList({ page: 1, page_size: 50 }),
  });

  // Merge API history with in-memory sessions, dedup by id
  const apiSessions: Session[] = (historyData?.records ?? []).map((r) => ({
    id: r.id,
    title: r.input.slice(0, 20),
    createdAt: new Date(r.created_at).getTime(),
  }));
  const mergedSessions = [...sessions];
  for (const s of apiSessions) {
    if (!mergedSessions.some((m) => m.id === s.id)) {
      mergedSessions.push(s);
    }
  }
  mergedSessions.sort((a, b) => b.createdAt - a.createdAt);

  const filteredSessions = sessionSearch
    ? mergedSessions.filter((s) => s.title.includes(sessionSearch))
    : mergedSessions;

  const scrollToBottom = () => {
    endRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages, loading]);

  const handleSend = async (text?: string, overrideSessionId?: string) => {
    const query = (text ?? input).trim();
    if (!query || loading) return;
    if (query.length > 500) return;

    const sid = overrideSessionId ?? sessionId;
    setInput('');
    setMessages((prev) => [...prev, { role: 'user', content: query }]);
    setLoading(true);

    try {
      const result = await queryNaturalLanguage({
        input: query,
        session_id: sid,
      });

      setSessionId(result.session_id);
      setMessages((prev) => [
        ...prev,
        { role: 'assistant', content: result.understanding, result },
      ]);

      // Add to sessions list if new session
      if (!sessionId) {
        setSessions((prev) => [
          { id: result.session_id, title: query.slice(0, 20), createdAt: Date.now() },
          ...prev,
        ]);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : '查询失败';
      setMessages((prev) => [
        ...prev,
        { role: 'assistant', content: '', error: msg, inputDraft: query },
      ]);
    } finally {
      setLoading(false);
    }
  };

  const handleRetry = (draft: string) => {
    setInput(draft);
  };

  const handleNewSession = () => {
    setMessages([]);
    setSessionId(undefined);
    setInput('');
  };

  const handleSelectSession = async (s: Session) => {
    if (s.id === sessionId) return;
    try {
      const detail = await getHistoryDetail(s.id);
      let parsedResult: QueryResultData | undefined;
      if (detail.result_data) {
        try {
          parsedResult = JSON.parse(detail.result_data) as QueryResultData;
          if (!parsedResult.session_id) parsedResult.session_id = detail.session_id;
        } catch { /* result_data 解析失败，仅展示文本 */ }
      }
      const assistantMsg: Message = {
        role: 'assistant',
        content: parsedResult?.understanding ?? `SQL: ${detail.sql}`,
        result: parsedResult,
        error: detail.status === 'failed' ? (detail.error_message || '查询失败') : undefined,
      };
      setMessages([{ role: 'user', content: detail.input }, assistantMsg]);
      setSessionId(detail.session_id);
    } catch {
      setMessages([{ role: 'user', content: s.title }]);
      setSessionId(s.id);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleFeedback = async (sessionId: string, query: string) => {
    try {
      await submitFeedback({ session_id: sessionId, query, rating: 'satisfied' });
      message.success('感谢反馈！');
    } catch {
      message.error('反馈提交失败，请稍后重试');
    }
  };

  const handleFeedbackDissatisfied = (sessionId: string, query: string) => {
    setFeedbackModal({ visible: true, reason: '', sessionId, query });
  };

  const handleFeedbackSubmit = async () => {
    if (!feedbackModal.reason.trim()) {
      message.warning('请填写不满意原因后再提交');
      return;
    }
    try {
      await submitFeedback({
        session_id: feedbackModal.sessionId,
        query: feedbackModal.query,
        rating: 'dissatisfied',
        reason: feedbackModal.reason.trim(),
      });
      message.success('感谢反馈！');
      setFeedbackModal({ visible: false, reason: '', sessionId: '', query: '' });
    } catch {
      message.error('反馈提交失败，请稍后重试');
    }
  };

  const handleOpenFavorite = (queryInput: string, querySql?: string) => {
    favoriteForm.resetFields();
    setFavoriteModal({ visible: true, input: queryInput, sql: querySql || '' });
  };

  const handleFavoriteSubmit = async () => {
    try {
      const values = await favoriteForm.validateFields();
      await createFavorite({ name: values.name, input: favoriteModal.input, sql: favoriteModal.sql || '(无SQL)', description: values.description });
      message.success('收藏成功');
      setFavoriteModal({ visible: false, input: '', sql: '' });
      queryClient.invalidateQueries({ queryKey: ['favorites'] });
      queryClient.invalidateQueries({ queryKey: ['workbench'] });
    } catch {
      // validation error — do nothing
    }
  };

  // Read navigation state from HistoryPage/FavoritesPage
  const autoSubmitDone = useRef(false);
  useEffect(() => {
    if (autoSubmitDone.current) return;
    const state = location.state as { input?: string; sessionId?: string } | null;
    if (state?.input) {
      autoSubmitDone.current = true;
      // eslint-disable-next-line react-hooks/set-state-in-effect
      handleSend(state.input, state.sessionId);
      window.history.replaceState({}, '');
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const suggestedQuestions = ['上个月销售额最高的产品是什么？', '各个区域的库存总量'];

  return (
    <div style={{ display: 'flex', height: 'calc(100vh - 160px)', gap: 16 }}>
      {/* Left sidebar - session list */}
      <div
        style={{
          width: 220,
          flexShrink: 0,
          display: 'flex',
          flexDirection: 'column',
          borderRight: '1px solid #f0f0f0',
          paddingRight: 16,
        }}
      >
        <Button
          type="dashed"
          icon={<PlusOutlined />}
          block
          onClick={handleNewSession}
          style={{ marginBottom: 12 }}
        >
          新建对话
        </Button>
        <Input
          placeholder="搜索会话"
          prefix={<SearchOutlined />}
          value={sessionSearch}
          onChange={(e) => setSessionSearch(e.target.value)}
          allowClear
          size="small"
          style={{ marginBottom: 8 }}
        />
        <div style={{ flex: 1, overflow: 'auto' }}>
          {filteredSessions.length === 0 ? (
            <Text type="secondary" style={{ fontSize: 12, display: 'block', textAlign: 'center', marginTop: 24 }}>
              暂无历史会话
            </Text>
          ) : (
            <List
              size="small"
              dataSource={filteredSessions}
              renderItem={(s) => (
                <List.Item
                  style={{
                    cursor: 'pointer',
                    padding: '8px 4px',
                    background: s.id === sessionId ? '#e6f4ff' : undefined,
                    borderRadius: 6,
                  }}
                  onClick={() => handleSelectSession(s)}
                >
                  <MessageOutlined style={{ marginRight: 8, color: '#1677ff' }} />
                  <Text ellipsis style={{ fontSize: 13 }}>{s.title}</Text>
                </List.Item>
              )}
            />
          )}
        </div>
      </div>

      {/* Main area */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
        <Typography.Title level={4} style={{ marginBottom: 16 }}>对话查询</Typography.Title>

        {/* Messages area */}
        <div style={{ flex: 1, overflow: 'auto', marginBottom: 16 }}>
          {messages.length === 0 && !loading && (
            <div style={{ textAlign: 'center', paddingTop: 80 }}>
              <ThunderboltOutlined style={{ fontSize: 48, color: '#1677ff', marginBottom: 16 }} />
              <div>
                <Text type="secondary">输入自然语言问题，即可查询 ERP 数据</Text>
              </div>
              <div style={{ marginTop: 12 }}>
                <Space>
                  {suggestedQuestions.map((q) => (
                    <Card
                      key={q}
                      size="small"
                      hoverable
                      style={{ cursor: 'pointer', borderColor: '#1677ff', maxWidth: 280 }}
                      styles={{ body: { padding: '8px 16px', fontSize: 13 } }}
                      onClick={() => handleSend(q)}
                    >
                      {q}
                    </Card>
                  ))}
                </Space>
              </div>
            </div>
          )}

          {messages.map((msg, i) => (
            <div
              key={i}
              style={{
                display: 'flex',
                justifyContent: msg.role === 'user' ? 'flex-end' : 'flex-start',
                marginBottom: 16,
              }}
            >
              {msg.role === 'user' ? (
                <Card
                  size="small"
                  style={{
                    maxWidth: '70%',
                    background: '#1677ff',
                    color: '#fff',
                    borderRadius: 12,
                  }}
                  styles={{ body: { color: '#fff', padding: '8px 16px' } }}
                >
                  {msg.content}
                </Card>
              ) : (
                <div style={{ maxWidth: '85%' }}>
                  {msg.error ? (
                    <Alert
                      type="error"
                      message="查询失败"
                      description={
                        <div>
                          <div>{msg.error}</div>
                          {msg.inputDraft && (
                            <Button
                              size="small"
                              type="link"
                              style={{ padding: 0, marginTop: 4 }}
                              onClick={() => handleRetry(msg.inputDraft!)}
                            >
                              重新输入
                            </Button>
                          )}
                        </div>
                      }
                      showIcon
                      style={{ borderRadius: 12 }}
                    />
                  ) : (
                    <Card size="small" style={{ borderRadius: 12 }}>
                      <Text type="secondary" style={{ fontSize: 13 }}>
                        {msg.content}
                      </Text>
                      {msg.result && (
                        <div style={{ marginTop: 12 }}>
                          {/* Table result */}
                          {msg.result.result_type === 'table' &&
                            msg.result.columns &&
                            msg.result.rows && (
                              <DataTable
                                columns={msg.result.columns}
                                rows={msg.result.rows}
                                maxRows={100}
                              />
                            )}
                          {/* Chart result with chart_config */}
                          {msg.result.result_type === 'chart' &&
                            msg.result.chart_config && (
                              <ChartRenderer
                                chartConfig={msg.result.chart_config}
                                columns={msg.result.columns}
                                rows={msg.result.rows}
                              />
                            )}
                          {/* Chart fallback: chart_config missing, show table */}
                          {msg.result.result_type === 'chart' &&
                            !msg.result.chart_config &&
                            msg.result.columns &&
                            msg.result.rows && (
                              <div>
                                <Alert
                                  type="warning"
                                  message="图表加载失败，已切换至表格模式"
                                  showIcon
                                  style={{ marginBottom: 8 }}
                                />
                                <DataTable
                                  columns={msg.result.columns}
                                  rows={msg.result.rows}
                                  maxRows={100}
                                />
                              </div>
                            )}
                          {msg.result.result_type === 'empty' && (
                            <EmptyState description="未找到匹配数据" />
                          )}
                          {/* Action bar: export + favorite */}
                          <div style={{ marginTop: 8, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                            <Space>
                              {msg.result.can_export && (
                                <Tooltip title="导出功能待后端接口">
                                  <Button size="small" icon={<ExportOutlined />} disabled>
                                    导出
                                  </Button>
                                </Tooltip>
                              )}
                              <Button
                                size="small"
                                icon={<StarOutlined />}
                                onClick={() => {
                                  const prevMsg = messages[i - 1];
                                  if (prevMsg?.role === 'user') {
                                    handleOpenFavorite(prevMsg.content, msg.result?.sql);
                                  }
                                }}
                              >
                                收藏
                              </Button>
                            </Space>
                          </div>
                          {/* Feedback buttons */}
                          <div style={{ marginTop: 4, display: 'flex', gap: 4 }}>
                            <Button
                              type="text"
                              size="small"
                              onClick={() => {
                                const prevMsg = messages[i - 1];
                                if (msg.result && prevMsg?.role === 'user') {
                                  handleFeedback(msg.result.session_id, prevMsg.content);
                                }
                              }}
                            >
                              {'👍'} 有帮助
                            </Button>
                            <Button
                              type="text"
                              size="small"
                              onClick={() => {
                                const prevMsg = messages[i - 1];
                                if (msg.result && prevMsg?.role === 'user') {
                                  handleFeedbackDissatisfied(msg.result.session_id, prevMsg.content);
                                }
                              }}
                            >
                              {'👎'} 没帮助
                            </Button>
                          </div>
                          {msg.result.suggested_questions &&
                            msg.result.suggested_questions.length > 0 && (
                              <div style={{ marginTop: 12 }}>
                                <Text type="secondary" style={{ fontSize: 12 }}>
                                  你可能还想问：
                                </Text>
                                <div style={{ marginTop: 6 }}>
                                  <Space wrap>
                                    {msg.result.suggested_questions.map((q) => (
                                      <Card
                                        key={q}
                                        size="small"
                                        hoverable
                                        style={{
                                          cursor: 'pointer',
                                          borderColor: '#1677ff',
                                          maxWidth: 260,
                                        }}
                                        styles={{ body: { padding: '4px 12px', fontSize: 12 } }}
                                        onClick={() => handleSend(q)}
                                      >
                                        {q}
                                      </Card>
                                    ))}
                                  </Space>
                                </div>
                              </div>
                            )}
                        </div>
                      )}
                    </Card>
                  )}
                </div>
              )}
            </div>
          ))}

          {loading && (
            <div style={{ display: 'flex', justifyContent: 'flex-start', marginBottom: 16 }}>
              <Card size="small" style={{ borderRadius: 12, width: 360 }}>
                <Skeleton active paragraph={{ rows: 2 }} />
              </Card>
            </div>
          )}

          <div ref={endRef} />
        </div>

        {/* Input area */}
        <div style={{ display: 'flex', gap: 8 }}>
          <Input.TextArea
            value={input}
            onChange={(e) => {
              if (e.target.value.length <= 500) setInput(e.target.value);
            }}
            onKeyDown={handleKeyDown}
            placeholder="输入自然语言问题，按 Enter 发送 (1-500字)"
            autoSize={{ minRows: 1, maxRows: 4 }}
            disabled={loading}
            style={{ flex: 1 }}
            showCount
            maxLength={500}
          />
          <Button
            type="primary"
            icon={<SendOutlined />}
            onClick={() => handleSend()}
            loading={loading}
            disabled={!input.trim() || input.length > 500}
          >
            发送
          </Button>
        </div>
      </div>

      {/* Feedback dissatisfaction reason modal */}
      <Modal
        title="反馈"
        open={feedbackModal.visible}
        onOk={handleFeedbackSubmit}
        onCancel={() => setFeedbackModal((prev) => ({ ...prev, visible: false }))}
        okText="提交"
        cancelText="取消"
      >
        <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
          请描述不满意原因：
        </Typography.Text>
        <Input.TextArea
          value={feedbackModal.reason}
          onChange={(e) => setFeedbackModal((prev) => ({ ...prev, reason: e.target.value }))}
          placeholder="请输入原因 (1-500字)"
          maxLength={500}
          rows={4}
          showCount
        />
      </Modal>

      {/* Favorite modal */}
      <Modal
        title="收藏查询"
        open={favoriteModal.visible}
        onOk={handleFavoriteSubmit}
        onCancel={() => setFavoriteModal({ visible: false, input: '', sql: '' })}
        okText="收藏"
        cancelText="取消"
      >
        <Form form={favoriteForm} layout="vertical">
          <Form.Item name="name" label="收藏名称" rules={[{ required: true, message: '请输入名称' }, { min: 1, max: 100, message: '1-100个字符' }]}>
            <Input placeholder="为该查询取个名称" maxLength={100} showCount />
          </Form.Item>
          <Form.Item name="description" label="备注">
            <Input.TextArea placeholder="可选备注" maxLength={200} rows={2} showCount />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
