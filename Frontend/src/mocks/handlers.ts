import { http, HttpResponse, delay } from 'msw';
import type {
  ApiResponse, LoginData, QueryResultData,
  HistoryRecord, HistoryListData, HistoryDetailData,
  FavoriteRecord, FavoriteListData,
  UserRecord, UserListData, UserDetailData, GroupListData, GroupDetailData,
  LogListData, ReportRecord, ReportListData, ReportDetailData,
  AlertRecord, AlertListData, AlertDetailData, PushRecordListData,
  ChangePasswordRequest,
  GroupMemberItem, GroupMemberListData,
  ConfigurePermissionData,
} from '../shared/api/types';

const BASE = '/api/v1';

// ===== Mutable mock data stores =====

// Mock user database (for login)
const MOCK_USERS: Record<string, { password: string; disabled?: boolean; data: LoginData }> = {
  admin: {
    password: 'admin123',
    data: {
      token: 'mock-jwt-admin-token-001',
      expires_at: new Date(Date.now() + 24 * 3600_000).toISOString(),
      user: { id: 'u001', username: 'admin', role: 'admin', groups: ['管理员组'], email: 'admin@jiufang.com' },
    },
  },
  user: {
    password: 'user123',
    data: {
      token: 'mock-jwt-user-token-002',
      expires_at: new Date(Date.now() + 24 * 3600_000).toISOString(),
      user: { id: 'u002', username: 'user', role: 'executive', groups: ['销售组'], email: 'zhangsan@jiufang.com' },
    },
  },
  disabled_user: {
    password: 'pass123',
    disabled: true,
    data: {
      token: '',
      expires_at: '',
      user: { id: 'u003', username: 'disabled_user', role: 'manager', groups: [], email: 'disabled@jiufang.com' },
    },
  },
};

// Mutable user list for CRUD (API-004~007)
let MOCK_USERS_LIST: UserRecord[] = [
  { id: 'u001', username: 'admin', email: 'admin@jiufang.com', role: 'admin', status: 1, created_at: '2026-01-01T00:00:00Z' },
  { id: 'u002', username: 'zhangsan', email: 'zhangsan@jiufang.com', role: 'manager', status: 1, created_at: '2026-03-15T08:00:00Z' },
  { id: 'u003', username: 'lisi', email: 'lisi@jiufang.com', role: 'executive', status: 0, created_at: '2026-04-20T10:00:00Z' },
];

// Mutable history records (API-013~015)
let MOCK_HISTORY_RECORDS: HistoryRecord[] = [
  {
    id: 'h001',
    input: '上个月销售额最高的产品是什么？',
    sql: "SELECT product_name, SUM(amount) AS total FROM sales WHERE date >= '2026-05-01' GROUP BY product_name ORDER BY total DESC LIMIT 1",
    status: 'success',
    result_count: 1,
    execution_time: 320,
    created_at: '2026-06-01T10:30:00Z',
  },
  {
    id: 'h002',
    input: '各个区域的库存总量',
    sql: 'SELECT region, SUM(stock) AS total_stock FROM inventory GROUP BY region',
    status: 'success',
    result_count: 5,
    execution_time: 180,
    created_at: '2026-05-31T14:20:00Z',
  },
  {
    id: 'h003',
    input: '本周新增客户数量',
    sql: "SELECT COUNT(*) AS cnt FROM customers WHERE created_at >= '2026-05-26'",
    status: 'failed',
    error_message: '表 customers 不存在',
    created_at: '2026-05-30T09:15:00Z',
  },
];

// Mutable favorites (API-016~018)
let MOCK_FAVORITES: FavoriteRecord[] = [
  {
    id: 'fav-001',
    name: '月度销售汇总',
    input: '上个月各产品销售额汇总',
    sql: "SELECT product_name, SUM(amount) FROM sales WHERE date >= '2026-05-01' GROUP BY product_name",
    description: '每月初查看',
    created_at: '2026-05-15T09:00:00Z',
  },
  {
    id: 'fav-002',
    name: '库存预警查询',
    input: '库存低于安全库存的产品',
    sql: 'SELECT * FROM inventory WHERE stock < safety_stock',
    created_at: '2026-05-20T14:30:00Z',
  },
];

// Mutable groups list (API-008~011)
let MOCK_GROUPS_LIST: GroupListData['groups'] = [
  { id: 'g001', name: '管理员组', description: '系统管理员', member_count: 1, created_at: '2026-01-01T00:00:00Z' },
  { id: 'g002', name: '销售组', description: '销售部门人员', member_count: 2, created_at: '2026-02-01T00:00:00Z' },
];

// Mutable group members data
const MOCK_GROUP_MEMBERS: Record<string, GroupMemberItem[]> = {
  g001: [
    { id: 'u001', username: 'admin', email: 'admin@jiufang.com', role: 'admin' },
  ],
  g002: [
    { id: 'u002', username: 'zhangsan', email: 'zhangsan@jiufang.com', role: 'manager' },
    { id: 'u003', username: 'lisi', email: 'lisi@jiufang.com', role: 'executive' },
  ],
};

