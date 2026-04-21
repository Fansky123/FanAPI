import { Badge } from '@/components/ui/badge'
import { ConsoleLayout } from '@/layouts/ConsoleLayout'
import { KeySquareIcon, LayoutDashboardIcon } from 'lucide-react'

export function VendorLayout() {
  return (
    <ConsoleLayout
      role="vendor"
      title="Vendor 端"
      subtitle="第二阶段迁移"
      items={[
        { label: '供应工作台', href: '/vendor/dashboard', icon: LayoutDashboardIcon },
        { label: '我的 API Key', href: '/vendor/keys', icon: KeySquareIcon },
      ]}
      identity={{ label: 'Vendor', description: '供应侧控制台' }}
      footer={<Badge variant="secondary">Phase 2</Badge>}
    />
  )
}
