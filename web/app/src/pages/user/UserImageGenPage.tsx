import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import { userApi, type ApiKeyRecord, type UserChannel } from '@/lib/api/user'

export function UserImageGenPage() {
  const [apiKeys, setApiKeys] = useState<ApiKeyRecord[]>([])
  const [channels, setChannels] = useState<UserChannel[]>([])
  const [selectedKeyId, setSelectedKeyId] = useState<number | undefined>()
  const [selectedModel, setSelectedModel] = useState('')
  const [prompt, setPrompt] = useState('')
  const [size, setSize] = useState('1k')
  const [aspectRatio, setAspectRatio] = useState('1:1')
  const [taskId, setTaskId] = useState('')
  const [images, setImages] = useState<string[]>([])
  const [running, setRunning] = useState(false)

  useEffect(() => {
    async function load() {
      try {
        const [keysRes, channelsRes] = await Promise.all([
          userApi.listApiKeys(),
          userApi.listChannels(),
        ])
        const nextKeys = Array.isArray(keysRes) ? keysRes : keysRes.api_keys ?? keysRes.keys ?? []
        const nextChannels = Array.isArray(channelsRes) ? channelsRes : channelsRes.channels ?? []
        setApiKeys(nextKeys)
        setChannels(nextChannels.filter((item) => item.type === 'image'))
        if (nextKeys.length > 0) setSelectedKeyId(nextKeys[0].id)
      } catch {
        // ignore
      }
    }

    void load()
  }, [])

  function currentApiKey() {
    const key = apiKeys.find((item) => item.id === selectedKeyId)
    return key?.raw_key || key?.key || ''
  }

  async function generate() {
    if (!prompt.trim()) return
    const apiKey = currentApiKey()
    if (!apiKey) return
    setRunning(true)
    setImages([])
    try {
      const response = await fetch('/v1/image', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${apiKey}`,
        },
        body: JSON.stringify({
          model: selectedModel || channels[0]?.routing_model || channels[0]?.name,
          prompt,
          size,
          aspect_ratio: aspectRatio,
        }),
      })
      const data = await response.json()
      if (data.task_id) {
        setTaskId(String(data.task_id))
      }
      if (Array.isArray(data.data)) {
        setImages(
          data.data
            .map((item: { url?: string }) => item.url)
            .filter((item: string | undefined): item is string => Boolean(item))
        )
      }
    } finally {
      setRunning(false)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Image"
        title="图片生成"
        description="接入真实 `/v1/image` 接口，支持基础参数提交和任务返回。"
      />
      <div className="grid gap-4 xl:grid-cols-[320px_1fr]">
        <Card>
          <CardContent className="space-y-4 p-6">
            <select className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm outline-none" value={selectedKeyId} onChange={(event) => setSelectedKeyId(Number(event.target.value))}>
              {apiKeys.map((key) => (
                <option key={key.id} value={key.id}>{key.name || key.masked_key || key.key}</option>
              ))}
            </select>
            <select className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm outline-none" value={selectedModel} onChange={(event) => setSelectedModel(event.target.value)}>
              <option value="">选择图片模型</option>
              {channels.map((channel) => (
                <option key={channel.id} value={channel.routing_model || channel.name}>{channel.name}</option>
              ))}
            </select>
            <Textarea value={prompt} onChange={(event) => setPrompt(event.target.value)} placeholder="描述你想生成的图片" />
            <select className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm outline-none" value={size} onChange={(event) => setSize(event.target.value)}>
              <option value="1k">1k</option>
              <option value="2k">2k</option>
              <option value="3k">3k</option>
              <option value="4k">4k</option>
            </select>
            <select className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm outline-none" value={aspectRatio} onChange={(event) => setAspectRatio(event.target.value)}>
              <option value="1:1">1:1</option>
              <option value="16:9">16:9</option>
              <option value="9:16">9:16</option>
            </select>
            <Button onClick={generate} disabled={running || !prompt.trim()}>
              {running ? '生成中...' : '生成图片'}
            </Button>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="space-y-4 p-6">
            {taskId ? <p className="text-sm text-muted-foreground">任务 ID：{taskId}</p> : null}
            {images.length > 0 ? (
              <div className="grid gap-4 md:grid-cols-2">
                {images.map((url) => (
                  <img key={url} className="rounded-xl border border-border/70" src={url} alt="generated" />
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">提交后将在这里展示结果。</p>
            )}
          </CardContent>
        </Card>
      </div>
    </>
  )
}
