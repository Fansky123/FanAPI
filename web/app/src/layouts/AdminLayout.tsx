import { ConsoleLayout, adminNavItems } from '@/layouts/ConsoleLayout'

export function AdminLayout() {
  return (
    <ConsoleLayout
      role="admin"
      items={adminNavItems}
    />
  )
}
