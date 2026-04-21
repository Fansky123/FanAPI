import { useEffect, useState } from 'react'
import { BlocksIcon } from 'lucide-react'

import { EmptyState } from '@/components/shared/EmptyState'
import { PageHeader } from '@/components/shared/PageHeader'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { getApiErrorMessage } from '@/lib/api/http'
import { userApi, type UserChannel } from '@/lib/api/user'

export function UserModelsPage() {
  const [channels, setChannels] = useState<UserChannel[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const response = await userApi.listChannels()
        setChannels(Array.isArray(response) ? response : response.channels ?? [])
      } catch (err) {
        setError(getApiErrorMessage(err))
      }
    }

    void load()
  }, [])

  return (
    <>
      <PageHeader
        eyebrow="Catalog"
        title="模型列表"
        description="按统一卡片密度和文本层级展示可用模型，替代旧版视觉不一致的模型页。"
      />
      {error ? (
        <Card className="border-destructive/25 bg-destructive/5">
          <CardContent className="py-4 text-sm text-destructive">{error}</CardContent>
        </Card>
      ) : null}
      {channels.length === 0 ? (
        <EmptyState
          icon={<BlocksIcon className="size-6 text-muted-foreground" />}
          title="暂无模型数据"
          description="当前接口没有返回可用模型，后续这里会承接搜索、分类和价格说明。"
        />
      ) : (
        <div className="grid gap-4 lg:grid-cols-2 2xl:grid-cols-3">
          {channels.map((channel, index) => (
            <Card key={channel.id ?? channel.routing_model ?? index}>
              <CardContent className="space-y-4 p-6">
                <div className="flex items-start justify-between gap-3">
                  <div className="space-y-1">
                    <h2 className="text-base font-semibold">
                      {channel.name ?? channel.routing_model ?? '未命名模型'}
                    </h2>
                    <p className="text-sm text-muted-foreground">
                      {channel.description ?? '模型描述待补充'}
                    </p>
                  </div>
                  <Badge variant="secondary">
                    {channel.type ?? channel.category ?? 'llm'}
                  </Badge>
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
