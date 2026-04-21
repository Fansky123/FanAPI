import { lazy, Suspense } from 'react'
import type { ReactNode } from 'react'
import { createBrowserRouter, Navigate, Outlet } from 'react-router-dom'

import { getRoleToken, getSiteModePreference } from '@/lib/auth/storage'

const AuthLayout = lazy(() => import('@/layouts/AuthLayout').then((m) => ({ default: m.AuthLayout })))
const UserLayout = lazy(() => import('@/layouts/UserLayout').then((m) => ({ default: m.UserLayout })))
const AdminLayout = lazy(() => import('@/layouts/AdminLayout').then((m) => ({ default: m.AdminLayout })))
const AgentLayout = lazy(() => import('@/layouts/AgentLayout').then((m) => ({ default: m.AgentLayout })))
const VendorLayout = lazy(() => import('@/layouts/VendorLayout').then((m) => ({ default: m.VendorLayout })))

const UserLoginPage = lazy(() => import('@/pages/public/UserLoginPage').then((m) => ({ default: m.UserLoginPage })))
const RegisterPage = lazy(() => import('@/pages/public/RegisterPage').then((m) => ({ default: m.RegisterPage })))
const ForgotPasswordPage = lazy(() => import('@/pages/public/ForgotPasswordPage').then((m) => ({ default: m.ForgotPasswordPage })))
const AppErrorPage = lazy(() => import('@/pages/public/AppErrorPage').then((m) => ({ default: m.AppErrorPage })))
const NotFoundPage = lazy(() => import('@/pages/public/NotFoundPage').then((m) => ({ default: m.NotFoundPage })))

const UserDashboardPage = lazy(() => import('@/pages/user/UserDashboardPage').then((m) => ({ default: m.UserDashboardPage })))
const UserModelsPage = lazy(() => import('@/pages/user/UserModelsPage').then((m) => ({ default: m.UserModelsPage })))
const UserKeysPage = lazy(() => import('@/pages/user/UserKeysPage').then((m) => ({ default: m.UserKeysPage })))
const UserPlaygroundPage = lazy(() => import('@/pages/user/UserPlaygroundPage').then((m) => ({ default: m.UserPlaygroundPage })))
const UserImageGenPage = lazy(() => import('@/pages/user/UserImageGenPage').then((m) => ({ default: m.UserImageGenPage })))
const UserVideoGenPage = lazy(() => import('@/pages/user/UserVideoGenPage').then((m) => ({ default: m.UserVideoGenPage })))
const UserTasksPage = lazy(() => import('@/pages/user/UserTasksPage').then((m) => ({ default: m.UserTasksPage })))
const UserLogsPage = lazy(() => import('@/pages/user/UserLogsPage').then((m) => ({ default: m.UserLogsPage })))
const UserBillingPage = lazy(() => import('@/pages/user/UserBillingPage').then((m) => ({ default: m.UserBillingPage })))
const UserDocsPage = lazy(() => import('@/pages/user/UserDocsPage').then((m) => ({ default: m.UserDocsPage })))
const UserProfilePage = lazy(() => import('@/pages/user/UserProfilePage').then((m) => ({ default: m.UserProfilePage })))
const UserStatsPage = lazy(() => import('@/pages/user/UserStatsPage').then((m) => ({ default: m.UserStatsPage })))
const UserExchangePage = lazy(() => import('@/pages/user/UserExchangePage').then((m) => ({ default: m.UserExchangePage })))
const UserInvitePage = lazy(() => import('@/pages/user/UserInvitePage').then((m) => ({ default: m.UserInvitePage })))

