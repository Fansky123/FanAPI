import { isRouteErrorResponse, Link, useRouteError } from 'react-router-dom'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function AppErrorPage() {
  const error = useRouteError()

  let title = '页面暂时不可用'
  let description = '系统遇到了意外错误，请稍后重试。'

  if (isRouteErrorResponse(error)) {
    title = `${error.status} ${error.statusText || '请求失败'}`
    description =
      typeof error.data === 'string'
        ? error.data
        : '当前路由在加载时发生错误。'
  } else if (error instanceof Error) {
    description = error.message
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-6">
      <Card className="w-full max-w-xl border-border/70 shadow-lg">
        <CardHeader className="flex flex-col gap-3">
          <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
            Application error
          </p>
          <CardTitle className="text-3xl tracking-tight">{title}</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <p className="text-sm leading-7 text-muted-foreground">{description}</p>
          <div className="flex gap-3">
            <Button asChild>
              <Link to="/">返回首页</Link>
            </Button>
            <Button variant="outline" onClick={() => window.location.reload()}>
              刷新重试
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