// Mutable reports list (API-019~022)
let MOCK_REPORTS_LIST: ReportRecord[] = [
  { id: 'r001', name: '月度销售汇总', schedule_type: 'monthly', schedule_time: '2026-06-01T08:00:00Z', recipients: ['u001', 'u002', 'u003'], recipients_count: 3, status: 'active', sql: "SELECT product_name, SUM(amount) AS total FROM sales WHERE date >= DATE_SUB(CURDATE(), INTERVAL 1 MONTH) GROUP BY product_name ORDER BY total DESC", last_run_at: '2026-05-01T08:00:00Z', created_at: '2026-04-15T10:00:00Z' },
  { id: 'r002', name: '周库存报告', schedule_type: 'weekly', schedule_time: '2026-06-02T09:00:00Z', recipients: ['u001', 'u002'], recipients_count: 2, status: 'inactive', sql: "SELECT warehouse, product_name, stock, safety_stock FROM inventory WHERE stock < safety_stock", created_at: '2026-05-01T14:00:00Z' },
];

// Mutable alerts list (API-023~026)
let MOCK_ALERTS_LIST: AlertRecord[] = [
  { id: 'a001', name: '库存低于安全线', condition: 'stock < safety_stock', recipients: ['u001', 'u002'], recipients_count: 2, status: 'active', last_triggered_at: '2026-06-01T06:00:00Z', created_at: '2026-04-20T10:00:00Z' },
  { id: 'a002', name: '销售额异常下降', condition: 'daily_amount < avg_amount * 0.5', recipients: ['u001', 'u002', 'u003'], recipients_count: 3, status: 'active', created_at: '2026-05-10T14:00:00Z' },
];

// Mutable permissions per group (API-012)
const MOCK_PERMISSIONS: Record<string, ConfigurePermissionData> = {};

// ===== End of mutable stores =====

// Mutable logs list (API-028)
const MOCK_LOGS_LIST: LogListData['logs'] = [
  { id: 'l001', user_id: 'u001', username: 'admin', operation_type: 'login', operation_detail: '{"ip":"192.168.1.100"}', ip_address: '192.168.1.100', created_at: '2026-06-02T08:00:00Z' },
  { id: 'l002', user_id: 'u002', username: 'zhangsan', operation_type: 'query', operation_detail: '{"input":"上个月销售额"}', ip_address: '192.168.1.101', created_at: '2026-06-02T09:15:00Z' },
  { id: 'l003', user_id: 'u001', username: 'admin', operation_type: 'create_user', operation_detail: '{"target":"lisi"}', ip_address: '192.168.1.100', created_at: '2026-06-02T10:30:00Z' },
  { id: 'l004', user_id: 'u001', username: 'admin', operation_type: 'logout', operation_detail: '{}', ip_address: '192.168.1.100', created_at: '2026-06-02T12:00:00Z' },
  { id: 'l005', user_id: 'u002', username: 'zhangsan', operation_type: 'query', operation_detail: '{"input":"库存总量"}', ip_address: '192.168.1.101', created_at: '2026-06-02T14:20:00Z' },
];

// Mutable push records (API-027)
const MOCK_PUSH_RECORDS: PushRecordListData['records'] = [
  { id: 'p001', push_type: 'report', source_id: 'r001', source_name: '月度销售汇总', recipient: '张三', status: 'success', pushed_at: '2026-06-01T08:00:00Z' },
  { id: 'p002', push_type: 'alert', source_id: 'a001', source_name: '库存低于安全线', recipient: '李四', status: 'success', pushed_at: '2026-06-01T06:00:00Z' },
  { id: 'p003', push_type: 'report', source_id: 'r001', source_name: '月度销售汇总', recipient: '王五', status: 'failed', error_message: '企业微信推送失败', pushed_at: '2026-06-01T08:01:00Z' },
];

// Mock query responses keyed by input keywords
function buildQueryResponse(input: string): QueryResultData {
  // BR-007 反问机制：含指代词时返回空结果+联想问题
  if (input.includes('那个') || input.includes('它的') || input.includes('对比一下')) {
    return {
      session_id: 'sess-clarify',
      understanding: `您的问题"${input}"存在歧义，请确认您想查询的具体内容。`,
      result_type: 'empty',
      suggested_questions: [
        '查询本月各产品销售额',
        '查询本月各区域库存量',
        '查询本月新增客户数量',
      ],
      can_export: false,
    };
  }
  if (input.includes('销售') || input.includes('金额')) {
    return {
      session_id: 'sess-001',
      understanding: `已理解您的问题："${input}"，正在查询销售数据。`,
      result_type: 'table',
      columns: [
        { name: '产品名称', type: 'string' },
        { name: '销售金额', type: 'number' },
        { name: '销售数量', type: 'number' },
      ],
      rows: [
        { '产品名称': '产品A', '销售金额': 128500, '销售数量': 320 },
        { '产品名称': '产品B', '销售金额': 96200, '销售数量': 215 },
        { '产品名称': '产品C', '销售金额': 73800, '销售数量': 189 },
      ],
      suggested_questions: ['哪个区域销售最好？', '与上月对比如何？', '销售趋势图表'],
      can_export: true,
    };
  }

  if (input.includes('库存')) {
    return {
      session_id: 'sess-002',
      understanding: `已理解您的问题："${input}"，正在查询库存数据。`,
      result_type: 'chart',
      columns: [
        { name: '仓库', type: 'string' },
        { name: '库存量', type: 'number' },
      ],
      rows: [
        { '仓库': '华东仓', '库存量': 5200 },
        { '仓库': '华南仓', '库存量': 3800 },
        { '仓库': '华北仓', '库存量': 4100 },
        { '仓库': '西南仓', '库存量': 2600 },
      ],
      chart_config: {
        chart_type: 'bar',
        title: '各仓库库存量',
        x_field: '仓库',
        y_field: '库存量',
      },
      suggested_questions: ['库存周转率是多少？', '哪些产品库存不足？'],
      can_export: true,
    };
  }

  // Default: empty result
  return {
    session_id: 'sess-003',
    understanding: `已理解您的问题："${input}"，未找到匹配数据。`,
    result_type: 'empty',
    suggested_questions: ['查询本月销售数据', '查看库存汇总'],
    can_export: false,
  };
}

