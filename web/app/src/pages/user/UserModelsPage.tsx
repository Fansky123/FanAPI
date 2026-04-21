import { BlocksIcon } from 'lucide-react'

import { EmptyState } from '@/components/shared/EmptyState'
import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { userApi, type UserChannel } from '@/lib/api/user'
import { useAsync } from '@/hooks/use-async'

export function UserModelsPage() {
  const { data: channels, loading, error, reload } = useAsync(async () => {
    const response = await userApi.listChannels()
    return Array.isArray(response) ? response : response.channels ?? []
  }, [] as UserChannel[])

  return (
    <>
      <PageHeader
        eyebrow="Catalog"
        title="模型列表"
        description="查看当前可用的 AI 模型渠道，包含路由键和类型信息。"
        actions={
          error ? (
            <Button size="sm" variant="outline" onClick={reload}>
              重试
            </Button>
          ) : null
        }
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      {loading ? (
        <div className="grid gap-4 lg:grid-cols-2 2xl:grid-cols-3">
          {Array.from({ length: 6 }).map((_, index) => (
            <Card key={index}>
              <CardContent className="flex flex-col gap-4 p-6">
                <div className="flex items-start justify-between gap-3">
                  <div className="flex flex-col gap-2">
                    <Skeleton className="h-5 w-32" />
                    <Skeleton className="h-4 w-48" />
                  </div>
                  <Skeleton className="h-5 w-12" />
                </div>
                <Skeleton className="h-10 w-full" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : channels.length === 0 ? (
        <EmptyState
          icon={<BlocksIcon className="size-6 text-muted-foreground" />}
          title="暂无模型数据"
          description="当前接口没有返回可用模型，请联系管理员添加渠道。"
        />
      ) : (
        <div className="grid gap-4 lg:grid-cols-2 2xl:grid-cols-3">
          {channels.map((channel, index) => (
            <Card key={channel.id ?? channel.routing_model ?? index}>
              <CardContent className="flex flex-col gap-4 p-6">
                <div className="flex items-start justify-between gap-3">
                  <div className="flex flex-col gap-1">
                    <h2 className="text-base font-semibold">
                      {channel.name ?? channel.routing_model ?? '未命名模型'}
                    </h2>
                    <p className="text-sm text-muted-foreground">
                      {channel.description ?? '暂无模型描述'}
                    </p>
                  </div>
                  <Badge variant="secondary">{channel.type ?? channel.category ?? 'llm'}</Badge>
                </div>
                <div className="rounded-xl border border-border/70 bg-muted/25 px-4 py-3 text-sm text-muted-foreground">
                  路由键：{channel.routing_model ?? channel.model ?? '未知'}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </>
  )
}
