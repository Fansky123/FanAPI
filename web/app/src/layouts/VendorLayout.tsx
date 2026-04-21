import { ConsoleLayout } from '@/layouts/ConsoleLayout'
import { KeySquareIcon, LayoutDashboardIcon } from 'lucide-react'

export function VendorLayout() {
  return (
    <ConsoleLayout
      role="vendor"
      items={[
        { label: '供应工作台', href: '/vendor/dashboard', icon: LayoutDashboardIcon },
        { label: '我的 API Key', href: '/vendor/keys', icon: KeySquareIcon },
      ]}
    />
  )
}