export const handlers = [
  // POST /api/v1/auth/login (BR-069: 区分三种错误)
  http.post(`${BASE}/auth/login`, async ({ request }) => {
    await delay(500);
    const body = (await request.json()) as { username: string; password: string };

    const user = MOCK_USERS[body.username];
    // detailed-design.md API-002: 用户不存在或密码不匹配 → 40101, 用户已停用 → 40301
    if (!user || user.password !== body.password) {
      return HttpResponse.json<ApiResponse>({
        code: 40101,
        message: '用户名或密码错误',
      }, { status: 401 });
    }
    if (user.disabled) {
      return HttpResponse.json<ApiResponse>({
        code: 40301,
        message: '账号已停用，请联系管理员',
      }, { status: 403 });
    }

    return HttpResponse.json<ApiResponse<LoginData>>({
      code: 200,
      message: 'success',
      data: user.data,
    });
  }),

  // POST /api/v1/auth/logout
  http.post(`${BASE}/auth/logout`, async () => {
    await delay(200);
    return HttpResponse.json<ApiResponse>({
      code: 200,
      message: 'success',
    });
  }),

  // GET /api/v1/profile (设计文档: 用户完整资料)
  http.get(`${BASE}/profile`, async () => {
    await delay(200);
    return HttpResponse.json<ApiResponse>({
      code: 200,
      message: 'success',
      data: {
        id: 'u001',
        username: 'admin',
        email: 'admin@jiufang.com',
        role: 'admin',
        avatar: '',
        groups: ['管理员组'],
        status: 1,
        created_at: '2026-01-01T00:00:00Z',
      },
    });
  }),

  // POST /api/v1/query
  http.post(`${BASE}/query`, async ({ request }) => {
    await delay(800);
    const body = (await request.json()) as { input: string; session_id?: string };

    if (!body.input?.trim()) {
      return HttpResponse.json<ApiResponse>({
        code: 40001,
        message: '参数校验失败：input 不能为空',
      }, { status: 400 });
    }

    const result = buildQueryResponse(body.input);

    if (body.session_id) {
      result.session_id = body.session_id;
    }

    return HttpResponse.json<ApiResponse<QueryResultData>>({
      code: 200,
      message: 'success',
      data: result,
    });
  }),

  // GET /api/v1/history (API-013)
  http.get(`${BASE}/history`, async ({ request }) => {
    await delay(300);
    const url = new URL(request.url);
    const page = Number(url.searchParams.get('page') || 1);
    const pageSize = Number(url.searchParams.get('page_size') || 20);
    const status = url.searchParams.get('status') || '';
    const startTime = url.searchParams.get('start_time') || '';
    const endTime = url.searchParams.get('end_time') || '';

    let filtered = [...MOCK_HISTORY_RECORDS];
    if (status) {
      filtered = filtered.filter((r) => r.status === status);
    }
    if (startTime) {
      filtered = filtered.filter((r) => r.created_at >= startTime);
    }
    if (endTime) {
      filtered = filtered.filter((r) => r.created_at <= endTime);
    }

    const start = (page - 1) * pageSize;
    const paged = filtered.slice(start, start + pageSize);

    return HttpResponse.json<ApiResponse<HistoryListData>>({
      code: 200,
      message: 'success',
      data: {
        records: paged,
        total: filtered.length,
        page,
        page_size: pageSize,
      },
    });
  }),

  // GET /api/v1/history/:record_id (API-014)
  http.get(`${BASE}/history/:recordId`, async ({ params }) => {
    await delay(200);
    const { recordId } = params;
    const record = MOCK_HISTORY_RECORDS.find((r) => r.id === recordId);
    if (!record) {
      return HttpResponse.json<ApiResponse>({
        code: 40401,
        message: '记录不存在',
      }, { status: 404 });
    }
    const detail: HistoryDetailData = {
      ...record,
      session_id: `sess-${recordId}`,
      result_data: JSON.stringify({ mock: true }),
    };
    return HttpResponse.json<ApiResponse<HistoryDetailData>>({
      code: 200,
      message: 'success',
      data: detail,
    });
  }),

  // DELETE /api/v1/history/:record_id (API-015)
  http.delete(`${BASE}/history/:recordId`, async ({ params }) => {
    await delay(200);
    MOCK_HISTORY_RECORDS = MOCK_HISTORY_RECORDS.filter((r) => r.id !== params.recordId);
    return HttpResponse.json<ApiResponse>({ code: 200, message: 'success' });
  }),

  // POST /api/v1/favorites (API-016)
  http.post(`${BASE}/favorites`, async ({ request }) => {
    await delay(300);
    const body = (await request.json()) as { name: string; input: string; sql: string; description?: string };
    const newFav: FavoriteRecord = {
      id: `fav-${Date.now()}`,
      name: body.name,
      input: body.input,
      sql: body.sql,
      description: body.description || '',
      created_at: new Date().toISOString(),
    };
    MOCK_FAVORITES.unshift(newFav);
    return HttpResponse.json<ApiResponse<FavoriteRecord>>({
      code: 200, message: 'success', data: newFav,
    });
  }),

  // GET /api/v1/favorites (API-017)
  http.get(`${BASE}/favorites`, async ({ request }) => {
    await delay(300);
    const url = new URL(request.url);
    const page = Number(url.searchParams.get('page') || 1);
    const pageSize = Number(url.searchParams.get('page_size') || 20);
    const name = url.searchParams.get('name') || '';

    let filtered = [...MOCK_FAVORITES];
    if (name) {
      filtered = filtered.filter((f) => f.name.includes(name));
    }

    const start = (page - 1) * pageSize;
    const paged = filtered.slice(start, start + pageSize);

    return HttpResponse.json<ApiResponse<FavoriteListData>>({
      code: 200, message: 'success',
      data: { favorites: paged, total: filtered.length, page, page_size: pageSize },
    });
  }),

  // DELETE /api/v1/favorites/:favorite_id (API-018)
  http.delete(`${BASE}/favorites/:favoriteId`, async ({ params }) => {
    await delay(200);
    MOCK_FAVORITES = MOCK_FAVORITES.filter((f) => f.id !== params.favoriteId);
    return HttpResponse.json<ApiResponse>({ code: 200, message: 'success' });
  }),

  // ============ 管理员后台 API ============

  // GET /api/v1/users (API-005)
  http.get(`${BASE}/users`, async ({ request }) => {
    await delay(300);
    const url = new URL(request.url);
    const page = Number(url.searchParams.get('page') || 1);
    const pageSize = Number(url.searchParams.get('page_size') || 20);
    const username = url.searchParams.get('username') || '';
    const role = url.searchParams.get('role') || '';
    const statusParam = url.searchParams.get('status');

    let filtered = [...MOCK_USERS_LIST];
    if (username) filtered = filtered.filter((u) => u.username.includes(username));
    if (role) filtered = filtered.filter((u) => u.role === role);
    if (statusParam !== null) filtered = filtered.filter((u) => u.status === Number(statusParam));

    const start = (page - 1) * pageSize;
    const paged = filtered.slice(start, start + pageSize);

    return HttpResponse.json<ApiResponse<UserListData>>({
      code: 200, message: 'success',
      data: { users: paged, total: filtered.length, page, page_size: pageSize },
    });
  }),

  // POST /api/v1/users (API-004)
  http.post(`${BASE}/users`, async ({ request }) => {
    await delay(400);
    const body = (await request.json()) as { username: string; email: string; role: string };
    const newUser: UserRecord = {
      id: `u-${Date.now()}`,
      username: body.username,
      email: body.email,
      role: body.role as 'admin' | 'manager' | 'executive',
      status: 1,
      created_at: new Date().toISOString(),
    };
    MOCK_USERS_LIST.unshift(newUser);
    const detail: UserDetailData = { ...newUser, groups: [] };
    return HttpResponse.json<ApiResponse<UserDetailData>>({ code: 200, message: 'success', data: detail });
  }),

  // PUT /api/v1/users/:user_id (API-006)
  http.put(`${BASE}/users/:userId`, async ({ request, params }) => {
    await delay(300);
    const body = (await request.json()) as Record<string, unknown>;
    const idx = MOCK_USERS_LIST.findIndex((u) => u.id === params.userId);
    if (idx < 0) {
      return HttpResponse.json<ApiResponse>({ code: 40401, message: '记录不存在' }, { status: 404 });
    }
    MOCK_USERS_LIST[idx] = {
      ...MOCK_USERS_LIST[idx],
      ...(body.username ? { username: body.username as string } : {}),
      ...(body.email ? { email: body.email as string } : {}),
      ...(body.role ? { role: body.role as 'admin' | 'manager' | 'executive' } : {}),
      ...(body.status !== undefined ? { status: body.status as number } : {}),
    };
    return HttpResponse.json<ApiResponse<UserDetailData>>({
      code: 200, message: 'success',
      data: { ...MOCK_USERS_LIST[idx], groups: [], updated_at: new Date().toISOString() },
    });
  }),

  // DELETE /api/v1/users/:user_id (API-007)
  http.delete(`${BASE}/users/:userId`, async ({ params }) => {
    await delay(200);
    MOCK_USERS_LIST = MOCK_USERS_LIST.filter((u) => u.id !== params.userId);
    return HttpResponse.json<ApiResponse>({ code: 200, message: 'success' });
  }),

  // GET /api/v1/groups (API-009)
  http.get(`${BASE}/groups`, async ({ request }) => {
    await delay(300);
    const url = new URL(request.url);
    const page = Number(url.searchParams.get('page') || 1);
    const pageSize = Number(url.searchParams.get('page_size') || 20);
    const start = (page - 1) * pageSize;
    const paged = MOCK_GROUPS_LIST.slice(start, start + pageSize);

    return HttpResponse.json<ApiResponse<GroupListData>>({
      code: 200, message: 'success',
      data: { groups: paged, total: MOCK_GROUPS_LIST.length, page, page_size: pageSize },
    });
  }),

  // POST /api/v1/groups (API-008)
  http.post(`${BASE}/groups`, async ({ request }) => {
    await delay(300);
    const body = (await request.json()) as { name: string; description?: string };
    const newGroup = {
      id: `g-${Date.now()}`,
      name: body.name,
      description: body.description || '',
      member_count: 0,
      created_at: new Date().toISOString(),
    };
    MOCK_GROUPS_LIST.unshift(newGroup);
    const detail: GroupDetailData = { ...newGroup, members: [] };
    return HttpResponse.json<ApiResponse<GroupDetailData>>({ code: 200, message: 'success', data: detail });
  }),

  // PUT /api/v1/groups/:group_id (API-010)
  http.put(`${BASE}/groups/:groupId`, async ({ request, params }) => {
    await delay(300);
    const body = (await request.json()) as Record<string, unknown>;
    const idx = MOCK_GROUPS_LIST.findIndex((g) => g.id === params.groupId);
    if (idx < 0) {
      return HttpResponse.json<ApiResponse>({ code: 40401, message: '记录不存在' }, { status: 404 });
    }
    MOCK_GROUPS_LIST[idx] = {
      ...MOCK_GROUPS_LIST[idx],
      ...(body.name !== undefined ? { name: body.name as string } : {}),
      ...(body.description !== undefined ? { description: body.description as string } : {}),
    };
    return HttpResponse.json<ApiResponse<GroupDetailData>>({
      code: 200, message: 'success',
      data: { ...MOCK_GROUPS_LIST[idx], members: MOCK_GROUP_MEMBERS[params.groupId as string] || [] },
    });
  }),

  // DELETE /api/v1/groups/:group_id (API-011)
  http.delete(`${BASE}/groups/:groupId`, async ({ params }) => {
    await delay(200);
    MOCK_GROUPS_LIST = MOCK_GROUPS_LIST.filter((g) => g.id !== params.groupId);
    return HttpResponse.json<ApiResponse>({ code: 200, message: 'success' });
  }),

  // GET /api/v1/groups/:group_id/permissions
  http.get(`${BASE}/groups/:groupId/permissions`, async ({ params }) => {
    await delay(200);
    const groupId = params.groupId as string;
    const stored = MOCK_PERMISSIONS[groupId];
    return HttpResponse.json<ApiResponse<ConfigurePermissionData>>({
      code: 200, message: 'success',
      data: stored ?? { group_id: groupId, permissions: [], created_at: new Date().toISOString() },
    });
  }),

  // POST /api/v1/groups/:group_id/permissions (API-012)
  http.post(`${BASE}/groups/:groupId/permissions`, async ({ params, request }) => {
    await delay(400);
    const body = (await request.json()) as { permissions: { table_name: string; allowed_fields: string; filter_condition?: string }[] };
    const groupId = params.groupId as string;
    const result: ConfigurePermissionData = {
      group_id: groupId,
      permissions: body.permissions.map((p, i) => ({ id: `perm-${Date.now()}-${i}`, ...p })),
      created_at: new Date().toISOString(),
    };
    MOCK_PERMISSIONS[groupId] = result;
    return HttpResponse.json<ApiResponse<ConfigurePermissionData>>({
      code: 200, message: 'success', data: result,
    });
  }),

  // GET /api/v1/logs (API-028)
  http.get(`${BASE}/logs`, async ({ request }) => {
    await delay(300);
    const url = new URL(request.url);
    const page = Number(url.searchParams.get('page') || 1);
    const pageSize = Number(url.searchParams.get('page_size') || 20);
    const operationType = url.searchParams.get('operation_type') || '';
    const userId = url.searchParams.get('user_id') || '';
    const startTime = url.searchParams.get('start_time') || '';
    const endTime = url.searchParams.get('end_time') || '';

    let filtered = [...MOCK_LOGS_LIST];
    if (operationType) filtered = filtered.filter((l) => l.operation_type.startsWith(operationType));
    if (userId) filtered = filtered.filter((l) => l.user_id === userId);
    if (startTime) filtered = filtered.filter((l) => l.created_at >= startTime);
    if (endTime) filtered = filtered.filter((l) => l.created_at <= endTime);

    const start = (page - 1) * pageSize;
    const paged = filtered.slice(start, start + pageSize);

    return HttpResponse.json<ApiResponse<LogListData>>({
      code: 200, message: 'success',
      data: { logs: paged, total: filtered.length, page, page_size: pageSize },
    });
  }),

  // ============ 报告/预警/推送记录 API ============

  // GET /api/v1/reports (API-020)
  http.get(`${BASE}/reports`, async ({ request }) => {
    await delay(300);
    const url = new URL(request.url);
    const page = Number(url.searchParams.get('page') || 1);
    const pageSize = Number(url.searchParams.get('page_size') || 20);
    const name = url.searchParams.get('name') || '';
    const status = url.searchParams.get('status') || '';

    let filtered = [...MOCK_REPORTS_LIST];
    if (name) filtered = filtered.filter((r) => r.name.includes(name));
    if (status) filtered = filtered.filter((r) => r.status === status);

    const start = (page - 1) * pageSize;
    const paged = filtered.slice(start, start + pageSize);

    return HttpResponse.json<ApiResponse<ReportListData>>({
      code: 200, message: 'success',
      data: { reports: paged, total: filtered.length, page, page_size: pageSize },
    });
  }),

  // POST /api/v1/reports (API-019)
  http.post(`${BASE}/reports`, async ({ request }) => {
    await delay(400);
    const body = (await request.json()) as Record<string, unknown>;
    const newReport: ReportRecord = {
      id: `r-${Date.now()}`,
      name: body.name as string,
      schedule_type: body.schedule_type as 'daily' | 'weekly' | 'monthly',
      schedule_time: body.schedule_time as string,
      recipients_count: (body.recipients as string[])?.length || 0,
      status: 'active',
      created_at: new Date().toISOString(),
    };
    MOCK_REPORTS_LIST.unshift(newReport);
    const detail: ReportDetailData = {
      ...newReport,
      sql: body.sql as string,
      recipients: (body.recipients as string[]) || [],
      description: body.description as string | undefined,
    };
    return HttpResponse.json<ApiResponse<ReportDetailData>>({ code: 200, message: 'success', data: detail });
  }),

  // GET /api/v1/reports/:report_id (详情)
  http.get(`${BASE}/reports/:reportId`, async ({ params }) => {
    await delay(200);
    const report = MOCK_REPORTS_LIST.find((r) => r.id === params.reportId);
    if (!report) {
      return HttpResponse.json<ApiResponse>({ code: 40401, message: '记录不存在' }, { status: 404 });
    }
    const MOCK_REPORT_DETAILS: Record<string, { recipients: string[]; description: string }> = {
      r001: { recipients: ['u001', 'u002', 'u003'], description: '每月初自动发送上月销售数据汇总' },
      r002: { recipients: ['u001', 'u002'], description: '每周一发送库存预警数据' },
    };
    const extra = MOCK_REPORT_DETAILS[params.reportId as string] || { recipients: [], description: '' };
    const detail: ReportDetailData = {
      ...report,
      sql: report.sql || '',
      recipients: extra.recipients,
      description: extra.description,
    };
    return HttpResponse.json<ApiResponse<ReportDetailData>>({ code: 200, message: 'success', data: detail });
  }),

  // PUT /api/v1/reports/:report_id (API-021)
  http.put(`${BASE}/reports/:reportId`, async ({ request, params }) => {
    await delay(300);
    const body = (await request.json()) as Record<string, unknown>;
    const idx = MOCK_REPORTS_LIST.findIndex((r) => r.id === params.reportId);
    if (idx >= 0) {
      MOCK_REPORTS_LIST[idx] = {
        ...MOCK_REPORTS_LIST[idx],
        ...(body.name !== undefined ? { name: body.name as string } : {}),
        ...(body.schedule_type !== undefined ? { schedule_type: body.schedule_type as 'daily' | 'weekly' | 'monthly' } : {}),
        ...(body.schedule_time !== undefined ? { schedule_time: body.schedule_time as string } : {}),
        ...(body.status !== undefined ? { status: body.status as 'active' | 'inactive' } : {}),
      };
      const merged = MOCK_REPORTS_LIST[idx];
      return HttpResponse.json<ApiResponse<ReportDetailData>>({
        code: 200, message: 'success',
        data: {
          id: merged.id,
          name: merged.name,
          sql: (body.sql as string) || merged.sql || '',
          schedule_type: merged.schedule_type,
          schedule_time: merged.schedule_time,
          recipients: (body.recipients as string[]) || [],
          description: (body.description as string) || '',
          status: merged.status,
          updated_at: new Date().toISOString(),
        },
      });
    }
    return HttpResponse.json<ApiResponse>({ code: 40401, message: '记录不存在' }, { status: 404 });
  }),

  // DELETE /api/v1/reports/:report_id (API-022)
  http.delete(`${BASE}/reports/:reportId`, async ({ params }) => {
    await delay(200);
    MOCK_REPORTS_LIST = MOCK_REPORTS_LIST.filter((r) => r.id !== params.reportId);
    return HttpResponse.json<ApiResponse>({ code: 200, message: 'success' });
  }),

  // GET /api/v1/alerts (API-024)
  http.get(`${BASE}/alerts`, async ({ request }) => {
    await delay(300);
    const url = new URL(request.url);
    const page = Number(url.searchParams.get('page') || 1);
    const pageSize = Number(url.searchParams.get('page_size') || 20);
    const name = url.searchParams.get('name') || '';
    const status = url.searchParams.get('status') || '';

    let filtered = [...MOCK_ALERTS_LIST];
    if (name) filtered = filtered.filter((a) => a.name.includes(name));
    if (status) filtered = filtered.filter((a) => a.status === status);

    const start = (page - 1) * pageSize;
    const paged = filtered.slice(start, start + pageSize);

    return HttpResponse.json<ApiResponse<AlertListData>>({
      code: 200, message: 'success',
      data: { alerts: paged, total: filtered.length, page, page_size: pageSize },
    });
  }),

  // POST /api/v1/alerts (API-023)
  http.post(`${BASE}/alerts`, async ({ request }) => {
    await delay(400);
    const body = (await request.json()) as Record<string, unknown>;
    const newAlert: AlertRecord = {
      id: `a-${Date.now()}`,
      name: body.name as string,
      condition: body.condition as string,
      recipients_count: (body.recipients as string[])?.length || 0,
      status: 'active',
      created_at: new Date().toISOString(),
    };
    MOCK_ALERTS_LIST.unshift(newAlert);
    const detail: AlertDetailData = {
      ...newAlert,
      sql: body.sql as string,
      recipients: (body.recipients as string[]) || [],
      description: body.description as string | undefined,
    };
    return HttpResponse.json<ApiResponse<AlertDetailData>>({ code: 200, message: 'success', data: detail });
  }),

  // GET /api/v1/alerts/:alert_id (详情)
  http.get(`${BASE}/alerts/:alertId`, async ({ params }) => {
    await delay(200);
    const alert = MOCK_ALERTS_LIST.find((a) => a.id === params.alertId);
    if (!alert) {
      return HttpResponse.json<ApiResponse>({ code: 40401, message: '记录不存在' }, { status: 404 });
    }
    const MOCK_ALERT_DETAILS: Record<string, { sql: string; recipients: string[]; description: string }> = {
      a001: { sql: "SELECT product_name, stock, safety_stock FROM inventory WHERE stock < safety_stock", recipients: ['u001', 'u002'], description: '当库存低于安全线时自动推送预警' },
      a002: { sql: "SELECT DATE(date) AS d, SUM(amount) AS daily_amount FROM sales GROUP BY DATE(date) HAVING daily_amount < (SELECT AVG(total) FROM (SELECT SUM(amount) AS total FROM sales GROUP BY DATE(date))) * 0.5", recipients: ['u001', 'u002', 'u003'], description: '监测日销售额异常下降' },
    };
    const extra = MOCK_ALERT_DETAILS[params.alertId as string] || { sql: '', recipients: [], description: '' };
    const detail: AlertDetailData = {
      ...alert,
      sql: extra.sql,
      recipients: extra.recipients,
      description: extra.description,
    };
    return HttpResponse.json<ApiResponse<AlertDetailData>>({ code: 200, message: 'success', data: detail });
  }),

  // PUT /api/v1/alerts/:alert_id (API-025)
  http.put(`${BASE}/alerts/:alertId`, async ({ request, params }) => {
    await delay(300);
    const body = (await request.json()) as Record<string, unknown>;
    const idx = MOCK_ALERTS_LIST.findIndex((a) => a.id === params.alertId);
    if (idx >= 0) {
      MOCK_ALERTS_LIST[idx] = {
        ...MOCK_ALERTS_LIST[idx],
        ...(body.name !== undefined ? { name: body.name as string } : {}),
        ...(body.condition !== undefined ? { condition: body.condition as string } : {}),
        ...(body.status !== undefined ? { status: body.status as 'active' | 'inactive' } : {}),
      };
      const merged = MOCK_ALERTS_LIST[idx];
      return HttpResponse.json<ApiResponse<AlertDetailData>>({
        code: 200, message: 'success',
        data: {
          id: merged.id,
          name: merged.name,
          sql: (body.sql as string) || '',
          condition: merged.condition,
          recipients: (body.recipients as string[]) || [],
          description: (body.description as string) || '',
          status: merged.status,
          updated_at: new Date().toISOString(),
        },
      });
    }
    return HttpResponse.json<ApiResponse>({ code: 40401, message: '记录不存在' }, { status: 404 });
  }),

  // DELETE /api/v1/alerts/:alert_id (API-026)
  http.delete(`${BASE}/alerts/:alertId`, async ({ params }) => {
    await delay(200);
    MOCK_ALERTS_LIST = MOCK_ALERTS_LIST.filter((a) => a.id !== params.alertId);
    return HttpResponse.json<ApiResponse>({ code: 200, message: 'success' });
  }),

  // GET /api/v1/push-records (API-027)
  http.get(`${BASE}/push-records`, async ({ request }) => {
    await delay(300);
    const url = new URL(request.url);
    const page = Number(url.searchParams.get('page') || 1);
    const pageSize = Number(url.searchParams.get('page_size') || 20);
    const pushType = url.searchParams.get('push_type') || '';
    const pushStatus = url.searchParams.get('status') || '';

    let filtered = [...MOCK_PUSH_RECORDS];
    if (pushType) filtered = filtered.filter((r) => r.push_type === pushType);
    if (pushStatus) filtered = filtered.filter((r) => r.status === pushStatus);

    const start = (page - 1) * pageSize;
    const paged = filtered.slice(start, start + pageSize);

    return HttpResponse.json<ApiResponse<PushRecordListData>>({
      code: 200, message: 'success',
      data: { records: paged, total: filtered.length, page, page_size: pageSize },
    });
  }),

  // POST /api/v1/feedbacks (PRD F13: 满意/不满意反馈)
  http.post(`${BASE}/feedbacks`, async ({ request }) => {
    await delay(200);
    const body = (await request.json()) as { rating: string; reason?: string };
    if (body.rating === 'dissatisfied' && !body.reason) {
      return HttpResponse.json<ApiResponse>({
        code: 40001,
        message: '请填写不满意原因',
      }, { status: 400 });
    }
    return HttpResponse.json<ApiResponse>({ code: 200, message: 'success' });
  }),

  // PUT /api/v1/profile/password (PRD F15: 修改密码)
  http.put(`${BASE}/profile/password`, async ({ request }) => {
    await delay(300);
    const body = (await request.json()) as ChangePasswordRequest;
    const MOCK_PASSWORDS = ['admin123', 'user123', 'pass123'];
    if (!MOCK_PASSWORDS.includes(body.old_password)) {
      return HttpResponse.json<ApiResponse>({
        code: 40001,
        message: '当前密码错误，请重新输入',
      }, { status: 400 });
    }
    if (body.new_password.length < 6 || body.new_password.length > 20) {
      return HttpResponse.json<ApiResponse>({
        code: 40001,
        message: '密码长度须在6~20位之间',
      }, { status: 400 });
    }
    if (body.new_password !== body.confirm_password) {
      return HttpResponse.json<ApiResponse>({
        code: 40001,
        message: '两次输入的密码不一致',
      }, { status: 400 });
    }
    return HttpResponse.json<ApiResponse>({ code: 200, message: 'success' });
  }),

  // POST /api/v1/profile/avatar (PRD F15: 头像上传)
  http.post(`${BASE}/profile/avatar`, async () => {
    await delay(500);
    return HttpResponse.json<ApiResponse<{ avatar_url: string }>>({
      code: 200,
      message: 'success',
      data: {
        avatar_url: `https://mock.avatar.com/${Date.now()}.jpg`,
      },
    });
  }),

  // GET /api/v1/groups/:groupId/members (用户组成员列表)
  http.get(`${BASE}/groups/:groupId/members`, async ({ params }) => {
    await delay(200);
    const members = MOCK_GROUP_MEMBERS[params.groupId as string] || [];
    return HttpResponse.json<ApiResponse<GroupMemberListData>>({
      code: 200, message: 'success',
      data: { members },
    });
  }),

  // POST /api/v1/groups/:groupId/members (添加成员)
  http.post(`${BASE}/groups/:groupId/members`, async ({ params, request }) => {
    await delay(300);
    const body = (await request.json()) as { user_id: string };
    const groupId = params.groupId as string;
    if (!MOCK_GROUP_MEMBERS[groupId]) {
      MOCK_GROUP_MEMBERS[groupId] = [];
    }
    if (MOCK_GROUP_MEMBERS[groupId].some((m) => m.id === body.user_id)) {
      return HttpResponse.json<ApiResponse>({
        code: 40901, message: '该用户已是组成员',
      }, { status: 409 });
    }
    MOCK_GROUP_MEMBERS[groupId].push({
      id: body.user_id,
      username: `user_${body.user_id.slice(0, 4)}`,
      email: `${body.user_id}@jiufang.com`,
      role: 'manager',
    });
    const groupIdx = MOCK_GROUPS_LIST.findIndex((g) => g.id === groupId);
    if (groupIdx >= 0) MOCK_GROUPS_LIST[groupIdx].member_count++;
    return HttpResponse.json<ApiResponse>({ code: 200, message: 'success' });
  }),

  // DELETE /api/v1/groups/:groupId/members/:userId (移除成员)
  http.delete(`${BASE}/groups/:groupId/members/:userId`, async ({ params }) => {
    await delay(200);
    const groupId = params.groupId as string;
    const userId = params.userId as string;
    if (MOCK_GROUP_MEMBERS[groupId]) {
      MOCK_GROUP_MEMBERS[groupId] = MOCK_GROUP_MEMBERS[groupId].filter((m) => m.id !== userId);
    }
    const groupIdx = MOCK_GROUPS_LIST.findIndex((g) => g.id === groupId);
    if (groupIdx >= 0 && MOCK_GROUPS_LIST[groupIdx].member_count > 0) MOCK_GROUPS_LIST[groupIdx].member_count--;
    return HttpResponse.json<ApiResponse>({ code: 200, message: 'success' });
  }),
];
