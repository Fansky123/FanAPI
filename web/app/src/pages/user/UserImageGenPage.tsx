import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { NativeSelect } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { userApi, type ApiKeyRecord, type UserChannel } from '@/lib/api/user'

export function UserImageGenPage() {
  const [apiKeys, setApiKeys] = useState<ApiKeyRecord[]>([])
  const [channels, setChannels] = useState<UserChannel[]>([])
  const [selectedKeyId, setSelectedKeyId] = useState<number | undefined>()
  const [selectedChannelId, setSelectedChannelId] = useState<number | undefined>()
  const [error, setError] = useState('')
  const [prompt, setPrompt] = useState('')
  const [size, setSize] = useState('1k')
  const [aspectRatio, setAspectRatio] = useState('1:1')
  const [referenceImages, setReferenceImages] = useState('')
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
        const nextChannels = (Array.isArray(channelsRes) ? channelsRes : channelsRes.channels ?? []).filter(
          (item) => item.type === 'image'
        )
        setApiKeys(nextKeys)
        setChannels(nextChannels)
        if (nextKeys.length > 0) setSelectedKeyId(nextKeys[0].id)
        if (nextChannels.length > 0) setSelectedChannelId(nextChannels[0].id)
      } catch {
        setError('读取图片渠道或 API 密钥失败')
      }
    }

    void load()
  }, [])

  function currentApiKey() {
    const key = apiKeys.find((item) => item.id === selectedKeyId)
    return key?.raw_key || key?.key || ''
  }

  function currentChannel() {
    return channels.find((item) => item.id === selectedChannelId) ?? channels[0]
  }

  async function generate() {
    if (!prompt.trim()) return
    const apiKey = currentApiKey()
    if (!apiKey) {
      setError('请选择可直接调用的 API 密钥')
      return
    }
    if (!selectedChannelId && channels.length === 0) {
      setError('当前没有可用的图片模型渠道')
      return
    }
    setRunning(true)
    setImages([])
    setTaskId('')
    setError('')
    try {
      const endpoint = currentChannel()?.id
        ? `/v1/image?channel_id=${currentChannel()?.id}`
        : '/v1/image'
      const refUrls = referenceImages
        .split('\n')
        .map((line) => line.trim())
        .filter(Boolean)
      const body: Record<string, unknown> = {
        model: currentChannel()?.routing_model || currentChannel()?.name,
        prompt,
        size,
        aspect_ratio: aspectRatio,
      }
      if (refUrls.length > 0) body.reference_images = refUrls
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${apiKey}`,
        },
        body: JSON.stringify(body),
      })
      if (!response.ok) {
        throw new Error((await response.text()) || `请求失败 (${response.status})`)
      }
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
    } catch (err) {
      setError(err instanceof Error ? err.message : '图片生成失败')
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
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <div className="grid gap-4 xl:grid-cols-[320px_1fr]">
        <Card>
          <CardContent className="flex flex-col gap-4 p-6">
            <div className="grid gap-1.5">
              <Label>API 密钥 <span className="text-destructive">*</span></Label>
              <NativeSelect value={selectedKeyId} onChange={(event) => setSelectedKeyId(Number(event.target.value))}>
                {apiKeys.map((key) => (
                  <option key={key.id} value={key.id}>{key.name || key.masked_key || key.key}</option>
                ))}
              </NativeSelect>
            </div>
            <div className="grid gap-1.5">
              <Label>模型 <span className="text-muted-foreground font-normal">(选填)</span></Label>
              <NativeSelect value={selectedChannelId} onChange={(event) => setSelectedChannelId(Number(event.target.value))}>
                {channels.map((channel) => (
                  <option key={channel.id} value={channel.id}>{channel.name}</option>
                ))}
              </NativeSelect>
              {channels.length === 0 ? (
                <p className="text-xs text-muted-foreground">当前没有可用的图片模型渠道。</p>
              ) : null}
            </div>
            <div className="grid gap-1.5">
              <Label>提示词 <span className="text-destructive">*</span></Label>
              <Textarea
                rows={5}
                value={prompt}
                onChange={(event) => setPrompt(event.target.value)}
                placeholder="描述你想生成的图片内容..."
              />
            </div>
            <div className="grid gap-1.5">
              <Label>分辨率档位</Label>
              <NativeSelect value={size} onChange={(event) => setSize(event.target.value)}>
                <option value="1k">1k (1024px)</option>
                <option value="2k">2k (2048px)</option>
                <option value="3k">3k (3072px)</option>
                <option value="4k">4k (4096px)</option>
              </NativeSelect>
            </div>
            <div className="grid gap-1.5">
              <Label>宽高比</Label>
              <NativeSelect value={aspectRatio} onChange={(event) => setAspectRatio(event.target.value)}>
                <option value="1:1">1:1 方图</option>
                <option value="16:9">16:9 横版</option>
                <option value="9:16">9:16 竖版</option>
                <option value="4:3">4:3</option>
                <option value="3:4">3:4</option>
                <option value="3:2">3:2</option>
                <option value="2:3">2:3</option>
                <option value="21:9">21:9 超宽</option>
              </NativeSelect>
            </div>
            <div className="grid gap-1.5">
              <Label>参考图 URL <span className="text-muted-foreground font-normal">(选填，每行一条)</span></Label>
              <Textarea
                rows={3}
                value={referenceImages}
                onChange={(event) => setReferenceImages(event.target.value)}
                placeholder={'https://example.com/ref1.png\nhttps://example.com/ref2.png'}
              />
            </div>
            <Button onClick={generate} disabled={running || !prompt.trim() || !currentApiKey() || channels.length === 0}>
              {running ? '生成中...' : '生成图片'}
            </Button>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex flex-col gap-4 p-6">
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
