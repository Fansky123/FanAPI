import type { ComponentType, ReactNode } from 'react'
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom'
import {
  BlocksIcon,
  BriefcaseBusinessIcon,
  DatabaseZapIcon,
  FileTextIcon,
  KeySquareIcon,
  LayoutDashboardIcon,
  LogOutIcon,
  NetworkIcon,
  SettingsIcon,
  UserRoundIcon,
  UsersIcon,
  WalletCardsIcon,
} from 'lucide-react'

import { AppLogo } from '@/components/shared/AppLogo'
import { ThemeToggle } from '@/components/shared/ThemeToggle'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarSeparator,
  SidebarTrigger,
} from '@/components/ui/sidebar'
import { useSiteSettings } from '@/hooks/use-site-settings'
import { clearRoleToken, setSiteModePreference } from '@/lib/auth/storage'

type NavItem = {
  label: string
  href: string
  icon: ComponentType<{ className?: string }>
}

type ConsoleLayoutProps = {
  role: 'user' | 'admin' | 'agent' | 'vendor'
  title: string
  subtitle: string
  items: NavItem[]
  identity: {
    label: string
    description: string
  }
  footer?: ReactNode
}

export function ConsoleLayout({
  role,
  title,
  subtitle,
  items,
  identity,
  footer,
}: ConsoleLayoutProps) {
  const location = useLocation()
  const navigate = useNavigate()
  const { settings: { siteName, logoUrl } } = useSiteSettings()

  return (
    <SidebarProvider>
      <Sidebar variant="inset" collapsible="icon">
        <SidebarHeader>
          <AppLogo siteName={siteName} logoUrl={logoUrl} label={title} />
        </SidebarHeader>
        <SidebarSeparator />
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupLabel>{subtitle}</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {items.map((item) => {
                  const active = location.pathname === item.href
                  return (
                    <SidebarMenuItem key={item.href}>
                      <SidebarMenuButton asChild isActive={active} tooltip={item.label}>
                        <Link to={item.href}>
                          <item.icon />
                          <span>{item.label}</span>
                        </Link>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  )
                })}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter>
          {footer}
          <div className="rounded-xl border border-border/70 bg-background/70 p-3">
            <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
              当前身份
            </p>
            <div className="mt-3 flex items-center gap-3">
              <Avatar className="size-10">
                <AvatarFallback>{identity.label.slice(0, 1).toUpperCase()}</AvatarFallback>
              </Avatar>
              <div className="min-w-0">
                <p className="truncate text-sm font-medium">{identity.label}</p>
                <p className="truncate text-xs text-muted-foreground">
                  {identity.description}
                </p>
              </div>
            </div>
          </div>
        </SidebarFooter>
      </Sidebar>
      <SidebarInset>
        <header className="sticky top-0 z-20 flex h-16 items-center justify-between border-b border-border/70 bg-background/90 px-4 backdrop-blur md:px-6">
          <div className="flex items-center gap-3">
            <SidebarTrigger />
            <div className="hidden md:block">
              <p className="text-sm font-medium">{title}</p>
              <p className="text-xs text-muted-foreground">{subtitle}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <ThemeToggle />
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" className="gap-2 rounded-full pl-2">
                  <Avatar className="size-7">
                    <AvatarFallback>{identity.label.slice(0, 1).toUpperCase()}</AvatarFallback>
                  </Avatar>
                  <span className="max-w-32 truncate">{identity.label}</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-48">
                <DropdownMenuItem
                  onClick={() => {
                    setSiteModePreference(role)
                    navigate(role === 'user' ? '/profile' : `/${role}/dashboard`)
                  }}
                >
                  <UserRoundIcon data-icon="inline-start" />
                  身份主页
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => {
                    clearRoleToken(role)
                    navigate(role === 'admin' ? '/admin/login' : '/login')
                  }}
                >
                  <LogOutIcon data-icon="inline-start" />
                  退出登录
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>
        <main className="flex-1 bg-[linear-gradient(180deg,color-mix(in_oklab,var(--background)_96%,var(--muted)_4%),var(--background))] px-4 py-6 md:px-6">
          <div className="mx-auto flex w-full max-w-7xl flex-col gap-6">
            <Outlet />
          </div>
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}

export const userNavItems: NavItem[] = [
  { label: '数据看板', href: '/dashboard', icon: LayoutDashboardIcon },
  { label: '模型列表', href: '/models', icon: BlocksIcon },
  { label: '文本对话', href: '/playground', icon: FileTextIcon },
  { label: '图片生成', href: '/image-gen', icon: BlocksIcon },
  { label: '视频生成', href: '/video-gen', icon: BlocksIcon },
  { label: 'API 密钥', href: '/keys', icon: KeySquareIcon },
  { label: '任务中心', href: '/tasks', icon: DatabaseZapIcon },
  { label: '调用日志', href: '/llm-logs', icon: FileTextIcon },
  { label: '接口文档', href: '/docs', icon: FileTextIcon },
  { label: '使用统计', href: '/stats', icon: LayoutDashboardIcon },
  { label: '兑换中心', href: '/exchange', icon: WalletCardsIcon },
  { label: '邀请中心', href: '/invite', icon: UsersIcon },
  { label: '账单订单', href: '/billing', icon: WalletCardsIcon },
  { label: '个人中心', href: '/profile', icon: UserRoundIcon },
]

export const adminNavItems: NavItem[] = [
  { label: '平台看板', href: '/admin/dashboard', icon: LayoutDashboardIcon },
  { label: '渠道管理', href: '/admin/channels', icon: NetworkIcon },
  { label: '用户管理', href: '/admin/users', icon: UsersIcon },
  { label: '账单流水', href: '/admin/billing', icon: WalletCardsIcon },
  { label: '卡密管理', href: '/admin/cards', icon: KeySquareIcon },
  { label: '号池管理', href: '/admin/key-pools', icon: KeySquareIcon },
  { label: 'OCPC 管理', href: '/admin/ocpc', icon: NetworkIcon },
  { label: '任务中心', href: '/admin/tasks', icon: DatabaseZapIcon },
  { label: '调用日志', href: '/admin/llm-logs', icon: FileTextIcon },
  { label: '系统设置', href: '/admin/settings', icon: SettingsIcon },
  { label: '号商管理', href: '/admin/vendors', icon: BriefcaseBusinessIcon },
  { label: '提现审核', href: '/admin/withdraw', icon: WalletCardsIcon },
]
