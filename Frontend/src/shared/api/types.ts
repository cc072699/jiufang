// ============================================================
// API 类型定义 - 严格对照 detailed-design.md 接口规格
// ============================================================

// --- 通用响应 ---
export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data?: T;
}

// --- 分页 ---
export interface PaginatedData {
  total: number;
  page: number;
  page_size: number;
}

// --- API-001: 自然语言查询 ---
export interface QueryRequest {
  input: string;
  session_id?: string;
  execute_immediately?: boolean;
}

export interface ColumnDef {
  name: string;
  type: 'string' | 'number' | 'date';
}

export interface QueryResultData {
  session_id: string;
  understanding: string;
  result_type: 'table' | 'chart' | 'empty';
  columns?: ColumnDef[];
  rows?: Record<string, unknown>[];
  chart_config?: ChartConfig;
  suggested_questions?: string[];
  can_export: boolean;
  sql?: string; // 非设计文档定义，为收藏功能预留；联调后端需确认是否返回
}

export interface ChartConfig {
  type: 'bar_chart' | 'line_chart' | 'pie_chart' | 'table';
  title: string;
  x_axis?: {
    type: string;
    name: string;
    unit?: string;
    data?: string[];
  };
  y_axis?: {
    type: string;
    name: string;
    unit?: string;
  };
  series?: {
    name: string;
    type: string;
    data: unknown[];
    format?: string;
    color?: string;
    label?: { show: boolean; position?: string; formatter?: string };
  }[];
  legend?: { show: boolean; position: string; data?: string[] };
  tooltip?: { show: boolean; trigger: string; formatter?: string };
  colors?: string[];
  data_limit?: number;
  empty_value_hint?: string;
  can_switch_type?: boolean;
}

// --- API-002: 登录 ---
export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginData {
  token: string;
  expires_at: string;
  user: UserInfo;
}

export interface UserInfo {
  id: string;
  username: string;
  role: 'admin' | 'manager' | 'executive';
  groups: string[];
  email?: string; // 非 API-002 登录响应定义，PermissionPage 成员展示使用
}

// --- GET /profile: 用户完整资料 (设计文档定义) ---
export interface ProfileData {
  id: string;
  username: string;
  email: string;
  role: 'admin' | 'manager' | 'executive';
  avatar?: string;
  groups: string[];
  status: number;
  created_at: string;
}

// --- API-004~007: 用户 CRUD ---
export interface CreateUserRequest {
  username: string;
  password: string;
  email: string;
  role: 'admin' | 'manager' | 'executive';
  groups?: string[];
  status?: number;
}

export interface UpdateUserRequest {
  username?: string;
  password?: string;
  email?: string;
  role?: 'admin' | 'manager' | 'executive';
  groups?: string[];
  status?: number;
}

export interface UserRecord {
  id: string;
  username: string;
  email: string;
  role: 'admin' | 'manager' | 'executive';
  status: number;
  created_at: string;
}

export interface UserListData extends PaginatedData {
  users: UserRecord[];
}

export interface UserDetailData extends UserRecord {
  groups: string[];
  updated_at?: string;
}

// --- API-008~011: 用户组 CRUD ---
export interface CreateGroupRequest {
  name: string;
  description?: string;
  members?: string[];
}

export interface UpdateGroupRequest {
  name?: string;
  description?: string;
  members?: string[];
}

export interface GroupRecord {
  id: string;
  name: string;
  description?: string;
  member_count: number;
  created_at: string;
}

export interface GroupListData extends PaginatedData {
  groups: GroupRecord[];
}

export interface GroupDetailData {
  id: string;
  name: string;
  description?: string;
  members: GroupMemberItem[];
  created_at?: string;
  updated_at?: string;
}

// --- API-012: 权限配置 ---
export interface PermissionItem {
  table_name: string;
  allowed_fields: string;
  filter_condition?: string;
}

export interface ConfigurePermissionRequest {
  permissions: PermissionItem[];
}

export interface PermissionRecord {
  id: string;
  table_name: string;
  allowed_fields: string;
  filter_condition?: string;
}

export interface ConfigurePermissionData {
  group_id: string;
  permissions: PermissionRecord[];
  created_at?: string;
}

// --- API-013~015: 查询历史 ---
export interface HistoryRecord {
  id: string;
  input: string;
  sql: string;
  status: 'success' | 'failed';
  error_message?: string;
  result_count?: number;
  execution_time?: number;
  created_at: string;
}

export interface HistoryListData extends PaginatedData {
  records: HistoryRecord[];
}

export interface HistoryDetailData extends HistoryRecord {
  session_id: string;
  result_data?: string;
}

// --- API-016~018: 收藏 ---
export interface CreateFavoriteRequest {
  name: string;
  input: string;
  sql: string;
  description?: string;
}

export interface FavoriteRecord {
  id: string;
  name: string;
  input: string;
  sql: string;
  description?: string;
  created_at: string;
}

export interface FavoriteListData extends PaginatedData {
  favorites: FavoriteRecord[];
}

// --- API-019~022: 定时报告 ---
export interface CreateReportRequest {
  name: string;
  sql: string;
  schedule_type: 'daily' | 'weekly' | 'monthly';
  schedule_time: string;
  recipients: string[];
  description?: string;
}

