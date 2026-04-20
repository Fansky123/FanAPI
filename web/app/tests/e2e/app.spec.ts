import { expect, test } from '@playwright/test'

test.beforeEach(async ({ page }) => {
  await page.route('**/api/public/settings', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        settings: {
          site_name: 'FanAPI',
          logo_url: '',
        },
      }),
    })
  })
})

test('renders user login page', async ({ page }) => {
  await page.goto('/login')

  await expect(page.getByText('登录用户端')).toBeVisible()
  await expect(page.getByPlaceholder('请输入用户名或邮箱')).toBeVisible()
})

test('renders admin login page', async ({ page }) => {
  await page.goto('/admin/login')

  await expect(page.getByText('登录管理后台')).toBeVisible()
  await expect(page.getByRole('button', { name: '进入后台' })).toBeVisible()
})

test('redirects protected user route to login when unauthenticated', async ({ page }) => {
  await page.goto('/dashboard')

  await expect(page).toHaveURL(/\/login$/)
})

test('renders user dashboard with authenticated mocks', async ({ page }) => {
  await page.addInitScript(() => {
    window.localStorage.setItem('token', 'mock-user-token')
  })

  await page.route('**/api/user/balance', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ balance_credits: 1800000 }),
    })
  })

  await page.route('**/api/user/stats', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_consumed: 5200000,
        today_consumed: 1200000,
      }),
    })
  })

  await page.goto('/dashboard')

  await expect(page.getByText('用户数据看板')).toBeVisible()
  await expect(page.getByText('1.80')).toBeVisible()
  await expect(page.getByText('5.20')).toBeVisible()
  await expect(page.getByText('1.20')).toBeVisible()
})

test('renders admin dashboard with authenticated mocks', async ({ page }) => {
  await page.addInitScript(() => {
    window.localStorage.setItem('admin_token', 'mock-admin-token')
    window.localStorage.setItem('fanapi_ui_mode', 'admin')
  })

  await page.route('**/api/admin/stats', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_users: 42,
        total_requests: 380,
        total_revenue: 9900000,
      }),
    })
  })

  await page.goto('/admin/dashboard')

  await expect(page.getByText('平台运营看板')).toBeVisible()
  await expect(page.getByText('42')).toBeVisible()
  await expect(page.getByText('380')).toBeVisible()
  await expect(page.getByText('9900000')).toBeVisible()
})
