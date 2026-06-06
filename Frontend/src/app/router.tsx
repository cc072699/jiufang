// ============================================================
// App Router - 路由配置
// ============================================================

import { createBrowserRouter, Navigate } from 'react-router-dom';
import AppLayout from './layout/AppLayout';
import { AuthGuard, GuestGuard, AdminGuard } from '../shared/auth/guards';
import LoginPage from '../pages/LoginPage';
import WorkbenchPage from '../pages/WorkbenchPage';
import QueryPage from '../pages/QueryPage';
import HistoryPage from '../pages/HistoryPage';
import FavoritesPage from '../pages/FavoritesPage';
import PermissionPage from '../pages/PermissionPage';
import LogsPage from '../pages/LogsPage';
import ReportsPage from '../pages/ReportsPage';
import ReportFormPage from '../pages/ReportFormPage';
import AlertsPage from '../pages/AlertsPage';
import AlertFormPage from '../pages/AlertFormPage';
import UsersPage from '../pages/UsersPage';
import UserFormPage from '../pages/UserFormPage';
import ProfilePage from '../pages/ProfilePage';
import PushRecordsPage from '../pages/PushRecordsPage';

export const router = createBrowserRouter([
  {
    path: '/login',
    element: (
      <GuestGuard>
        <LoginPage />
      </GuestGuard>
    ),
  },
  {
    path: '/',
    element: (
      <AuthGuard>
        <AppLayout />
      </AuthGuard>
    ),
    children: [
      { index: true, element: <Navigate to="/workbench" replace /> },
      { path: 'workbench', element: <WorkbenchPage /> },
      { path: 'query', element: <QueryPage /> },
      { path: 'history', element: <HistoryPage /> },
      { path: 'favorites', element: <FavoritesPage /> },
      {
        path: 'permissions',
        element: (
          <AdminGuard>
            <PermissionPage />
          </AdminGuard>
        ),
      },
      {
        path: 'logs',
        element: (
          <AdminGuard>
            <LogsPage />
          </AdminGuard>
        ),
      },
      {
        path: 'reports',
        element: (
          <AdminGuard>
            <ReportsPage />
          </AdminGuard>
        ),
      },
      {
        path: 'reports/new',
        element: (
          <AdminGuard>
            <ReportFormPage />
          </AdminGuard>
        ),
      },
      {
        path: 'reports/:id/edit',
        element: (
          <AdminGuard>
            <ReportFormPage />
          </AdminGuard>
        ),
      },
      {
        path: 'alerts',
        element: (
          <AdminGuard>
            <AlertsPage />
          </AdminGuard>
        ),
      },
      {
        path: 'alerts/new',
        element: (
          <AdminGuard>
            <AlertFormPage />
          </AdminGuard>
        ),
      },
      {
        path: 'alerts/:id/edit',
        element: (
          <AdminGuard>
            <AlertFormPage />
          </AdminGuard>
        ),
      },
      {
        path: 'users',
        element: (
          <AdminGuard>
            <UsersPage />
          </AdminGuard>
        ),
      },
      {
        path: 'users/new',
        element: (
          <AdminGuard>
            <UserFormPage />
          </AdminGuard>
        ),
      },
      {
        path: 'users/:id/edit',
        element: (
          <AdminGuard>
            <UserFormPage />
          </AdminGuard>
        ),
      },
      { path: 'profile', element: <ProfilePage /> },
      {
        path: 'push-records',
        element: (
          <AdminGuard>
            <PushRecordsPage />
          </AdminGuard>
        ),
      },
    ],
  },
]);