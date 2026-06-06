// ============================================================
// Zustand Auth Store - 登录态管理
// ============================================================

import { create } from 'zustand';
import type { UserInfo } from '../api/types';
import {
  getStoredUser,
  getToken,
  removeToken,
  setStoredUser,
  setToken,
  setTokenExpiresAt,
  isTokenExpired,
} from '../auth/token';

interface AuthState {
  user: UserInfo | null;
  token: string | null;
  isAuthenticated: boolean;
  isAdmin: boolean;

  setAuth: (token: string, user: UserInfo, expiresAt?: string) => void;
  clearAuth: () => void;
  loadFromStorage: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  isAuthenticated: false,
  isAdmin: false,

  setAuth: (token: string, user: UserInfo, expiresAt?: string) => {
    setToken(token);
    setStoredUser(user);
    if (expiresAt) setTokenExpiresAt(expiresAt);
    set({
      token,
      user,
      isAuthenticated: true,
      isAdmin: user.role === 'admin',
    });
  },

  clearAuth: () => {
    removeToken();
    set({
      token: null,
      user: null,
      isAuthenticated: false,
      isAdmin: false,
    });
  },

  loadFromStorage: () => {
    const token = getToken();
    const user = getStoredUser();
    if (token && user) {
      if (isTokenExpired()) {
        removeToken();
        return;
      }
      set({
        token,
        user,
        isAuthenticated: true,
        isAdmin: user.role === 'admin',
      });
    }
  },
}));