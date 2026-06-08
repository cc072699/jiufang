// ============================================================
// API 函数封装 - 严格对照 detailed-design.md 28 个接口
// ============================================================

import apiClient from './client';
import type {
  QueryRequest,
  QueryResultData,
  LoginRequest,
  LoginData,
  CreateUserRequest,
  UpdateUserRequest,
  UserListData,
  UserListParams,
  UserDetailData,
  CreateGroupRequest,
  UpdateGroupRequest,
  GroupListData,
  GroupListParams,
  GroupDetailData,
  ConfigurePermissionRequest,
  ConfigurePermissionData,
  HistoryListData,
  HistoryListParams,
  HistoryDetailData,
  CreateFavoriteRequest,
  FavoriteListData,
  FavoriteListParams,
  FavoriteRecord,
  CreateReportRequest,
  UpdateReportRequest,
  ReportListData,
  ReportListParams,
  ReportDetailData,
  CreateAlertRequest,
  UpdateAlertRequest,
  AlertListData,
  AlertListParams,
  AlertDetailData,
  PushRecordListData,
  PushRecordListParams,
  LogListData,
  LogListParams,
  SubmitFeedbackRequest,
  ExportQueryRequest,
  ExportQueryResult,
  ChangePasswordRequest,
  AvatarUploadResponse,
  ProfileData,
  GroupMemberListData,
} from './types';
import type { ApiResponse } from './types';

function getData<T>(response: { data: ApiResponse<T> }): T {
  return response.data.data!;
}

// API-001: 自然语言查询
export async function queryNaturalLanguage(params: QueryRequest) {
  const res = await apiClient.post<ApiResponse<QueryResultData>>('/query', {
    execute_immediately: true,
    ...params,
  });
  return getData(res);
}

// API-002: 登录
export async function login(params: LoginRequest) {
  const res = await apiClient.post<ApiResponse<LoginData>>('/auth/login', params);
  return getData(res);
}

// API-003: 登出
export async function logout() {
  await apiClient.post<ApiResponse<null>>('/auth/logout');
}

// GET /profile: 获取当前用户完整资料 (含 email, avatar)
export async function getProfile() {
  const res = await apiClient.get<ApiResponse<ProfileData>>('/profile');
  return getData(res);
}

// API-004: 创建用户
export async function createUser(params: CreateUserRequest) {
  const res = await apiClient.post<ApiResponse<UserDetailData>>('/users', params);
  return getData(res);
}

// API-005: 查询用户列表
export async function getUserList(params?: UserListParams) {
  const res = await apiClient.get<ApiResponse<UserListData>>('/users', { params });
  return getData(res);
}

// API-006: 更新用户
export async function updateUser(userId: string, params: UpdateUserRequest) {
  const res = await apiClient.put<ApiResponse<UserDetailData>>(`/users/${userId}`, params);
  return getData(res);
}

// API-007: 删除用户
export async function deleteUser(userId: string) {
  await apiClient.delete<ApiResponse<null>>(`/users/${userId}`);
}

// API-008: 创建用户组
export async function createGroup(params: CreateGroupRequest) {
  const res = await apiClient.post<ApiResponse<GroupDetailData>>('/groups', params);
  return getData(res);
}

// API-009: 查询用户组列表
export async function getGroupList(params?: GroupListParams) {
  const res = await apiClient.get<ApiResponse<GroupListData>>('/groups', { params });
  return getData(res);
}

// API-010: 更新用户组
export async function updateGroup(groupId: string, params: UpdateGroupRequest) {
  const res = await apiClient.put<ApiResponse<GroupDetailData>>(`/groups/${groupId}`, params);
  return getData(res);
}

// API-011: 删除用户组
export async function deleteGroup(groupId: string) {
  await apiClient.delete<ApiResponse<null>>(`/groups/${groupId}`);
}

// API-012: 配置用户组权限
export async function configurePermission(groupId: string, params: ConfigurePermissionRequest) {
  const res = await apiClient.post<ApiResponse<ConfigurePermissionData>>(
    `/groups/${groupId}/permissions`,
    params,
  );
  return getData(res);
}

// API-012 补充: 查询用户组权限 (后端 OpenAPI 已定义 GET /groups/{id}/permissions)
export async function getPermissions(groupId: string) {
  const res = await apiClient.get<ApiResponse<ConfigurePermissionData>>(
    `/groups/${groupId}/permissions`,
  );
  return getData(res);
}

// API-013: 查询历史列表
export async function getHistoryList(params?: HistoryListParams) {
  const res = await apiClient.get<ApiResponse<HistoryListData>>('/history', { params });
  return getData(res);
}

// API-014: 查询历史详情
export async function getHistoryDetail(recordId: string) {
  const res = await apiClient.get<ApiResponse<HistoryDetailData>>(`/history/${recordId}`);
  return getData(res);
}

// 按 session_id 获取完整对话历史
export async function getHistoryBySessionID(sessionId: string) {
  const res = await apiClient.get<ApiResponse<HistoryDetailData[]>>(`/history/session/${sessionId}`);
  return getData(res);
}

// API-015: 删除查询历史
export async function deleteHistory(recordId: string) {
  await apiClient.delete<ApiResponse<null>>(`/history/${recordId}`);
}