const AdminLoginPage = lazy(() => import('@/pages/admin/AdminLoginPage').then((m) => ({ default: m.AdminLoginPage })))
const AdminDashboardPage = lazy(() => import('@/pages/admin/AdminDashboardPage').then((m) => ({ default: m.AdminDashboardPage })))
const AdminChannelsPage = lazy(() => import('@/pages/admin/AdminChannelsPage').then((m) => ({ default: m.AdminChannelsPage })))
const AdminUsersPage = lazy(() => import('@/pages/admin/AdminUsersPage').then((m) => ({ default: m.AdminUsersPage })))
const AdminBillingPage = lazy(() => import('@/pages/admin/AdminBillingPage').then((m) => ({ default: m.AdminBillingPage })))
const AdminCardsPage = lazy(() => import('@/pages/admin/AdminCardsPage').then((m) => ({ default: m.AdminCardsPage })))
const AdminKeyPoolsPage = lazy(() => import('@/pages/admin/AdminKeyPoolsPage').then((m) => ({ default: m.AdminKeyPoolsPage })))
const AdminOcpcPage = lazy(() => import('@/pages/admin/AdminOcpcPage').then((m) => ({ default: m.AdminOcpcPage })))
const AdminTasksPage = lazy(() => import('@/pages/admin/AdminTasksPage').then((m) => ({ default: m.AdminTasksPage })))
const AdminLogsPage = lazy(() => import('@/pages/admin/AdminLogsPage').then((m) => ({ default: m.AdminLogsPage })))
const AdminSettingsPage = lazy(() => import('@/pages/admin/AdminSettingsPage').then((m) => ({ default: m.AdminSettingsPage })))
const AdminVendorsPage = lazy(() => import('@/pages/admin/AdminVendorsPage').then((m) => ({ default: m.AdminVendorsPage })))
const AdminWithdrawPage = lazy(() => import('@/pages/admin/AdminWithdrawPage').then((m) => ({ default: m.AdminWithdrawPage })))

const AgentLoginPage = lazy(() => import('@/pages/agent/AgentLoginPage').then((m) => ({ default: m.AgentLoginPage })))
const AgentDashboardPage = lazy(() => import('@/pages/agent/AgentDashboardPage').then((m) => ({ default: m.AgentDashboardPage })))

const VendorLoginPage = lazy(() => import('@/pages/vendor/VendorLoginPage').then((m) => ({ default: m.VendorLoginPage })))
const VendorRegisterPage = lazy(() => import('@/pages/vendor/VendorRegisterPage').then((m) => ({ default: m.VendorRegisterPage })))
const VendorDashboardPage = lazy(() => import('@/pages/vendor/VendorDashboardPage').then((m) => ({ default: m.VendorDashboardPage })))
const VendorKeysPage = lazy(() => import('@/pages/vendor/VendorKeysPage').then((m) => ({ default: m.VendorKeysPage })))

function renderLazy(node: ReactNode) {
  return (
    <Suspense fallback={<div className="min-h-screen bg-background" />}>
      {node}
    </Suspense>
  )
}

function RequireRole({
  role,
  redirectTo,
}: {
  role: 'user' | 'admin' | 'agent' | 'vendor'
  redirectTo: string
}) {
  const token = getRoleToken(role)
  if (!token) {
    return <Navigate replace to={redirectTo} />
  }

  return <Outlet />
}

function PublicOnly({
  role,
  redirectTo,
}: {
  role: 'user' | 'admin' | 'agent' | 'vendor'
  redirectTo: string
}) {
  const token = getRoleToken(role)
  if (token) {
    return <Navigate replace to={redirectTo} />
  }

  return <Outlet />
}

function RootRedirect() {
  const mode = getSiteModePreference()
  const adminToken = getRoleToken('admin')
  const agentToken = getRoleToken('agent')
  const vendorToken = getRoleToken('vendor')

  if (mode === 'admin' && adminToken) return <Navigate replace to="/admin/dashboard" />
  if (mode === 'agent' && agentToken) return <Navigate replace to="/agent/dashboard" />
  if (mode === 'vendor' && vendorToken) return <Navigate replace to="/vendor/dashboard" />
  // Always go to /dashboard — public pages are visible without login
  return <Navigate replace to="/dashboard" />
}

