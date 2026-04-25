import type { ComponentType, ReactNode } from 'react'
import { useEffect, useState } from 'react'
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom'
import {
  BlocksIcon,
  BookOpenIcon,
  BriefcaseBusinessIcon,
  CreditCardIcon,
  FileClockIcon,
  HeadphonesIcon,
  ImageIcon,
  KeySquareIcon,
  LayoutDashboardIcon,
  ListIcon,
  LogOutIcon,
  MessageCircleIcon,
  MessageSquareIcon,
  MegaphoneIcon,
  NetworkIcon,
  SettingsIcon,
  ShareIcon,
  ShoppingCartIcon,
  TicketIcon,
  TrendingUpIcon,
  UserRoundIcon,
  UsersIcon,
  UsersRoundIcon,
  VideoIcon,
  WalletCardsIcon,
  WalletIcon,
} from 'lucide-react'

import { AppLogo } from '@/components/shared/AppLogo'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
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
import { clearRoleToken, getRoleToken } from '@/lib/auth/storage'
import { userApi } from '@/lib/api/user'

type NavItem = {
  label: string
  href: string
  icon: ComponentType<{ className?: string }>
}

type NavGroup = {
  label?: string
  items: NavItem[]
  requiresAuth?: boolean
}

type ConsoleLayoutProps = {
  role: 'user' | 'admin' | 'agent' | 'vendor'
  // Support flat list (legacy) or grouped navigation
  items?: NavItem[]
  groups?: NavGroup[]
  // For legacy callers (admin/agent/vendor layouts)
  title?: string
  subtitle?: string
  identity?: {
    label: string
    description: string
  }
  footer?: ReactNode
}

function HeaderBalanceChip() {
  const navigate = useNavigate()
  const [balance, setBalance] = useState<number | null>(null)
  useEffect(() => {
    let cancelled = false
    userApi.getBalance()
      .then((res) => { if (!cancelled) setBalance(res.balance_credits ?? 0) })
      .catch(() => {})
    return () => { cancelled = true }
  }, [])
  if (balance == null) return null
  return (
    <Button
      variant="outline"
      size="sm"
      className="hidden gap-1.5 rounded-full pl-2 pr-3 sm:inline-flex"
      onClick={() => navigate('/billing')}
      title="点击充值"
    >
      <WalletIcon className="size-4 text-muted-foreground" />
      <span className="font-mono text-xs tabular-nums">¥{(balance / 1e6).toFixed(4)}</span>
    </Button>
  )
}

function ContactPopover({
  imageUrl,
  label,
  description,
  Icon,
}: {
  imageUrl: string
  label: string
  description: string
  Icon: ComponentType<{ className?: string }>
}) {
  if (!imageUrl) return null
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="ghost" size="sm" className="hidden gap-1.5 px-2 text-muted-foreground hover:text-foreground md:inline-flex">
          <Icon className="size-4" />
          <span className="text-xs">{label}</span>
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-[260px]">
        <DialogHeader>
          <DialogTitle>{label}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <div className="flex justify-center pb-2">
          <img src={imageUrl} alt={label} className="h-48 w-48 rounded-lg border object-contain p-1" />
        </div>
      </DialogContent>
    </Dialog>
  )
}

