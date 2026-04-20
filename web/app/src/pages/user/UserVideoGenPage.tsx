import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import { userApi, type ApiKeyRecord, type UserChannel } from '@/lib/api/user'

export function UserVideoGenPage() {
  const [apiKeys, setApiKeys] = useState<ApiKeyRecord[]>([])
  const [channels, setChannels] = useState<UserChannel[]>([])
  const [selectedKeyId, setSelectedKeyId] = useState<number | undefined>()
  const [selectedModel, setSelectedModel] = useState('')
  const [prompt, setPrompt] = useState('')
  const [size, setSize] = useState('720p')
  const [aspectRatio, setAspectRatio] = useState('16:9')
  const [duration, setDuration] = useState('5')
  const [taskId, setTaskId] = useState('')
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
        setChannels(nextChannels.filter((item) => item.type === 'video'))
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
    try {
      const response = await fetch('/v1/video', {
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
          duration,
        }),
      })
      const data = await response.json()
      setTaskId(String(data.task_id ?? ''))
    } finally {
      setRunning(false)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Video"
        title="视频生成"
        description="接入真实 `/v1/video`，支持发起生成任务并返回任务 ID。"
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
              <option value="">选择视频模型</option>
              {channels.map((channel) => (
                <option key={channel.id} value={channel.routing_model || channel.name}>{channel.name}</option>
              ))}
            </select>
            <Textarea value={prompt} onChange={(event) => setPrompt(event.target.value)} placeholder="描述你想生成的视频" />
            <select className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm outline-none" value={size} onChange={(event) => setSize(event.target.value)}>
              <option value="720p">720p</option>
              <option value="1080p">1080p</option>
            </select>
            <select className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm outline-none" value={aspectRatio} onChange={(event) => setAspectRatio(event.target.value)}>
              <option value="16:9">16:9</option>
              <option value="9:16">9:16</option>
              <option value="1:1">1:1</option>
            </select>
            <select className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm outline-none" value={duration} onChange={(event) => setDuration(event.target.value)}>
              <option value="5">5 秒</option>
              <option value="10">10 秒</option>
              <option value="15">15 秒</option>
            </select>
            <Button onClick={generate} disabled={running || !prompt.trim()}>
              {running ? '生成中...' : '生成视频'}
            </Button>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            {taskId ? (
              <p className="text-sm text-muted-foreground">任务已创建，任务 ID：{taskId}</p>
            ) : (
              <p className="text-sm text-muted-foreground">提交后将在这里显示任务信息。</p>
            )}
          </CardContent>
        </Card>
      </div>
    </>
  )
}
