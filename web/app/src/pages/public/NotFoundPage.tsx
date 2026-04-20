import { Link } from 'react-router-dom'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function NotFoundPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-6">
      <Card className="w-full max-w-xl border-border/70 shadow-lg">
        <CardHeader className="space-y-3">
          <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
            404
          </p>
          <CardTitle className="text-3xl tracking-tight">页面不存在</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm leading-7 text-muted-foreground">
            当前地址没有对应页面。你可以返回首页，或回到所属角色的控制台。
          </p>
          <Button asChild>
            <Link to="/">返回首页</Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
