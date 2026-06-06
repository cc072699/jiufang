import { StrictMode } from 'react';
import AppProviders from './app/providers';
import { useAuthStore } from './shared/auth/store';

async function bootstrap() {
  // MSW 已禁用，准备连接真实后端
  // 如需重新启用，取消下面的注释
  /*
  if (import.meta.env.DEV) {
    try {
      const { worker } = await import('./mocks/browser');
      const msw = await import('msw');
      await worker.start({ onUnhandledRequest: 'bypass' });
      (window as unknown as Record<string, unknown>).__msw_worker = worker;
      (window as unknown as Record<string, unknown>).__msw_http = msw.http;
      (window as unknown as Record<string, unknown>).__msw_httpResponse = msw.HttpResponse;
    } catch (e) {
      console.warn('[MSW] 启动失败，Mock 不可用:', e);
    }
  }
  */

  useAuthStore.getState().loadFromStorage();

  const { createRoot } = await import('react-dom/client');
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <AppProviders />
    </StrictMode>,
  );
}

bootstrap();
