import { useEffect, useRef, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
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
  const [selectedModel, setSelectedModel] = useState('')
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
        const nextChannels = Array.isArray(channelsRes) ? channelsRes : channelsRes.channels ?? []
        setApiKeys(nextKeys)
        setChannels(nextChannels.filter((item) => item.type === 'llm'))
        if (nextKeys.length > 0) setSelectedKeyId(nextKeys[0].id)
      } catch {
        // keep empty state
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

  async function sendMessage() {
    if (!inputText.trim() || streaming) return
    const apiKey = currentApiKey()
    if (!apiKey || apiKey.includes('...')) return

    const userMessage: Message = { role: 'user', content: inputText.trim() }
    const nextMessages = [...messages, userMessage]
    setMessages(nextMessages)
    setInputText('')
    setStreaming(true)
    setStreamingText('')

    const body = {
      model: selectedModel || channels[0]?.routing_model || channels[0]?.name || 'gpt-3.5-turbo',
      messages: [
        ...(systemPrompt.trim()
          ? [{ role: 'system', content: systemPrompt.trim() }]
          : []),
        ...nextMessages,
      ],
      stream: true,
    }

    try {
      const response = await fetch('/v1/chat/completions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${apiKey}`,
        },
        body: JSON.stringify(body),
      })

      if (!response.ok || !response.body) {
        throw new Error(await response.text())
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
      <div className="grid gap-4 xl:grid-cols-[320px_1fr]">
        <Card>
          <CardContent className="space-y-4 p-6">
            <div className="space-y-2">
              <label className="text-sm font-medium">API 密钥</label>
              <select
                className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm outline-none"
                value={selectedKeyId}
                onChange={(event) => setSelectedKeyId(Number(event.target.value))}
              >
                {apiKeys.map((key) => (
                  <option key={key.id} value={key.id}>
                    {key.name || key.masked_key || key.key}
                  </option>
                ))}
              </select>
              <p className="text-xs text-muted-foreground">
                仅新创建的密钥含明文，旧密钥因安全原因不可直接用于此页面。
              </p>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">模型</label>
              <select
                className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm outline-none"
                value={selectedModel}
                onChange={(event) => setSelectedModel(event.target.value)}
              >
                <option value="">自动选择</option>
                {channels.map((channel) => (
                  <option key={channel.id} value={channel.routing_model || channel.name}>
                    {channel.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">系统提示词</label>
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
            <div ref={scrollRef} className="flex-1 space-y-4 overflow-auto p-6">
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
                    {message.content}
                  </div>
                </div>
              ))}
              {streaming && streamingText ? (
                <div className="flex justify-start">
                  <div className="max-w-[80%] rounded-2xl bg-muted px-4 py-3 text-sm leading-7">
                    {streamingText}
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
                <Button onClick={sendMessage} disabled={streaming || !inputText.trim()}>
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
