import { ConsoleLayout, userNavGroups } from '@/layouts/ConsoleLayout'

export function UserLayout() {
  return (
    <ConsoleLayout
      role="user"
      groups={userNavGroups}
    />
  )
}
