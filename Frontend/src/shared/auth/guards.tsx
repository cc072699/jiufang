// ============================================================
// 路由守卫组件
// ============================================================

import React, { useState, useEffect } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { message } from 'antd';
import { useAuthStore } from './store';

/** 需要登录才能访问，未登录跳转 /login */
export function AuthGuard({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  return <>{children}</>;
}

/** 已登录用户访问 /login 时跳转工作台 */
export function GuestGuard({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  if (isAuthenticated) {
    return <Navigate to="/workbench" replace />;
  }
  return <>{children}</>;
}

/** 仅 admin 可访问，非 admin 展示 403 提示并跳转工作台 */
export function AdminGuard({ children }: { children: React.ReactNode }) {
  const { user } = useAuthStore();
  const isAdmin = user?.role === 'admin';
  const [shouldRedirect, setShouldRedirect] = useState(false);

  useEffect(() => {
    if (!isAdmin) {
      message.warning('权限不足，无法访问该页面');
      setShouldRedirect(true);
    }
  }, [isAdmin]);

  if (shouldRedirect) {
    return <Navigate to="/workbench" replace />;
  }

  return <>{children}</>;
}