export const router = createBrowserRouter([
  { path: '/', element: <RootRedirect />, errorElement: renderLazy(<AppErrorPage />) },
  {
    element: <PublicOnly role="user" redirectTo="/dashboard" />,
    errorElement: renderLazy(<AppErrorPage />),
    children: [
      {
        element: <AuthLayout />,
        children: [
          { path: '/login', element: renderLazy(<UserLoginPage />) },
          { path: '/register', element: renderLazy(<RegisterPage />) },
          { path: '/forgot-password', element: renderLazy(<ForgotPasswordPage />) },
        ],
      },
    ],
  },
  // Public user routes — visible without login
  {
    element: renderLazy(<UserLayout />),
    errorElement: renderLazy(<AppErrorPage />),
    children: [
      { path: '/dashboard', element: renderLazy(<UserDashboardPage />) },
      { path: '/models', element: renderLazy(<UserModelsPage />) },
      { path: '/docs', element: renderLazy(<UserDocsPage />) },
    ],
  },
  // Auth-required user routes
  {
    element: <RequireRole role="user" redirectTo="/login" />,
    errorElement: renderLazy(<AppErrorPage />),
    children: [
      {
        element: renderLazy(<UserLayout />),
        children: [
          { path: '/playground', element: renderLazy(<UserPlaygroundPage />) },
          { path: '/image-gen', element: renderLazy(<UserImageGenPage />) },
          { path: '/video-gen', element: renderLazy(<UserVideoGenPage />) },
          { path: '/keys', element: renderLazy(<UserKeysPage />) },
          { path: '/tasks', element: renderLazy(<UserTasksPage />) },
          { path: '/llm-logs', element: renderLazy(<UserLogsPage />) },
          { path: '/billing', element: renderLazy(<UserBillingPage />) },
          { path: '/recharge', element: <Navigate replace to="/billing" /> },
          { path: '/stats', element: renderLazy(<UserStatsPage />) },
          { path: '/exchange', element: renderLazy(<UserExchangePage />) },
          { path: '/invite', element: renderLazy(<UserInvitePage />) },
          { path: '/profile', element: renderLazy(<UserProfilePage />) },
        ],
      },
    ],
  },
  {
    element: <PublicOnly role="admin" redirectTo="/admin/dashboard" />,
    errorElement: renderLazy(<AppErrorPage />),
    children: [
      {
        element: renderLazy(<AuthLayout adminMode />),
        children: [{ path: '/admin/login', element: renderLazy(<AdminLoginPage />) }],
      },
    ],
  },
  {
    element: <RequireRole role="admin" redirectTo="/admin/login" />,
    errorElement: renderLazy(<AppErrorPage />),
    children: [
      {
        path: '/admin',
        element: renderLazy(<AdminLayout />),
        children: [
          { path: 'dashboard', element: renderLazy(<AdminDashboardPage />) },
          { path: 'channels', element: renderLazy(<AdminChannelsPage />) },
          { path: 'users', element: renderLazy(<AdminUsersPage />) },
          { path: 'billing', element: renderLazy(<AdminBillingPage />) },
          { path: 'cards', element: renderLazy(<AdminCardsPage />) },
          { path: 'key-pools', element: renderLazy(<AdminKeyPoolsPage />) },
          { path: 'ocpc', element: renderLazy(<AdminOcpcPage />) },
          { path: 'tasks', element: renderLazy(<AdminTasksPage />) },
          { path: 'llm-logs', element: renderLazy(<AdminLogsPage />) },
          { path: 'settings', element: renderLazy(<AdminSettingsPage />) },
          { path: 'vendors', element: renderLazy(<AdminVendorsPage />) },
          { path: 'withdraw', element: renderLazy(<AdminWithdrawPage />) },
        ],
      },
    ],
  },
  {
    element: <PublicOnly role="agent" redirectTo="/agent/dashboard" />,
    errorElement: renderLazy(<AppErrorPage />),
    children: [
      {
        element: renderLazy(<AuthLayout />),
        children: [{ path: '/agent/login', element: renderLazy(<AgentLoginPage />) }],
      },
    ],
  },
  {
    element: <RequireRole role="agent" redirectTo="/agent/login" />,
    errorElement: renderLazy(<AppErrorPage />),
    children: [
      {
        path: '/agent',
        element: renderLazy(<AgentLayout />),
        children: [{ path: 'dashboard', element: renderLazy(<AgentDashboardPage />) }],
      },
    ],
  },
  {
    element: <PublicOnly role="vendor" redirectTo="/vendor/dashboard" />,
    errorElement: renderLazy(<AppErrorPage />),
    children: [
      {
        element: renderLazy(<AuthLayout />),
        children: [
          { path: '/vendor/login', element: renderLazy(<VendorLoginPage />) },
          { path: '/vendor/register', element: renderLazy(<VendorRegisterPage />) },
        ],
      },
    ],
  },
  {
    element: <RequireRole role="vendor" redirectTo="/vendor/login" />,
    errorElement: renderLazy(<AppErrorPage />),
    children: [
      {
        path: '/vendor',
        element: renderLazy(<VendorLayout />),
        children: [
          { path: 'dashboard', element: renderLazy(<VendorDashboardPage />) },
          { path: 'keys', element: renderLazy(<VendorKeysPage />) },
        ],
      },
    ],
  },
  { path: '*', element: renderLazy(<NotFoundPage />) },
])
