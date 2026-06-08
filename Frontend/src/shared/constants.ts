// ============================================================
// 统一常量定义 - 消除各页面重复硬编码
// ============================================================

// --- 角色 ---
export const ROLE_LABELS: Record<string, string> = {
  admin: '管理员',
  manager: '部门经理',
  executive: '高管',
};

export const ROLE_COLORS: Record<string, string> = {
  admin: 'red',
  manager: 'blue',
  executive: 'green',
};

// --- 状态 (启用/停用) ---
export const STATUS_LABELS: Record<string, string> = {
  active: '启用',
  inactive: '停用',
};

export const STATUS_COLORS: Record<string, string> = {
  active: 'green',
  inactive: 'default',
};

// --- 查询状态 (成功/失败) ---
export const QUERY_STATUS_MAP: Record<string, { color: string; label: string }> = {
  success: { color: 'green', label: '成功' },
  failed: { color: 'red', label: '失败' },
};

// --- 用户状态 (正常/停用, number 类型) ---
export const USER_STATUS_LABELS: Record<number, string> = {
  1: '正常',
  0: '停用',
};

export const USER_STATUS_COLORS: Record<number, string> = {
  1: 'green',
  0: 'default',
};

// --- 操作日志类型 ---
export const OP_TYPE_LABELS: Record<string, string> = {
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

export const OP_TYPE_COLORS: Record<string, string> = {
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

export const OP_FILTER_OPTIONS = [
  { label: '登录', value: 'login' },
  { label: '登出', value: 'logout' },
  { label: '查询', value: 'query' },
  { label: '创建', value: 'create' },
  { label: '更新', value: 'update' },
  { label: '删除', value: 'delete' },
];

// --- 定时报告频率 ---
export const SCHEDULE_LABELS: Record<string, string> = {
  daily: '每日',
  weekly: '每周',
  monthly: '每月',
};

// --- 推送类型 ---
export const PUSH_TYPE_LABELS: Record<string, string> = {
  report: '报告',
  alert: '预警',
};

export const PUSH_TYPE_COLORS: Record<string, string> = {
  report: 'blue',
  alert: 'orange',
};

// --- 推送状态 ---
export const PUSH_STATUS_LABELS: Record<string, string> = {
  success: '成功',
  failed: '失败',
  retrying: '重试中',
};

export const PUSH_STATUS_COLORS: Record<string, string> = {
  success: 'green',
  failed: 'red',
  retrying: 'orange',
};

// --- 角色筛选选项 (Select/Radio) ---
export const ROLE_FILTER_OPTIONS = [
  { label: '全部角色', value: undefined },
  { label: '管理员', value: 'admin' },
  { label: '部门经理', value: 'manager' },
  { label: '高管', value: 'executive' },
];

export const ROLE_SELECT_OPTIONS = [
  { label: '管理员', value: 'admin' },
  { label: '部门经理', value: 'manager' },
  { label: '高管', value: 'executive' },
];

// --- 用户状态筛选选项 ---
export const USER_STATUS_FILTER_OPTIONS = [
  { label: '全部状态', value: undefined },
  { label: '正常', value: 1 },
  { label: '停用', value: 0 },
];

export const USER_STATUS_SELECT_OPTIONS = [
  { label: '正常', value: 1 },
  { label: '停用', value: 0 },
];

// --- 启用/停用筛选选项 ---
export const ACTIVE_STATUS_FILTER_OPTIONS = [
  { label: '全部', value: undefined },
  { label: '启用', value: 'active' },
  { label: '停用', value: 'inactive' },
];

export const ACTIVE_STATUS_SELECT_OPTIONS = [
  { label: '启用', value: 'active' },
  { label: '停用', value: 'inactive' },
];

// --- 查询状态筛选选项 ---
export const QUERY_STATUS_FILTER_OPTIONS = [
  { label: '全部', value: undefined },
  { label: '成功', value: 'success' },
  { label: '失败', value: 'failed' },
];

// --- 定时报告频率筛选选项 ---
export const SCHEDULE_SELECT_OPTIONS = [
  { label: '每日', value: 'daily' },
  { label: '每周', value: 'weekly' },
  { label: '每月', value: 'monthly' },
];

// --- 推送类型筛选选项 ---
export const PUSH_TYPE_FILTER_OPTIONS = [
  { label: '全部', value: undefined },
  { label: '报告', value: 'report' },
  { label: '预警', value: 'alert' },
];

// --- 推送状态筛选选项 ---
export const PUSH_STATUS_FILTER_OPTIONS = [
  { label: '全部', value: undefined },
  { label: '成功', value: 'success' },
  { label: '失败', value: 'failed' },
];

// --- 推送渠道 ---
export const PUSH_CHANNEL_LABELS: Record<string, string> = {
  wechat: '企业微信',
  email: '邮箱',
};

export const PUSH_CHANNEL_OPTIONS = [
  { label: '邮箱', value: 'email' },
  { label: '企业微信', value: 'wechat' },
];
