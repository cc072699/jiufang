import { StrictMode } from 'react';
import AppProviders from './app/providers';
import { useAuthStore } from './shared/auth/store';

async function bootstrap() {
  useAuthStore.getState().loadFromStorage();

  const { createRoot } = await import('react-dom/client');
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <AppProviders />
    </StrictMode>,
  );
}

bootstrap();
