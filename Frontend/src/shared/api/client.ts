// ============================================================
// API Client - Axios 封装 + Token 注入 + 错误拦截
// ============================================================

import axios from 'axios';
import type { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { getToken } from '../auth/token';
import { AUTH_ERROR_CODES, PERMISSION_ERROR_CODES, LOGIN_ERROR_CODES, ERROR_CODES } from './types';
import type { ApiResponse } from './types';

const ID_FIELDS = new Set(['id', 'user_id', 'group_id', 'record_id', 'source_id', 'session_id']);

function isIdField(key: string): boolean {
  return ID_FIELDS.has(key) || (key.endsWith('_id') && key.length > 3);
}

function normalizeIds(data: unknown): unknown {
  if (Array.isArray(data)) return data.map(normalizeIds);
  if (data !== null && typeof data === 'object') {
    const result: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(data as Record<string, unknown>)) {
      result[key] = isIdField(key) && typeof value === 'number' ? String(value) : normalizeIds(value);
    }
    return result;
  }
  return data;
}

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
});

// 请求拦截：注入 Token
apiClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = getToken();
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 业务错误类
export class ApiError extends Error {
  code: number;
  constructor(code: number, message: string) {
    super(message);
    this.code = code;
    this.name = 'ApiError';
  }
}

// 401 回调
let onAuthError: (() => void) | null = null;
export function setAuthErrorHandler(handler: () => void) {
  onAuthError = handler;
}

// 403 回调
let onPermissionError: ((msg: string) => void) | null = null;
export function setPermissionErrorHandler(handler: (msg: string) => void) {
  onPermissionError = handler;
}

// 响应拦截：统一处理业务错误
apiClient.interceptors.response.use(
  (response) => {
    const body = response.data as ApiResponse;
    if (body.code !== 200) {
      // 登录接口错误码由 LoginPage 自行处理，使用后端返回的 message
      const isLoginRequest = response.config.url?.includes('/auth/login');
      const msg = (isLoginRequest && LOGIN_ERROR_CODES.includes(body.code))
        ? body.message
        : (ERROR_CODES[body.code] || body.message || '请求失败');

      // 401 跳转 - 仅令牌过期(40102)触发全局退出
      if (AUTH_ERROR_CODES.includes(body.code) && !isLoginRequest) {
        onAuthError?.();
      }

      // 403 提示但不清理登录态 - 登录接口由 LoginPage 自行处理
      if (PERMISSION_ERROR_CODES.includes(body.code) && !isLoginRequest) {
        onPermissionError?.(msg);
      }

      throw new ApiError(body.code, msg);
    }
    if (body.data !== undefined) {
      body.data = normalizeIds(body.data) as typeof body.data;
    }
    return response;
  },
  (error: AxiosError<ApiResponse>) => {
    if (error.response?.data) {
      const body = error.response.data;
      const msg = ERROR_CODES[body.code] || body.message || '网络错误';

      if (body.code && AUTH_ERROR_CODES.includes(body.code) && !error.response?.config?.url?.includes('/auth/login')) {
        onAuthError?.();
      }

      if (body.code && PERMISSION_ERROR_CODES.includes(body.code) && !error.response?.config?.url?.includes('/auth/login')) {
        onPermissionError?.(msg);
      }

      throw new ApiError(body.code || error.response.status, msg);
    }
    if (error.code === 'ECONNABORTED') {
      throw new ApiError(50301, '请求超时，请稍后重试');
    }
    throw new ApiError(0, '网络连接失败，请检查网络');
  },
);

export default apiClient;