export function ConsoleLayout({
  role,
  items = [],
  groups,
  title,
  subtitle,
  identity,
  footer,
}: ConsoleLayoutProps) {
  const location = useLocation()
  const navigate = useNavigate()
  const { settings } = useSiteSettings()
  const { siteName, logoUrl, qqGroupUrl, wechatCsUrl, headerHtml, footerHtml } = settings

  const isLoggedIn = !!getRoleToken(role)
  const displayName = identity?.label
  const isFullBleedPage = role === 'user' && location.pathname === '/docs'

  // Build nav groups from either `groups` or flat `items`, filter auth-gated when not logged in
  const rawGroups: NavGroup[] = groups ?? (subtitle ? [{ label: subtitle, items }] : [{ items }])
  const navGroups = rawGroups.filter((g) => !g.requiresAuth || isLoggedIn)

  // Find current page title from active nav item
  const allItems = navGroups.flatMap((g) => g.items)
  const currentItem = allItems.find((item) => location.pathname === item.href)
  const pageTitle = currentItem?.label ?? title ?? siteName

  function logout() {
    clearRoleToken(role)
    navigate(
      role === 'admin' ? '/admin/login' :
      role === 'agent' ? '/agent/login' :
      role === 'vendor' ? '/vendor/login' : '/login'
    )
  }

  return (
    <SidebarProvider>
      <Sidebar collapsible="offcanvas">
        <SidebarHeader>
          <AppLogo siteName={siteName} logoUrl={logoUrl} label={siteName} />
        </SidebarHeader>
        <SidebarSeparator />
        <SidebarContent>
          {navGroups.map((group, i) => (
            <SidebarGroup key={i}>
              {group.label && <SidebarGroupLabel>{group.label}</SidebarGroupLabel>}
              <SidebarGroupContent>
                <SidebarMenu>
                  {group.items.map((item) => {
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
          ))}
        </SidebarContent>
        <SidebarFooter>
          {footer}
          <div className="flex flex-col gap-1 px-1 pb-2">
            {isLoggedIn ? (
              <>
                {displayName && (
                  <div className="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm text-muted-foreground">
                    <Avatar className="size-6">
                      <AvatarFallback className="text-xs">
                        {displayName.slice(0, 1).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <span className="truncate">{displayName}</span>
                  </div>
                )}
                <Button
                  variant="ghost"
                  size="sm"
                  className="justify-start gap-2 text-muted-foreground hover:text-foreground"
                  onClick={logout}
                >
                  <LogOutIcon className="size-4" />
                  退出登录
                </Button>
              </>
            ) : role === 'user' ? (
              <div className="flex flex-col gap-1 px-1">
                <Button asChild size="sm" className="w-full">
                  <Link to="/login">登录</Link>
                </Button>
                <Button asChild size="sm" variant="outline" className="w-full">
                  <Link to="/register">注册</Link>
                </Button>
              </div>
            ) : null}
          </div>
        </SidebarFooter>
      </Sidebar>
      <SidebarInset>
        <header className="sticky top-0 z-20 flex h-[54px] items-center justify-between border-b border-border/60 bg-background px-4">
          <div className="flex items-center gap-3">
            <SidebarTrigger />
            <span className="text-sm font-semibold">{pageTitle}</span>
          </div>
          <div className="flex items-center gap-1.5 sm:gap-2">
            <ContactPopover
              imageUrl={qqGroupUrl}
              label="QQ 交流群"
              description="扫码加入 QQ 交流群"
              Icon={UsersRoundIcon}
            />
            <ContactPopover
              imageUrl={wechatCsUrl}
              label="微信客服"
              description="扫码添加微信客服"
              Icon={MessageCircleIcon}
            />
            {isLoggedIn && role === 'user' && <HeaderBalanceChip />}
            {isLoggedIn && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" size="sm" className="gap-2 rounded-full pl-2">
                    <Avatar className="size-6">
                      <AvatarFallback className="text-xs">
                        {displayName
                          ? displayName.slice(0, 1).toUpperCase()
                          : role.slice(0, 1).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <span className="hidden max-w-28 truncate text-sm sm:inline">
                      {displayName ??
                        (role === 'admin' ? '管理员' :
                         role === 'agent' ? 'Agent' :
                         role === 'vendor' ? 'Vendor' : '用户')}
                    </span>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-36">
                  <DropdownMenuItem onClick={logout}>
                    <LogOutIcon data-icon="inline-start" />
                    退出登录
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            )}
          </div>
        </header>
        {headerHtml ? (
          <div
            className="border-b border-border/60 bg-card px-4 py-2 text-sm md:px-6"
            dangerouslySetInnerHTML={{ __html: headerHtml }}
          />
        ) : null}
        <main className={isFullBleedPage ? 'flex-1 px-0 py-0' : 'flex-1 px-4 py-6 md:px-6'}>
          <div className={isFullBleedPage ? 'flex w-full flex-col' : 'mx-auto flex w-full max-w-7xl flex-col gap-6'}>
            <Outlet />
          </div>
        </main>
        {footerHtml ? (
          <footer
            className="border-t border-border/60 bg-background px-4 py-3 text-xs text-muted-foreground md:px-6"
            dangerouslySetInnerHTML={{ __html: footerHtml }}
          />
        ) : null}
      </SidebarInset>
    </SidebarProvider>
  )
}

export const userNavGroups: NavGroup[] = [
  {
    items: [
      { label: '数据看板', href: '/dashboard', icon: LayoutDashboardIcon },
      { label: '模型列表', href: '/models', icon: BlocksIcon },
      { label: '调用日志', href: '/llm-logs', icon: FileClockIcon },
      { label: '任务中心', href: '/tasks', icon: ListIcon },
      { label: '使用统计', href: '/stats', icon: TrendingUpIcon },
      { label: '接口文档', href: '/docs', icon: BookOpenIcon },
    ],
  },
  {
    label: '在线体验',
    requiresAuth: true,
    items: [
      { label: '文本对话', href: '/playground', icon: MessageSquareIcon },
      { label: '图片生成', href: '/image-gen', icon: ImageIcon },
      { label: '视频生成', href: '/video-gen', icon: VideoIcon },
      { label: '音乐生成', href: '/music-gen', icon: HeadphonesIcon },
    ],
  },
  {
    label: '账户管理',
    requiresAuth: true,
    items: [
      { label: 'API 密钥', href: '/keys', icon: KeySquareIcon },
      { label: '积分充值', href: '/billing', icon: ShoppingCartIcon },
      { label: '兑换中心', href: '/exchange', icon: TicketIcon },
      { label: '个人中心', href: '/profile', icon: UserRoundIcon },
      { label: '邀请中心', href: '/invite', icon: ShareIcon },
    ],
  },
]

/** @deprecated Use userNavGroups instead */
export const userNavItems: NavItem[] = userNavGroups.flatMap((g) => g.items)

export const adminNavItems: NavItem[] = [
  { label: '数据概览', href: '/admin/dashboard', icon: LayoutDashboardIcon },
  { label: '渠道管理', href: '/admin/channels', icon: NetworkIcon },
  { label: '号池管理', href: '/admin/key-pools', icon: KeySquareIcon },
  { label: '用户管理', href: '/admin/users', icon: UsersIcon },
  { label: '账单流水', href: '/admin/billing', icon: WalletCardsIcon },
  { label: '任务中心', href: '/admin/tasks', icon: ListIcon },
  { label: '调用日志', href: '/admin/llm-logs', icon: FileClockIcon },
  { label: '卡密管理', href: '/admin/cards', icon: CreditCardIcon },
  { label: 'OCPC 上报', href: '/admin/ocpc', icon: MegaphoneIcon },
  { label: '号商管理', href: '/admin/vendors', icon: BriefcaseBusinessIcon },
  { label: '提现管理', href: '/admin/withdraw', icon: WalletIcon },
  { label: '系统设置', href: '/admin/settings', icon: SettingsIcon },
]
