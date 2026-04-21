import { ConsoleLayout } from '@/layouts/ConsoleLayout'
import { LayoutDashboardIcon, UsersIcon } from 'lucide-react'

export function AgentLayout() {
  return (
    <ConsoleLayout
      role="agent"
      items={[
        { label: '工作台', href: '/agent/dashboard', icon: LayoutDashboardIcon },
        { label: '用户管理', href: '/agent/users', icon: UsersIcon },
      ]}
    />
  )
}
