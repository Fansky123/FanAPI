import { useEffect, useRef, useState } from 'react'

import { MessageContent } from '@/components/shared/MessageContent'
import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { NativeSelect } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { userApi, type ApiKeyRecord, type UserChannel } from '@/lib/api/user'

type Message = {
  role: 'user' | 'assistant'
  content: string
}

export function UserPlaygroundPage() {
  const [apiKeys, setApiKeys] = useState<ApiKeyRecord[]>([])
  const [channels, setChannels] = useState<UserChannel[]>([])
  const [selectedKeyId, setSelectedKeyId] = useState<number | undefined>()
  const [selectedChannelId, setSelectedChannelId] = useState<number | undefined>()
  const [error, setError] = useState('')
  const [systemPrompt, setSystemPrompt] = useState('')
  const [inputText, setInputText] = useState('')
  const [messages, setMessages] = useState<Message[]>([])
  const [streaming, setStreaming] = useState(false)
  const [streamingText, setStreamingText] = useState('')
  const scrollRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    async function load() {
      try {
        const [keysRes, channelsRes] = await Promise.all([
          userApi.listApiKeys(),
          userApi.listChannels(),
        ])
        const nextKeys = Array.isArray(keysRes) ? keysRes : keysRes.api_keys ?? keysRes.keys ?? []
        const nextChannels = (Array.isArray(channelsRes) ? channelsRes : channelsRes.channels ?? []).filter(
          (item) => item.type === 'llm'
        )
        setApiKeys(nextKeys)
        setChannels(nextChannels)
        if (nextKeys.length > 0) setSelectedKeyId(nextKeys[0].id)
        if (nextChannels.length > 0) setSelectedChannelId(nextChannels[0].id)
      } catch {
        setError('读取 API 密钥或模型列表失败')
      }
    }

    void load()
  }, [])

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight, behavior: 'smooth' })
  }, [messages, streamingText])

  function currentApiKey() {
    const key = apiKeys.find((item) => item.id === selectedKeyId)
    return key?.raw_key || key?.key || ''
  }

  function currentChannel() {
    return channels.find((item) => item.id === selectedChannelId) ?? channels[0]
  }

  async function sendMessage() {
    if (!inputText.trim() || streaming) return
    const apiKey = currentApiKey()
    if (!apiKey || apiKey.includes('...')) {
      setError('请选择可直接调用的 API 密钥')
      return
    }
    if (!selectedChannelId && channels.length === 0) {
      setError('当前没有可用的文本模型渠道')
      return
    }

    const userMessage: Message = { role: 'user', content: inputText.trim() }
    const nextMessages = [...messages, userMessage]
    setError('')
    setMessages(nextMessages)
    setInputText('')
    setStreaming(true)
    setStreamingText('')

    const body = {
      model:
        currentChannel()?.routing_model ||
        currentChannel()?.name ||
        'gpt-3.5-turbo',
      messages: [
        ...(systemPrompt.trim()
          ? [{ role: 'system', content: systemPrompt.trim() }]
          : []),
        ...nextMessages,
      ],
      stream: true,
    }

    try {
      const channel = currentChannel()
      const endpoint = channel?.id
        ? `/v1/chat/completions?channel_id=${channel.id}`
        : '/v1/chat/completions'
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${apiKey}`,
        },
        body: JSON.stringify(body),
      })

      if (!response.ok || !response.body) {
        throw new Error((await response.text()) || `请求失败 (${response.status})`)
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let accum = ''

      while (true) {
        const { value, done } = await reader.read()
        if (done) break
        const chunk = decoder.decode(value)
        const lines = chunk.split('\n')
        for (const line of lines) {
          if (!line.startsWith('data: ')) continue
          const data = line.slice(6).trim()
          if (data === '[DONE]') continue
          try {
            const parsed = JSON.parse(data)
            const delta = parsed.choices?.[0]?.delta?.content || ''
            accum += delta
            setStreamingText(accum)
          } catch {
            // skip malformed chunks
          }
        }
      }

      setMessages((current) => [...current, { role: 'assistant', content: accum }])
      setStreamingText('')
    } catch (error) {
      setMessages((current) => [
        ...current,
        { role: 'assistant', content: `请求失败：${error instanceof Error ? error.message : '未知错误'}` },
      ])
      setError(error instanceof Error ? error.message : '请求失败')
    } finally {
      setStreaming(false)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Playground"
        title="文本对话"
        description="已经接上真实 `/v1/chat/completions`，可直接用已有 API Key 做对话验证。"
        actions={<Button onClick={() => setMessages([])}>新对话</Button>}
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <div className="grid gap-4 xl:grid-cols-[320px_1fr]">
        <Card>
          <CardContent className="flex flex-col gap-4 p-6">
            <div className="flex flex-col gap-2">
              <Label>API 密钥</Label>
              <NativeSelect
                value={selectedKeyId}
                onChange={(event) => setSelectedKeyId(Number(event.target.value))}
              >
                {apiKeys.map((key) => (
                  <option key={key.id} value={key.id}>
                    {key.name || key.masked_key || key.key}
                  </option>
                ))}
              </NativeSelect>
              <p className="text-xs text-muted-foreground">
                仅新创建的密钥含明文，旧密钥因安全原因不可直接用于此页面。
              </p>
            </div>
            <div className="flex flex-col gap-2">
              <Label>模型</Label>
              <NativeSelect
                value={selectedChannelId}
                onChange={(event) => setSelectedChannelId(Number(event.target.value))}
              >
                {channels.map((channel) => (
                  <option key={channel.id} value={channel.id}>
                    {channel.name} {channel.routing_model ? `· ${channel.routing_model}` : ''}
                  </option>
                ))}
              </NativeSelect>
              {channels.length === 0 ? (
                <p className="text-xs text-muted-foreground">当前没有可用的文本模型渠道。</p>
              ) : null}
            </div>
            <div className="flex flex-col gap-2">
              <Label>系统提示词</Label>
              <Textarea
                value={systemPrompt}
                onChange={(event) => setSystemPrompt(event.target.value)}
                placeholder="例如：你是一个专业的 AI 助手"
              />
            </div>
          </CardContent>
        </Card>
        <Card className="flex min-h-[70vh] flex-col overflow-hidden">
          <CardContent className="flex min-h-0 flex-1 flex-col p-0">
            <div ref={scrollRef} className="flex-1 flex flex-col gap-4 overflow-auto p-6">
              {messages.length === 0 && !streaming ? (
                <div className="flex min-h-[300px] items-center justify-center text-sm text-muted-foreground">
                  开始一段对话吧。
                </div>
              ) : null}
              {messages.map((message, index) => (
                <div
                  key={`${message.role}-${index}`}
                  className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}
                >
                  <div
                    className={`max-w-[80%] rounded-2xl px-4 py-3 text-sm leading-7 ${
                      message.role === 'user'
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-muted text-foreground'
                    }`}
                  >
                    <MessageContent content={message.content} role={message.role} />
                  </div>
                </div>
              ))}
              {streaming && streamingText ? (
                <div className="flex justify-start">
                  <div className="max-w-[80%] rounded-2xl bg-muted px-4 py-3 text-sm leading-7">
                    <MessageContent content={streamingText} role="assistant" />
                  </div>
                </div>
              ) : null}
            </div>
            <div className="border-t border-border/70 p-4">
              <div className="flex gap-3">
                <Textarea
                  className="min-h-24 flex-1"
                  value={inputText}
                  onChange={(event) => setInputText(event.target.value)}
                  placeholder="输入消息，Enter 发送"
                />
                <Button
                  onClick={sendMessage}
                  disabled={streaming || !inputText.trim() || !currentApiKey() || channels.length === 0}
                >
                  {streaming ? '生成中...' : '发送'}
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  )
}