export interface UpdateReportRequest {
  name?: string;
  sql?: string;
  schedule_type?: 'daily' | 'weekly' | 'monthly';
  schedule_time?: string;
  recipients?: string[];
  description?: string;
  status?: 'active' | 'inactive';
}

export interface ReportRecord {
  id: string;
  name: string;
  schedule_type: 'daily' | 'weekly' | 'monthly';
  schedule_time: string;
  recipients: string[];
  recipients_count: number;
  status: 'active' | 'inactive';
  description?: string;
  sql?: string;
  last_run_at?: string;
  created_by?: string;
  created_at: string;
}

export interface ReportListData extends PaginatedData {
  reports: ReportRecord[];
}

export interface ReportDetailData {
  id: string;
  name: string;
  sql: string;
  schedule_type: 'daily' | 'weekly' | 'monthly';
  schedule_time: string;
  recipients: string[];
  description?: string;
  status: 'active' | 'inactive';
  created_by?: string;
  created_at?: string;
  updated_at?: string;
}

// --- API-023~026: 预警规则 ---
export interface CreateAlertRequest {
  name: string;
  sql: string;
  condition: string;
  recipients: string[];
  description?: string;
}

export interface UpdateAlertRequest {
  name?: string;
  sql?: string;
  condition?: string;
  recipients?: string[];
  description?: string;
  status?: 'active' | 'inactive';
}

export interface AlertRecord {
  id: string;
  name: string;
  condition: string;
  recipients: string[];
  recipients_count: number;
  status: 'active' | 'inactive';
  description?: string;
  last_triggered_at?: string;
  created_at: string;
}

export interface AlertListData extends PaginatedData {
  alerts: AlertRecord[];
}

export interface AlertDetailData {
  id: string;
  name: string;
  sql: string;
  condition: string;
  recipients: string[];
  description?: string;
  status: 'active' | 'inactive';
  created_by?: string;
  created_at?: string;
  updated_at?: string;
}

// --- API-027: 推送记录 ---
export interface PushRecord {
  id: string;
  push_type: 'report' | 'alert';
  source_id: string;
  source_name: string;
  recipient: string;
  status: 'success' | 'failed' | 'retrying';
  error_message?: string;
  pushed_at: string;
}

export interface PushRecordListData extends PaginatedData {
  records: PushRecord[];
}

// --- API-028: 操作日志 ---
export interface LogRecord {
  id: string;
  user_id: string;
  username: string;
  operation_type: string;
  operation_object: string;
  operation_detail: string;
  operation_result: string;
  ip_address: string;
  created_at: string;
}

export interface LogListData extends PaginatedData {
  logs: LogRecord[];
}

// --- 反馈提交 (PRD F13) ---
export interface SubmitFeedbackRequest {
  query_record_id: string;
  rating: 'satisfied' | 'unsatisfied';
  reason?: string;
}

// --- 修改密码 (PRD F15) ---
export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
  confirm_password: string;
}

// --- 头像上传 (PRD F15) ---
export interface AvatarUploadResponse {
  avatar_url: string;
}

// --- 用户组成员管理 ---
export interface GroupMemberItem {
  id: string;
  username: string;
  email: string;
  role: 'admin' | 'manager' | 'executive';
}

export interface GroupMemberListData {
  members: GroupMemberItem[];
}

// --- 错误码映射 (对照 detailed-design.md 接口异常定义) ---
export const ERROR_CODES: Record<number, string> = {
  40001: '参数校验失败',
  40002: '检测到不安全操作，已拒绝执行',
  40003: '无法理解问题，请换种说法',
  40101: '用户名或密码错误',
  40102: '登录已过期，请重新登录',
  40103: '账号不存在',
  40104: '账号已停用，请联系管理员',
  40105: '用户名或密码错误',
  40301: '权限不足，请联系管理员',
  40401: '记录不存在',
  40901: '用户名已存在',
  40902: '邮箱已存在',
  50301: '数据库连接超时，请稍后重试',
  50302: '智能解析服务不可用，请稍后重试',
};

// API-002 登录专用错误码 (LoginPage 自行处理，不触发全局退出)
export const LOGIN_ERROR_CODES = [401, 40101, 40103, 40104, 40105, 40301];

// 全局认证错误码 (触发自动退出登录)
export const AUTH_ERROR_CODES = [40102];
export const PERMISSION_ERROR_CODES = [40301];

// --- 列表查询参数类型 ---
export interface UserListParams {
  page?: number;
  page_size?: number;
  username?: string;
  role?: string;
  status?: number;
}

export interface GroupListParams {
  page?: number;
  page_size?: number;
  name?: string;
}

export interface HistoryListParams {
  page?: number;
  page_size?: number;
  start_time?: string;
  end_time?: string;
  status?: string;
}

export interface FavoriteListParams {
  page?: number;
  page_size?: number;
  name?: string;
}

export interface ReportListParams {
  page?: number;
  page_size?: number;
  name?: string;
  status?: string;
}

export interface AlertListParams {
  page?: number;
  page_size?: number;
  name?: string;
  status?: string;
}

export interface PushRecordListParams {
  page?: number;
  page_size?: number;
  push_type?: string;
  status?: string;
  start_time?: string;
  end_time?: string;
}

export interface LogListParams {
  page?: number;
  page_size?: number;
  user_id?: string;
  operation_type?: string;
  start_time?: string;
  end_time?: string;
}
