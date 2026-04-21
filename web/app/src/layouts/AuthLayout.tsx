import { Outlet } from 'react-router-dom'

import { AppLogo } from '@/components/shared/AppLogo'
import { useSiteSettings } from '@/hooks/use-site-settings'

export function AuthLayout({ adminMode = false }: { adminMode?: boolean }) {
  const { siteName, logoUrl } = useSiteSettings()

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,rgba(15,118,110,0.08),transparent_34%),linear-gradient(180deg,var(--background),color-mix(in_oklab,var(--background)_92%,var(--muted)_8%))] px-6 py-10">
      <div className="mx-auto flex min-h-[calc(100vh-5rem)] max-w-6xl items-center justify-center">
        <div className="grid w-full gap-12 lg:grid-cols-[1.1fr_0.9fr]">
          <div className="hidden flex-col justify-between rounded-3xl border border-border/60 bg-card/85 p-10 shadow-sm backdrop-blur lg:flex">
            <div className="space-y-10">
              <AppLogo
                siteName={siteName}
                logoUrl={logoUrl}
                label={adminMode ? '管理后台重构版' : '新一代控制台'}
              />
              <div className="space-y-4">
                <p className="text-sm font-medium uppercase tracking-[0.18em] text-muted-foreground">
                  Frontend rebuild
                </p>
                <h1 className="max-w-lg text-4xl font-semibold tracking-tight text-foreground">
                  更清晰、更统一、更适合长期协作的 FanAPI 前端。
                </h1>
                <p className="max-w-xl text-base leading-7 text-muted-foreground">
                  新版本按统一设计系统重构，优先保证清晰度、信息密度和可维护性，
                  让后续继续开发时不再回到“页面各写各的”的状态。
                </p>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4 text-sm text-muted-foreground">
              <div className="rounded-2xl border border-border/60 bg-background/70 p-4">
                一个前端应用承载四端
              </div>
              <div className="rounded-2xl border border-border/60 bg-background/70 p-4">
                React + Tailwind + shadcn/ui
              </div>
              <div className="rounded-2xl border border-border/60 bg-background/70 p-4">
                DESIGN.md 作为最高规范
              </div>
              <div className="rounded-2xl border border-border/60 bg-background/70 p-4">
                面向最终上线版本
              </div>
            </div>
          </div>
          <div className="flex items-center justify-center">
            <Outlet />
          </div>
        </div>
      </div>
    </div>
  )
}
