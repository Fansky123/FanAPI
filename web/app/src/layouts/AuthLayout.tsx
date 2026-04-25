import { Outlet } from 'react-router-dom'
import { ActivityIcon, GaugeIcon, ShieldCheckIcon, SparklesIcon } from 'lucide-react'

import { AppLogo } from '@/components/shared/AppLogo'
import { ThemeToggle } from '@/components/shared/ThemeToggle'
import { Badge } from '@/components/ui/badge'
import { useSiteSettings } from '@/hooks/use-site-settings'

export function AuthLayout({ adminMode = false }: { adminMode?: boolean }) {
  const { settings: { siteName, logoUrl } } = useSiteSettings()
  const label = adminMode ? '管理后台' : '用户控制台'
  const headline = adminMode ? '把运营、账务和风控放进同一个清晰入口。' : '更快进入模型调用、密钥和账单管理。'
  const description = adminMode
    ? '面向日常管理场景优化信息密度，减少跳转，让核心指标和待处理事项进入后台后就能被快速扫描。'
    : '登录后可以继续管理 API Key、查看调用统计、处理账单和生成任务，界面保持一致、稳定和易读。'
  const highlights = adminMode
    ? [
        { label: '运营监控', value: '实时' },
        { label: '账务核对', value: '统一' },
        { label: '权限入口', value: '独立' },
      ]
    : [
        { label: 'API Key', value: '集中' },
        { label: '调用统计', value: '清晰' },
        { label: '账单余额', value: '可查' },
      ]

  return (
    <div className="relative min-h-screen overflow-hidden bg-[radial-gradient(circle_at_18%_18%,color-mix(in_oklab,var(--primary)_12%,transparent),transparent_28%),radial-gradient(circle_at_82%_12%,color-mix(in_oklab,var(--chart-3)_13%,transparent),transparent_26%),linear-gradient(135deg,color-mix(in_oklab,var(--background)_96%,var(--muted)_4%),var(--background)_48%,color-mix(in_oklab,var(--background)_92%,var(--accent)_8%))] px-5 py-5 sm:px-8 lg:px-10">
      <div className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-border to-transparent" />
      <div className="mx-auto flex min-h-[calc(100vh-2.5rem)] w-full max-w-7xl flex-col">
        <div className="flex items-center justify-between py-3 lg:hidden">
          <AppLogo siteName={siteName} logoUrl={logoUrl} label={label} />
          <ThemeToggle />
        </div>
        <div className="grid flex-1 items-center gap-10 py-6 lg:grid-cols-[1.05fr_0.95fr] lg:gap-14 lg:py-10">
          <section className="hidden min-h-[640px] flex-col justify-between lg:flex">
            <div className="flex items-center justify-between">
              <AppLogo
                siteName={siteName}
                logoUrl={logoUrl}
                label={label}
              />
              <ThemeToggle />
            </div>

            <div className="flex max-w-2xl flex-col gap-8">
              <div className="flex flex-col gap-4">
                <Badge variant="secondary" className="w-fit">
                  <SparklesIcon />
                  FanAPI Console
                </Badge>
                <h1 className="max-w-2xl text-5xl font-semibold leading-tight tracking-tight text-foreground">
                  {headline}
                </h1>
                <p className="max-w-xl text-base leading-7 text-muted-foreground">{description}</p>
              </div>

              <div className="grid max-w-xl grid-cols-3 gap-3">
                {highlights.map((item) => (
                  <div
                    key={item.label}
                    className="rounded-xl border border-border/70 bg-card/70 p-4 shadow-sm backdrop-blur"
                  >
                    <p className="text-2xl font-semibold tracking-tight text-foreground">
                      {item.value}
                    </p>
                    <p className="mt-1 text-xs text-muted-foreground">{item.label}</p>
                  </div>
                ))}
              </div>
            </div>

            <div className="rounded-2xl border border-border/70 bg-card/72 p-4 shadow-sm backdrop-blur">
              <div className="flex items-center justify-between border-b border-border/70 pb-4">
                <div>
                  <p className="text-sm font-medium text-foreground">
                    {adminMode ? '今日后台概览' : '账户使用概览'}
                  </p>
                  <p className="text-xs text-muted-foreground">最近 24 小时</p>
                </div>
                <Badge variant="outline">
                  <ActivityIcon />
                  Online
                </Badge>
              </div>
              <div className="grid gap-4 pt-4">
                <div className="grid grid-cols-[1fr_0.78fr] gap-4">
                  <div className="rounded-xl bg-muted/55 p-4">
                    <div className="mb-5 flex items-center justify-between">
                      <div className="flex items-center gap-2 text-sm font-medium">
                        <GaugeIcon className="text-muted-foreground" />
                        {adminMode ? '请求与结算' : '请求趋势'}
                      </div>
                      <span className="text-xs text-muted-foreground">Stable</span>
                    </div>
                    <div className="flex h-28 items-end gap-2">
                      {[42, 64, 52, 78, 58, 88, 72, 96].map((height, index) => (
                        <div
                          key={`${height}-${index}`}
                          className="flex-1 rounded-t-md bg-primary/80"
                          style={{ height: `${height}%` }}
                        />
                      ))}
                    </div>
                  </div>
                  <div className="flex flex-col justify-between rounded-xl bg-muted/55 p-4">
                    <ShieldCheckIcon className="text-muted-foreground" />
                    <div>
                      <p className="text-3xl font-semibold tracking-tight">
                        {adminMode ? '99.9%' : '24/7'}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {adminMode ? '服务可用性' : '接口可用'}
                      </p>
                    </div>
                  </div>
                </div>
                <div className="grid grid-cols-3 gap-3">
                  <div className="h-2 rounded-full bg-chart-2/70" />
                  <div className="h-2 rounded-full bg-chart-3/70" />
                  <div className="h-2 rounded-full bg-chart-5/70" />
                </div>
              </div>
            </div>
          </section>

          <div className="flex items-center justify-center lg:justify-end">
            <Outlet />
          </div>
        </div>
      </div>
    </div>
  )
}
