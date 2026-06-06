// ============================================================
// App Providers 组合
// ============================================================

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ConfigProvider, App as AntApp, message } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { RouterProvider } from 'react-router-dom';
import { useEffect } from 'react';
import { router } from './router';
import { useAuthStore } from '../shared/auth/store';
import { setAuthErrorHandler, setPermissionErrorHandler } from '../shared/api/client';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 0,
      refetchOnWindowFocus: false,
      staleTime: 30_000,
    },
  },
});

export default function AppProviders() {
  const clearAuth = useAuthStore((s) => s.clearAuth);

  useEffect(() => {
    setAuthErrorHandler(() => {
      queryClient.clear();
      clearAuth();
    });
    setPermissionErrorHandler((msg) => message.warning(msg));
  }, [clearAuth]);

  return (
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={zhCN}>
        <AntApp>
          <RouterProvider router={router} />
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>
  );
}