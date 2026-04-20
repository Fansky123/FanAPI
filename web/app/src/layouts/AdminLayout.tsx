import { Badge } from '@/components/ui/badge'
import { ConsoleLayout, adminNavItems } from '@/layouts/ConsoleLayout'

export function AdminLayout() {
  return (
    <ConsoleLayout
      role="admin"
      title="管理后台"
      subtitle="运营、监控与配置中心"
      items={adminNavItems}
      identity={{ label: '管理员', description: '平台运维与配置控制' }}
      footer={<Badge>Admin</Badge>}
    />
  )
}
