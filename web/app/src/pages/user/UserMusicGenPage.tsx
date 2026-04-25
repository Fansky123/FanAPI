import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { NativeSelect } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { userApi, type ApiKeyRecord, type UserChannel } from '@/lib/api/user'

type MusicMode = '10' | '20'

type MusicItem = {
  audio_url?: string
  video_url?: string
  image_url?: string
  title?: string
  duration?: number | string
  tags?: string
}

export function UserMusicGenPage() {
  const [apiKeys, setApiKeys] = useState<ApiKeyRecord[]>([])
  const [channels, setChannels] = useState<UserChannel[]>([])
  const [selectedKeyId, setSelectedKeyId] = useState<number | undefined>()
  const [selectedChannelId, setSelectedChannelId] = useState<number | undefined>()
  const [error, setError] = useState('')

  const [mode, setMode] = useState<MusicMode>('10')
  const [gptDescription, setGptDescription] = useState('')
  const [lyrics, setLyrics] = useState('')
  const [tags, setTags] = useState('')
  const [title, setTitle] = useState('')
  const [makeInstrumental, setMakeInstrumental] = useState(false)

  const [taskId, setTaskId] = useState('')
  const [taskStatus, setTaskStatus] = useState<'idle' | 'polling' | 'done' | 'failed'>('idle')
  const [items, setItems] = useState<MusicItem[]>([])
  const [taskError, setTaskError] = useState('')
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
          (item) => item.type === 'music'
        )
        setApiKeys(nextKeys)
        setChannels(nextChannels)
        if (nextKeys.length > 0) setSelectedKeyId(nextKeys[0].id)
        if (nextChannels.length > 0) setSelectedChannelId(nextChannels[0].id)
      } catch {
        setError('读取音乐渠道或 API 密钥失败')
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

  useEffect(() => {
    if (!taskId || taskStatus !== 'polling') return
    const key = apiKeys.find((item) => item.id === selectedKeyId)
    const apiKey = key?.raw_key || key?.key || ''
    if (!apiKey) return
    let cancelled = false

    const tick = async () => {
      try {
        const resp = await fetch(`/v1/tasks/${taskId}`, {
          headers: { Authorization: `Bearer ${apiKey}` },
        })
        if (!resp.ok) return
        const data = await resp.json()
        if (cancelled) return
        if (data.code === 200 || data.status === 2) {
          const list: MusicItem[] = Array.isArray(data.items)
            ? data.items
            : Array.isArray(data.result?.items)
              ? data.result.items
              : []
          setItems(list)
          setTaskStatus('done')
          setRunning(false)
        } else if ((typeof data.code === 'number' && data.code >= 400) || data.status === 3) {
          setTaskError(data.msg || '生成失败')
          setTaskStatus('failed')
          setRunning(false)
        }
      } catch {
        // ignore polling failures
      }
    }

    const timer = setInterval(tick, 3000)
    return () => {
      cancelled = true
      clearInterval(timer)
    }
  }, [taskId, taskStatus, apiKeys, selectedKeyId])

  async function generate() {
    const apiKey = currentApiKey()
    if (!apiKey) {
      setError('请选择可直接调用的 API 密钥')
      return
    }
    if (!selectedChannelId && channels.length === 0) {
      setError('当前没有可用的音乐模型渠道')
      return
    }
    if (mode === '10' && !gptDescription.trim()) {
      setError('请输入灵感描述')
      return
    }

    setError('')
    setItems([])
    setTaskError('')
    setTaskStatus('idle')
    setTaskId('')
    setRunning(true)

    const body: Record<string, unknown> = {
      model: currentChannel()?.routing_model || currentChannel()?.name,
      input_type: mode,
      make_instrumental: makeInstrumental,
    }
    if (mode === '10') {
      body.gpt_description_prompt = gptDescription
    } else {
      body.prompt = lyrics
      body.tags = tags
      body.title = title
    }

    try {
      const endpoint = currentChannel()?.id
        ? `/v1/music?channel_id=${currentChannel()?.id}`
        : '/v1/music'
      const resp = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${apiKey}`,
        },
        body: JSON.stringify(body),
      })
      const data = await resp.json()
      if (data.task_id) {
        setTaskId(String(data.task_id))
        setTaskStatus('polling')
      } else if (Array.isArray(data.items)) {
        setItems(data.items)
        setTaskStatus('done')
        setRunning(false)
      } else {
        setTaskError(data.msg || `生成失败 (HTTP ${resp.status})`)
        setTaskStatus('failed')
        setRunning(false)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '请求失败')
      setRunning(false)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Music"
        title="音乐生成"
        description="接入 `/v1/music`，支持灵感模式和自定义模式两种创作方式。"
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
              <Label>API 密钥</Label>
              <NativeSelect value={selectedKeyId} onChange={(event) => setSelectedKeyId(Number(event.target.value))}>
                {apiKeys.map((key) => (
                  <option key={key.id} value={key.id}>{key.name || key.masked_key || key.key}</option>
                ))}
              </NativeSelect>
            </div>
            <div className="grid gap-1.5">
              <Label>模型</Label>
              <NativeSelect value={selectedChannelId} onChange={(event) => setSelectedChannelId(Number(event.target.value))}>
                {channels.map((channel) => (
                  <option key={channel.id} value={channel.id}>{channel.name}</option>
                ))}
              </NativeSelect>
              {channels.length === 0 ? (
                <p className="text-xs text-muted-foreground">当前没有可用的音乐模型渠道。</p>
              ) : null}
            </div>
            <div className="grid gap-1.5">
              <Label>创作模式</Label>
              <div className="flex rounded-md border border-input p-0.5">
                <button
                  type="button"
                  onClick={() => setMode('10')}
                  className={`flex-1 rounded-sm px-3 py-1.5 text-xs font-medium transition ${
                    mode === '10' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  灵感模式
                </button>
                <button
                  type="button"
                  onClick={() => setMode('20')}
                  className={`flex-1 rounded-sm px-3 py-1.5 text-xs font-medium transition ${
                    mode === '20' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  自定义模式
                </button>
              </div>
            </div>

            {mode === '10' ? (
              <div className="grid gap-1.5">
                <Label>灵感描述 <span className="text-destructive">*</span></Label>
                <Textarea
                  rows={5}
                  value={gptDescription}
                  onChange={(event) => setGptDescription(event.target.value)}
                  placeholder="例如：一首轻快的电子流行乐，节奏明快"
                />
              </div>
            ) : (
              <>
                <div className="grid gap-1.5">
                  <Label>歌词 <span className="text-muted-foreground font-normal">(选填)</span></Label>
                  <Textarea
                    rows={4}
                    value={lyrics}
                    onChange={(event) => setLyrics(event.target.value)}
                    placeholder="填写歌词内容，支持 [Verse]/[Chorus] 等标记"
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label>风格标签 <span className="text-muted-foreground font-normal">(选填)</span></Label>
                  <Input
                    value={tags}
                    onChange={(event) => setTags(event.target.value)}
                    placeholder="如：pop, female voice, upbeat"
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label>歌曲标题 <span className="text-muted-foreground font-normal">(选填)</span></Label>
                  <Input
                    value={title}
                    onChange={(event) => setTitle(event.target.value)}
                    placeholder="歌曲名称"
                  />
                </div>
              </>
            )}

            <Label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={makeInstrumental}
                onChange={(event) => setMakeInstrumental(event.target.checked)}
              />
              纯音乐（无人声）
            </Label>

            <Button
              onClick={generate}
              disabled={running || !currentApiKey() || channels.length === 0}
              className="w-full"
            >
              {running ? '生成中...' : '生成音乐'}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="flex flex-col gap-4 p-6">
            {taskStatus === 'idle' && items.length === 0 ? (
              <p className="text-sm text-muted-foreground">提交后将在这里显示生成的音乐。</p>
            ) : null}
            {taskStatus === 'polling' ? (
              <Alert>
                <AlertDescription>
                  任务生成中，任务 ID：<span className="font-mono">{taskId}</span>。完成后将自动展示。
                </AlertDescription>
              </Alert>
            ) : null}
            {taskStatus === 'failed' && taskError ? (
              <Alert variant="destructive">
                <AlertDescription>{taskError}</AlertDescription>
              </Alert>
            ) : null}
            {items.map((item, index) => (
              <div key={index} className="flex flex-col gap-3 rounded-lg border border-border/70 p-4">
                {item.image_url ? (
                  <img src={item.image_url} alt={item.title || '封面'} className="h-32 w-32 rounded-md object-cover" />
                ) : null}
                <div className="text-sm font-medium">{item.title || `音乐 ${index + 1}`}</div>
                {item.tags ? <div className="text-xs text-muted-foreground">{item.tags}</div> : null}
                {item.audio_url ? (
                  <audio controls src={item.audio_url} className="w-full">
                    <track kind="captions" />
                  </audio>
                ) : null}
                {item.video_url ? (
                  <a
                    href={item.video_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-xs text-primary underline"
                  >
                    查看视频版本 →
                  </a>
                ) : null}
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </>
  )
}
