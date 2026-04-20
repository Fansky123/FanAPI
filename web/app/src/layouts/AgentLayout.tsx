import { Badge } from '@/components/ui/badge'
import { ConsoleLayout } from '@/layouts/ConsoleLayout'
import { LayoutDashboardIcon } from 'lucide-react'

export function AgentLayout() {
  return (
    <ConsoleLayout
      role="agent"
      title="Agent 端"
      subtitle="第二阶段迁移"
      items={[{ label: '工作台', href: '/agent/dashboard', icon: LayoutDashboardIcon }]}
      identity={{ label: 'Agent', description: '业务协作视图' }}
      footer={<Badge variant="secondary">Phase 2</Badge>}
    />
  )
}