// API-016: 创建收藏
export async function createFavorite(params: CreateFavoriteRequest) {
  const res = await apiClient.post<ApiResponse<FavoriteRecord>>('/favorites', params);
  return getData(res);
}

// API-017: 查询收藏列表
export async function getFavoriteList(params?: FavoriteListParams) {
  const res = await apiClient.get<ApiResponse<FavoriteListData>>('/favorites', { params });
  return getData(res);
}

// API-018: 删除收藏
export async function deleteFavorite(favoriteId: string) {
  await apiClient.delete<ApiResponse<null>>(`/favorites/${favoriteId}`);
}

// API-019: 创建定时报告
export async function createReport(params: CreateReportRequest) {
  const res = await apiClient.post<ApiResponse<ReportDetailData>>('/reports', params);
  return getData(res);
}

// API-020: 查询定时报告列表
export async function getReportList(params?: ReportListParams) {
  const res = await apiClient.get<ApiResponse<ReportListData>>('/reports', { params });
  return getData(res);
}

// API-020 补充: 查询定时报告详情 (后端已实现 GET /reports/{id})
export async function getReportDetail(reportId: string) {
  const res = await apiClient.get<ApiResponse<ReportDetailData>>(`/reports/${reportId}`);
  return getData(res);
}

// API-021: 更新定时报告
export async function updateReport(reportId: string, params: UpdateReportRequest) {
  const res = await apiClient.put<ApiResponse<ReportDetailData>>(`/reports/${reportId}`, params);
  return getData(res);
}

// API-022: 删除定时报告
export async function deleteReport(reportId: string) {
  await apiClient.delete<ApiResponse<null>>(`/reports/${reportId}`);
}

// API-023: 创建预警规则
export async function createAlert(params: CreateAlertRequest) {
  const res = await apiClient.post<ApiResponse<AlertDetailData>>('/alerts', params);
  return getData(res);
}

// API-024: 查询预警规则列表
export async function getAlertList(params?: AlertListParams) {
  const res = await apiClient.get<ApiResponse<AlertListData>>('/alerts', { params });
  return getData(res);
}

// API-024 补充: 查询预警规则详情 (后端已实现 GET /alerts/{id})
export async function getAlertDetail(alertId: string) {
  const res = await apiClient.get<ApiResponse<AlertDetailData>>(`/alerts/${alertId}`);
  return getData(res);
}

// API-025: 更新预警规则
export async function updateAlert(alertId: string, params: UpdateAlertRequest) {
  const res = await apiClient.put<ApiResponse<AlertDetailData>>(`/alerts/${alertId}`, params);
  return getData(res);
}

// API-026: 删除预警规则
export async function deleteAlert(alertId: string) {
  await apiClient.delete<ApiResponse<null>>(`/alerts/${alertId}`);
}

// API-027: 查询推送记录
export async function getPushRecordList(params?: PushRecordListParams) {
  const res = await apiClient.get<ApiResponse<PushRecordListData>>('/push-records', { params });
  return getData(res);
}

// API-028: 查询操作日志
export async function getLogList(params?: LogListParams) {
  const res = await apiClient.get<ApiResponse<LogListData>>('/logs', { params });
  return getData(res);
}

// 提交查询反馈 (PRD F13)
export async function submitFeedback(params: SubmitFeedbackRequest) {
  await apiClient.post<ApiResponse<null>>('/feedbacks', params);
}

// 导出查询结果 (PRD F14)
export async function exportQueryResult(params: ExportQueryRequest) {
  const res = await apiClient.post<ApiResponse<ExportQueryResult>>('/export', params);
  return getData(res);
}

// 修改密码 (PRD F15)
export async function changePassword(params: ChangePasswordRequest) {
  const res = await apiClient.put<ApiResponse<null>>('/profile/password', params);
  return getData(res);
}

// 上传头像 (PRD F15)
export async function uploadAvatar(file: File) {
  const formData = new FormData();
  formData.append('avatar', file);
  const res = await apiClient.post<ApiResponse<AvatarUploadResponse>>('/profile/avatar', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return getData(res);
}

// 获取用户组成员 (PRD F09)
export async function getGroupMembers(groupId: string) {
  const res = await apiClient.get<ApiResponse<GroupMemberListData>>(`/groups/${groupId}/members`);
  return getData(res);
}

// 添加用户组成员
export async function addGroupMember(groupId: string, userId: string) {
  await apiClient.post<ApiResponse<null>>(`/groups/${groupId}/members`, { user_id: userId });
}

// 移除用户组成员
export async function removeGroupMember(groupId: string, userId: string) {
  await apiClient.delete<ApiResponse<null>>(`/groups/${groupId}/members/${userId}`);
}

// ERP 元数据：获取表结构列表
export interface TableColumnInfo {
  name: string;
  type: string;
  comment: string;
}
export interface TableSchemaInfo {
  name: string;
  label: string;
  columns: TableColumnInfo[];
}
export async function getERPTableSchemas(): Promise<TableSchemaInfo[]> {
  const res = await apiClient.get<ApiResponse<TableSchemaInfo[]>>('/metadata/tables');
  return getData(res);
}
