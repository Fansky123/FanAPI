import { Badge } from '@/components/ui/badge'
import { ConsoleLayout, userNavItems } from '@/layouts/ConsoleLayout'

export function UserLayout() {
  return (
    <ConsoleLayout
      role="user"
      title="用户控制台"
      subtitle="模型、密钥、账单与用量"
      items={userNavItems}
      identity={{ label: '用户', description: '产品使用与账号管理' }}
      footer={<Badge variant="secondary">Phase 1</Badge>}
    />
  )
}
