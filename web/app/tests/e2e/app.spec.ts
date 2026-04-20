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

test('renders extended user routes with authenticated session', async ({ page }) => {
  await page.addInitScript(() => {
    window.localStorage.setItem('token', 'mock-user-token')
  })

  await page.route('**/api/user/apikeys', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        api_keys: [{ id: 1, name: 'playground-key', key: 'sk-test-key' }],
      }),
    })
  })

  await page.route('**/api/user/channels', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        channels: [
          { id: 1, name: 'GPT-4o', routing_model: 'gpt-4o', type: 'llm' },
          { id: 2, name: 'Image Model', routing_model: 'nano-banana-pro', type: 'image' },
          { id: 3, name: 'Video Model', routing_model: 'video-pro', type: 'video' },
        ],
      }),
    })
  })

  await page.route('**/api/user/cards/redeem-history**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ records: [] }),
    })
  })

  await page.route('**/api/user/invite', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ invite_code: 'ABC123', invite_count: 2, frozen_balance: 500000 }),
    })
  })

  await page.route('**/api/user/payment-qr', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ wechat_qr: '', alipay_qr: '' }),
    })
  })

  await page.route('**/api/user/withdraw/history**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ records: [] }),
    })
  })

  await page.route('**/api/user/stats', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_consumed: 5200000,
        today_consumed: 1200000,
        daily_credits: [{ day: '04-20', credits: 1200000 }],
      }),
    })
  })

  for (const route of ['/playground', '/image-gen', '/video-gen', '/docs', '/stats', '/exchange', '/invite']) {
    await page.goto(route)
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
  }
})

test('renders extended admin routes with authenticated session', async ({ page }) => {
  await page.addInitScript(() => {
    window.localStorage.setItem('admin_token', 'mock-admin-token')
    window.localStorage.setItem('fanapi_ui_mode', 'admin')
  })

  await page.route('**/api/admin/key-pools**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ pools: [{ id: 1, name: 'Pool A', channel_id: 1, is_active: true, vendor_submittable: true }] }),
    })
  })

  await page.route('**/api/admin/ocpc/platforms**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ list: [{ id: 1, platform: 'baidu', name: 'Baidu Main', enabled: true, baidu_page_url: 'https://example.com' }] }),
    })
  })

  await page.route('**/api/admin/ocpc/schedule', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ schedule: { ocpc_schedule_enabled: 'true', ocpc_schedule_interval: '30' } }),
    })
  })

  await page.route('**/api/admin/cards**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ cards: [], total: 0 }),
    })
  })

  await page.route('**/api/admin/withdrawals**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ records: [], total: 0 }),
    })
  })

  await page.route('**/api/admin/withdrawals/pending-count', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ count: 0 }),
    })
  })

  for (const route of ['/admin/key-pools', '/admin/ocpc', '/admin/cards', '/admin/withdraw']) {
    await page.goto(route)
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
  }
})
