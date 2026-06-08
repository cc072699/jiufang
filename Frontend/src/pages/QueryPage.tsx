import { useState, useRef, useEffect } from 'react';
import { Input, Button, Card, Typography, Space, Alert, Skeleton, List, message, Modal, Form, Dropdown } from 'antd';
import { SendOutlined, ThunderboltOutlined, PlusOutlined, MessageOutlined, ExportOutlined, SearchOutlined, StarOutlined, DownOutlined } from '@ant-design/icons';
import { useLocation } from 'react-router-dom';
import { useQueryClient, useQuery } from '@tanstack/react-query';
import { queryNaturalLanguage, submitFeedback, createFavorite, getHistoryList, getHistoryBySessionID, exportQueryResult } from '../shared/api';
import { getToken } from '../shared/auth/token';
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
  const [feedbackModal, setFeedbackModal] = useState<{ visible: boolean; reason: string; queryRecordId: string; query: string }>({
    visible: false, reason: '', queryRecordId: '', query: '',
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

  // Merge API history with in-memory sessions, grouped by session_id
  const sessionMap = new Map<string, Session>();
  for (const r of historyData?.records ?? []) {
    const existing = sessionMap.get(r.session_id);
    if (!existing || new Date(r.created_at).getTime() < existing.createdAt) {
      sessionMap.set(r.session_id, {
        id: r.session_id,
        title: r.input.slice(0, 20),
        createdAt: new Date(r.created_at).getTime(),
      });
    }
  }
  const apiSessions = Array.from(sessionMap.values());
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
      const records = await getHistoryBySessionID(s.id);
      const msgs: Message[] = [];
      for (const r of records) {
        msgs.push({ role: 'user', content: r.input });
        let parsedResult: QueryResultData | undefined;
        if (r.result_data) {
          try {
            parsedResult = JSON.parse(r.result_data) as QueryResultData;
            if (!parsedResult.session_id) parsedResult.session_id = r.session_id;
            // Inject query_record_id from history record so export/feedback buttons work
            parsedResult.query_record_id = r.id;
          } catch { /* ignore parse failure */ }
        }
        msgs.push({
          role: 'assistant',
          content: parsedResult?.understanding ?? r.input,
          result: parsedResult,
          error: r.status === 'failed' ? (r.error_message || '查询失败') : undefined,
        });
      }
      setMessages(msgs);
      setSessionId(s.id);
    } catch {
      setMessages([]);
      setSessionId(s.id);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleFeedback = async (queryRecordId: string) => {
    try {
      await submitFeedback({ query_record_id: Number(queryRecordId), rating: 'satisfied' });
      message.success('感谢反馈！');
    } catch {
      message.error('反馈提交失败，请稍后重试');
    }
  };

  const handleFeedbackDissatisfied = (queryRecordId: string, query: string) => {
    setFeedbackModal({ visible: true, reason: '', queryRecordId, query });
  };

  const handleFeedbackSubmit = async () => {
    if (!feedbackModal.reason.trim()) {
      message.warning('请填写不满意原因后再提交');
      return;
    }
    try {
      await submitFeedback({
        query_record_id: Number(feedbackModal.queryRecordId),
        rating: 'unsatisfied',
        reason: feedbackModal.reason.trim(),
      });
      message.success('感谢反馈！');
      setFeedbackModal({ visible: false, reason: '', queryRecordId: '', query: '' });
    } catch {
      message.error('反馈提交失败，请稍后重试');
    }
  };

  const downloadBlob = (blob: Blob, fileName: string) => {
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = fileName;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const handleExportCSV = (result: QueryResultData) => {
    if (!result.rows || !result.columns) {
      message.warning('无数据可导出');
      return;
    }
    const headers = result.columns.map((c) => c.name);
    const csvRows = [headers.join(',')];
    for (const row of result.rows) {
      csvRows.push(headers.map((h) => {
        const val = row[h];
        if (val == null) return '';
        const str = String(val);
        return str.includes(',') || str.includes('"') || str.includes('\n')
          ? `"${str.replace(/"/g, '""')}"` : str;
      }).join(','));
    }
    const bom = '﻿';
    const blob = new Blob([bom + csvRows.join('\n')], { type: 'text/csv;charset=utf-8' });
    downloadBlob(blob, `${result.understanding || '查询结果'}.csv`);
    message.success('CSV 导出成功');
  };

  const handleExportExcel = async (queryRecordId: string, title?: string) => {
    try {
      message.loading({ content: '正在导出...', key: 'export' });
      const result = await exportQueryResult({
        query_record_id: queryRecordId,
        format: 'excel',
        title: title || '查询结果',
      });
      // Download via direct fetch (bypass apiClient baseURL prefix)
      const token = getToken();
      const resp = await fetch(result.file_url, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      });
      if (!resp.ok) throw new Error('下载失败');
      const blob = await resp.blob();
      downloadBlob(blob, result.file_name);
      message.success({ content: 'Excel 导出成功', key: 'export' });
    } catch (err: unknown) {
      message.error({ content: err instanceof Error ? err.message : '导出失败', key: 'export' });
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
    const state = location.state as { input?: string; sessionId?: string; loadSession?: boolean } | null;
    if (!state) return;

    if (state.loadSession && state.sessionId) {
      // Load full conversation from history (追问 flow)
      autoSubmitDone.current = true;
      handleSelectSession({ id: state.sessionId, title: '', createdAt: 0 });
      window.history.replaceState({}, '');
    } else if (state.input) {
      // Auto-submit a query (favorites flow)
      autoSubmitDone.current = true;
      handleSend(state.input, state.sessionId);
      window.history.replaceState({}, '');
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const suggestedQuestions = [
    '查询所有供应商列表',
    '采购单总数和总金额是多少',
    '查询采购单及供应商名称',
    '查询当前库存情况',
    '销售订单总金额是多少',
    '查询所有客户信息',
    '查询生产任务列表',
  ];

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
              <div style={{ marginTop: 12, maxWidth: 640, marginInline: 'auto' }}>
                <Space wrap>
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
                              {msg.result.can_export && msg.result.rows && msg.result.rows.length > 0 && (
                                <Dropdown
                                  menu={{
                                    items: [
                                      ...(msg.result.query_record_id ? [{
                                        key: 'excel',
                                        label: '导出 Excel',
                                        onClick: () => handleExportExcel(msg.result!.query_record_id, msg.result!.understanding),
                                      }] : []),
                                      {
                                        key: 'csv',
                                        label: '导出 CSV',
                                        onClick: () => handleExportCSV(msg.result!),
                                      },
                                    ],
                                  }}
                                >
                                  <Button size="small" icon={<ExportOutlined />}>
                                    导出 <DownOutlined />
                                  </Button>
                                </Dropdown>
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
                            {msg.result.query_record_id && (
                              <>
                                <Button
                                  type="text"
                                  size="small"
                                  onClick={() => handleFeedback(msg.result!.query_record_id)}
                                >
                                  {'👍'} 有帮助
                                </Button>
                                <Button
                                  type="text"
                                  size="small"
                                  onClick={() => {
                                    const prevMsg = messages[i - 1];
                                    handleFeedbackDissatisfied(msg.result!.query_record_id, prevMsg?.content ?? '');
                                  }}
                                >
                                  {'👎'} 没帮助
                                </Button>
                              </>
                            )}
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
