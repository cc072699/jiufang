// ============================================================
// Token 存储工具 - 使用 localStorage
// ============================================================

const TOKEN_KEY = 'jiufang_token';
const USER_KEY = 'jiufang_user';
const EXPIRES_KEY = 'jiufang_token_expires';

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
  localStorage.removeItem(EXPIRES_KEY);
}

export function getTokenExpiresAt(): string | null {
  return localStorage.getItem(EXPIRES_KEY);
}

export function setTokenExpiresAt(expiresAt: string): void {
  localStorage.setItem(EXPIRES_KEY, expiresAt);
}

export function isTokenExpired(): boolean {
  const raw = getTokenExpiresAt();
  if (!raw) return false; // 无过期时间则不强制过期（兼容旧数据）
  return new Date(raw).getTime() <= Date.now();
}

export function getStoredUser(): import('../api/types').UserInfo | null {
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

export function setStoredUser(user: import('../api/types').UserInfo): void {
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}
