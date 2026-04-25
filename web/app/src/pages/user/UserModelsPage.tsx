import { useMemo, useState } from 'react'
import { BlocksIcon, Copy, Search, TerminalSquare } from 'lucide-react'

import { EmptyState } from '@/components/shared/EmptyState'
import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { copyToClipboard } from '@/lib/clipboard'
import { useAsync } from '@/hooks/use-async'
import { userApi, type UserChannel } from '@/lib/api/user'
import { cn } from '@/lib/utils'

type DocMode = 'channel' | 'balance' | 'task'
type LangTab = 'curl' | 'python' | 'php' | 'go' | 'java'
type SunoMode = 'inspire' | 'custom' | 'extend' | 'overpainting' | 'underpainting'

const typeOptions = [
  { label: '全部', value: '' },
  { label: 'LLM 对话', value: 'llm' },
  { label: 'Image 绘图', value: 'image' },
  { label: 'Video 视频', value: 'video' },
  { label: 'Audio 语音', value: 'audio' },
  { label: 'Music 音乐', value: 'music' },
]

const protocolLabels: Record<string, string> = {
  openai: 'OpenAI 兼容',
  claude: 'Claude 原生',
  gemini: 'Gemini 原生',
}

const billingTypeLabels: Record<string, string> = {
  token: 'Token 计费',
  image: '按图计费',
  video: '按秒计费',
  audio: '按秒计费',
  music: '按任务计费',
  count: '按次计费',
  custom: '自定义计费',
}

function copyText(text: string, label = '已复制') {
  void copyToClipboard(text, { successMessage: label })
}

function buildChannelRequestBody(channel: UserChannel, sunoMode: SunoMode) {
  const model = channel.routing_model || channel.name
  if (channel.type === 'llm') {
    if (channel.protocol === 'gemini') {
      return JSON.stringify({
        contents: [{ role: 'user', parts: [{ text: '你好，请介绍一下自己' }] }],
      }, null, 2)
    }
    return JSON.stringify({
      model,
      messages: [{ role: 'user', content: '你好，请介绍一下自己' }],
      stream: false,
    }, null, 2)
  }
  if (channel.type === 'image') {
    return JSON.stringify({
      model,
      prompt: '一只可爱的橘猫坐在阳光下',
      size: '1k',
      aspect_ratio: '1:1',
      n: 1,
    }, null, 2)
  }
  if (channel.type === 'video') {
    return JSON.stringify({
      model,
      prompt: '海浪拍打岸边，夕阳西下',
      size: '720p',
      aspect_ratio: '16:9',
      duration: '5',
    }, null, 2)
  }
  if (channel.type === 'audio') {
    return JSON.stringify({
      model,
      input: '你好，欢迎使用语音合成服务',
      voice: 'alloy',
    }, null, 2)
  }
  if (channel.type === 'music') {
    if (sunoMode === 'custom') {
      return JSON.stringify({
        model,
        input_type: '20',
        prompt: '[主歌]\n周四的阳光晒脸庞\n微风轻轻吹过窗\n\n[副歌]\n周四快乐不散场\n欢声笑语满心房',
        title: '周四快乐',
        tags: 'pop,female voice',
        mv_version: 'chirp-v5',
        make_instrumental: false,
      }, null, 2)
    }
    if (sunoMode === 'extend') {
      return JSON.stringify({
        model,
        input_type: '20',
        prompt: '[Verse 1]\n小狗汪汪叫\n尾巴甩甩跳\n\n[Chorus]\n汪汪汪谁在听\n汪汪汪快乐行',
        title: '为你歌唱',
        tags: '',
        mv_version: 'chirp-v5',
        make_instrumental: false,
        continue_clip_id: 'https://cdn1.suno.ai/7c395650-62f2-4c4f-8b68-cf55b874c96c.mp3',
        continue_at: '27',
      }, null, 2)
    }
    if (sunoMode === 'overpainting') {
      return JSON.stringify({
        model,
        input_type: '20',
        prompt: '[Verse 1]\nUsah lepas kau pergi',
        tags: 'pop,female voice',
        title: 'Hi, melancholic',
        task: 'overpainting',
        metadata_params: {
          overpainting_clip_id: 'https://cdn1.suno.ai/21ae9c64-86ab-435a-b810-ed62727caf0a.mp3',
          overpainting_start_s: 0,
          overpainting_end_s: 57.9,
        },
      }, null, 2)
    }
    if (sunoMode === 'underpainting') {
      return JSON.stringify({
        model,
        input_type: '20',
        prompt: '',
        tags: 'pop,female voice',
        title: 'Hi, melancholic',
        task: 'underpainting',
        make_instrumental: true,
        metadata_params: {
          underpainting_clip_id: 'https://cdn1.suno.ai/21ae9c64-86ab-435a-b810-ed62727caf0a.mp3',
          underpainting_start_s: 0,
          underpainting_end_s: 57.9,
        },
      }, null, 2)
    }
    return JSON.stringify({
      model,
      input_type: '10',
      gpt_description_prompt: '轻快的爵士乐，适合咖啡馆氛围，女声演唱',
      mv_version: 'chirp-v5',
      make_instrumental: false,
    }, null, 2)
  }
  return JSON.stringify({ model, prompt: '...' }, null, 2)
}

function getChannelEndpoint(channel: UserChannel) {
  const endpointMap: Record<string, string> = {
    llm: '/v1/chat/completions',
    image: '/v1/image',
    video: '/v1/video',
    audio: '/v1/audio',
    music: '/v1/music',
  }
  if (channel.type === 'llm' && channel.protocol === 'gemini') {
    const model = channel.routing_model || channel.name
    return `/v1beta/models/${model}:generateContent`
  }
  return endpointMap[channel.type ?? 'llm'] || '/v1/chat/completions'
}

function getChannelCode(channel: UserChannel, lang: LangTab, sunoMode: SunoMode) {
  const origin = window.location.origin
  const endpoint = getChannelEndpoint(channel)
  const body = buildChannelRequestBody(channel, sunoMode)
  if (lang === 'curl') {
    return `curl -X POST "${origin}${endpoint}" \\\n+  -H "Content-Type: application/json" \\\n+  -H "Authorization: Bearer YOUR_API_KEY" \\\n+  -d '${body}'`
  }
  if (lang === 'python') {
    return `import requests\nimport json\n\nurl = "${origin}${endpoint}"\nheaders = {\n    "Authorization": "Bearer YOUR_API_KEY",\n    "Content-Type": "application/json"\n}\nbody = json.loads('''${body}''')\n\nresponse = requests.post(url, headers=headers, json=body)\nprint(response.json())`
  }
  if (lang === 'php') {
    const safeBody = body.replace(/'/g, "\\'")
    return `<?php\n$url = "${origin}${endpoint}";\n$body = '${safeBody}';\n\n$ch = curl_init($url);\ncurl_setopt_array($ch, [\n    CURLOPT_RETURNTRANSFER => true,\n    CURLOPT_POST           => true,\n    CURLOPT_HTTPHEADER     => [\n        'Authorization: Bearer YOUR_API_KEY',\n        'Content-Type: application/json',\n    ],\n    CURLOPT_POSTFIELDS     => $body,\n]);\n\n$response = curl_exec($ch);\ncurl_close($ch);\necho $response;`
  }
  if (lang === 'go') {
    return `package main\n\nimport (\n\t"bytes"\n\t"fmt"\n\t"io"\n\t"net/http"\n)\n\nfunc main() {\n\tbody := []byte(\`${body}\`)\n\n\treq, _ := http.NewRequest("POST", "${origin}${endpoint}", bytes.NewBuffer(body))\n\treq.Header.Set("Authorization", "Bearer YOUR_API_KEY")\n\treq.Header.Set("Content-Type", "application/json")\n\n\tresp, _ := (&http.Client{}).Do(req)\n\tdefer resp.Body.Close()\n\tdata, _ := io.ReadAll(resp.Body)\n\tfmt.Println(string(data))\n}`
  }
  const escapedBody = body.replace(/\\/g, '\\\\').replace(/"/g, '\\"').replace(/\n/g, '\\n')
  return `import java.net.http.*;\nimport java.net.URI;\n\npublic class Main {\n    public static void main(String[] args) throws Exception {\n        String body = "${escapedBody}";\n\n        var request = HttpRequest.newBuilder()\n            .uri(URI.create("${origin}${endpoint}"))\n            .header("Authorization", "Bearer YOUR_API_KEY")\n            .header("Content-Type", "application/json")\n            .POST(HttpRequest.BodyPublishers.ofString(body))\n            .build();\n\n        var response = HttpClient.newHttpClient()\n            .send(request, HttpResponse.BodyHandlers.ofString());\n        System.out.println(response.body());\n    }\n}`
}

function getBalanceCode(lang: LangTab) {
  const origin = window.location.origin
  if (lang === 'curl') return `curl -X GET "${origin}/user/balance" \\\n+  -H "Authorization: Bearer YOUR_API_KEY"`
  if (lang === 'python') return `import requests\n\nurl = "${origin}/user/balance"\nheaders = {"Authorization": "Bearer YOUR_API_KEY"}\n\nresponse = requests.get(url, headers=headers)\nprint(response.json())`
  if (lang === 'php') return `<?php\n$url = "${origin}/user/balance";\n\n$ch = curl_init($url);\ncurl_setopt_array($ch, [\n    CURLOPT_RETURNTRANSFER => true,\n    CURLOPT_HTTPHEADER     => ['Authorization: Bearer YOUR_API_KEY'],\n]);\n\necho curl_exec($ch);\ncurl_close($ch);`
  if (lang === 'go') return `package main\n\nimport (\n\t"fmt"\n\t"io"\n\t"net/http"\n)\n\nfunc main() {\n\treq, _ := http.NewRequest("GET", "${origin}/user/balance", nil)\n\treq.Header.Set("Authorization", "Bearer YOUR_API_KEY")\n\n\tresp, _ := (&http.Client{}).Do(req)\n\tdefer resp.Body.Close()\n\tdata, _ := io.ReadAll(resp.Body)\n\tfmt.Println(string(data))\n}`
  return `import java.net.http.*;\nimport java.net.URI;\n\npublic class Main {\n    public static void main(String[] args) throws Exception {\n        var request = HttpRequest.newBuilder()\n            .uri(URI.create("${origin}/user/balance"))\n            .header("Authorization", "Bearer YOUR_API_KEY")\n            .GET()\n            .build();\n\n        var response = HttpClient.newHttpClient()\n            .send(request, HttpResponse.BodyHandlers.ofString());\n        System.out.println(response.body());\n    }\n}`
}

function getTaskCode(lang: LangTab) {
  const origin = window.location.origin
  if (lang === 'curl') return `curl -X GET "${origin}/v1/tasks/YOUR_TASK_ID" \\\n+  -H "Authorization: Bearer YOUR_API_KEY"`
  if (lang === 'python') return `import requests\n\nurl = "${origin}/v1/tasks/YOUR_TASK_ID"\nheaders = {"Authorization": "Bearer YOUR_API_KEY"}\n\nresponse = requests.get(url, headers=headers)\nprint(response.json())`
  if (lang === 'php') return `<?php\n$url = "${origin}/v1/tasks/YOUR_TASK_ID";\n\n$ch = curl_init($url);\ncurl_setopt_array($ch, [\n    CURLOPT_RETURNTRANSFER => true,\n    CURLOPT_HTTPHEADER     => ['Authorization: Bearer YOUR_API_KEY'],\n]);\n\necho curl_exec($ch);\ncurl_close($ch);`
  if (lang === 'go') return `package main\n\nimport (\n\t"fmt"\n\t"io"\n\t"net/http"\n)\n\nfunc main() {\n\treq, _ := http.NewRequest("GET", "${origin}/v1/tasks/YOUR_TASK_ID", nil)\n\treq.Header.Set("Authorization", "Bearer YOUR_API_KEY")\n\n\tresp, _ := (&http.Client{}).Do(req)\n\tdefer resp.Body.Close()\n\tdata, _ := io.ReadAll(resp.Body)\n\tfmt.Println(string(data))\n}`
  return `import java.net.http.*;\nimport java.net.URI;\n\npublic class Main {\n    public static void main(String[] args) throws Exception {\n        var request = HttpRequest.newBuilder()\n            .uri(URI.create("${origin}/v1/tasks/YOUR_TASK_ID"))\n            .header("Authorization", "Bearer YOUR_API_KEY")\n            .GET()\n            .build();\n\n        var response = HttpClient.newHttpClient()\n            .send(request, HttpResponse.BodyHandlers.ofString());\n        System.out.println(response.body());\n    }\n}`
}

function getChannelResponse(channel: UserChannel) {
  if (channel.type === 'llm') {
    return JSON.stringify({
      id: 'chatcmpl-abc123',
      object: 'chat.completion',
      model: channel.routing_model || channel.name,
      choices: [{
        index: 0,
        message: { role: 'assistant', content: '你好！我是一个人工智能助手，很高兴认识你。请问有什么我可以帮助你的吗？' },
        finish_reason: 'stop',
      }],
      usage: { prompt_tokens: 12, completion_tokens: 34, total_tokens: 46 },
    }, null, 2)
  }
  return JSON.stringify({ task_id: 'task_abc1234xyz', status: 'pending' }, null, 2)
}

export function UserModelsPage() {
  const { data: channels, loading, error, reload } = useAsync(async () => {
    const response = await userApi.listChannels()
    return Array.isArray(response) ? response : response.channels ?? []
  }, [] as UserChannel[])

  const [filterType, setFilterType] = useState('')
  const [filterName, setFilterName] = useState('')
  const [filterProtocol, setFilterProtocol] = useState('')
  const [docVisible, setDocVisible] = useState(false)
  const [docMode, setDocMode] = useState<DocMode>('channel')
  const [docChannel, setDocChannel] = useState<UserChannel | null>(null)
  const [langTab, setLangTab] = useState<LangTab>('curl')
  const [sunoMode, setSunoMode] = useState<SunoMode>('inspire')

  const protocolOptions = useMemo(
    () => Array.from(new Set(channels.map((channel) => channel.protocol || 'openai'))),
    [channels],
  )

  const filteredChannels = useMemo(() => {
    return channels.filter((channel) => {
      if (filterType && channel.type !== filterType) return false
      if (filterProtocol && (channel.protocol || 'openai') !== filterProtocol) return false
      if (!filterName) return true

      const keyword = filterName.toLowerCase()
      return [channel.name, channel.routing_model, channel.description]
        .some((value) => value?.toLowerCase().includes(keyword))
    })
  }, [channels, filterName, filterProtocol, filterType])

  function openDoc(channel: UserChannel) {
    setDocChannel(channel)
    setDocMode('channel')
    setLangTab('curl')
    setSunoMode('inspire')
    setDocVisible(true)
  }

  function openBalanceDocs() {
    setDocMode('balance')
    setLangTab('curl')
    setDocVisible(true)
  }

  function openTaskDocs() {
    setDocMode('task')
    setLangTab('curl')
    setDocVisible(true)
  }

  return (
    <>
      <PageHeader
        eyebrow="Catalog"
        title="模型列表"
        description="查看当前可用模型、协议、价格说明与调用示例。"
        actions={
          <>
            <Button variant="outline" onClick={openBalanceDocs} className="hidden sm:inline-flex">
              <TerminalSquare data-icon="inline-start" />
              查余额 API
            </Button>
            <Button variant="outline" onClick={openTaskDocs} className="hidden sm:inline-flex">
              <TerminalSquare data-icon="inline-start" />
              异步任务 API
            </Button>
            {error ? <Button size="sm" variant="outline" onClick={reload}>重试</Button> : null}
          </>
        }
      />
      {error ? (
        <Alert className="mb-4" variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <div className="mb-6 flex flex-col gap-4">
        <div className="flex flex-wrap items-center gap-2">
          {typeOptions.map((option) => (
            <Badge
              key={option.value}
              variant={filterType === option.value ? 'default' : 'secondary'}
              className="cursor-pointer px-3 py-1"
              onClick={() => setFilterType(option.value)}
            >
              {option.label}
            </Badge>
          ))}
        </div>
        {protocolOptions.length > 1 ? (
          <div className="flex flex-wrap items-center gap-2">
            <Badge
              variant={filterProtocol === '' ? 'default' : 'secondary'}
              className="cursor-pointer px-3 py-1"
              onClick={() => setFilterProtocol('')}
            >
              全部协议
            </Badge>
            {protocolOptions.map((protocol) => (
              <Badge
                key={protocol}
                variant={filterProtocol === protocol ? 'default' : 'secondary'}
                className="cursor-pointer px-3 py-1"
                onClick={() => setFilterProtocol(protocol)}
              >
                {protocolLabels[protocol] ?? protocol}
              </Badge>
            ))}
          </div>
        ) : null}
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <span className="text-sm font-medium text-muted-foreground">共 {filteredChannels.length} 个模型</span>
          <div className="relative">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              className="w-[280px] pl-9"
              placeholder="搜索模型名称、标识或描述"
              value={filterName}
              onChange={(event) => setFilterName(event.target.value)}
            />
          </div>
        </div>
      </div>

      {loading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {Array.from({ length: 8 }).map((_, index) => (
            <Card key={index}>
              <CardContent className="flex flex-col gap-3 p-5">
                <Skeleton className="h-14 w-full" />
                <Skeleton className="h-5 w-2/3" />
                <Skeleton className="h-12 w-full" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : filteredChannels.length === 0 ? (
        <EmptyState
          icon={<BlocksIcon className="size-6 text-muted-foreground" />}
          title="暂无模型数据"
          description="没有找到匹配条件的模型数据。"
        />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {filteredChannels.map((channel, index) => (
            <Card
              key={channel.id ?? index}
              className="group cursor-pointer overflow-hidden transition-colors hover:border-primary/50"
              onClick={() => openDoc(channel)}
            >
              <CardContent className="flex h-full flex-col gap-4 p-5">
                <div className="flex items-start gap-3">
                  <div className="flex size-10 shrink-0 items-center justify-center rounded-lg border bg-muted/30 text-lg font-bold">
                    {channel.icon_url ? (
                      <img alt="" src={channel.icon_url} className="h-full w-full rounded-lg object-cover" />
                    ) : (
                      (channel.name || '?').charAt(0).toUpperCase()
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center justify-between gap-2">
                      <h3 className="truncate text-sm font-semibold" title={channel.name}>{channel.name}</h3>
                      <button
                        className="hidden shrink-0 text-muted-foreground hover:text-foreground group-hover:block"
                        onClick={(event) => {
                          event.stopPropagation()
                          copyText(channel.routing_model || channel.name || '', '已复制模型标识')
                        }}
                      >
                        <Copy className="h-3.5 w-3.5" />
                      </button>
                    </div>
                    <div className="mt-2 flex flex-wrap gap-2 text-[11px]">
                      <Badge variant="outline">{protocolLabels[channel.protocol || 'openai'] ?? (channel.protocol || 'openai')}</Badge>
                      {channel.billing_type ? <Badge variant="outline">{billingTypeLabels[channel.billing_type] ?? channel.billing_type}</Badge> : null}
                    </div>
                    <div className="mt-2 text-xs text-muted-foreground">
                      {channel.price_display ? (
                        channel.price_display.split('\n').map((line, lineIndex) => (
                          <div key={lineIndex} className={lineIndex === 0 ? 'font-medium text-primary/80' : 'text-[10px]'}>{line}</div>
                        ))
                      ) : (
                        <div>按量计费</div>
                      )}
                      {channel.group_price ? <div className="mt-1 font-medium text-emerald-600">专属价：{channel.group_price}</div> : null}
                    </div>
                  </div>
                </div>
                <p className="min-h-10 text-xs leading-5 text-muted-foreground line-clamp-2">
                  {channel.description || '暂无模型描述，可打开详情查看调用示例与响应格式。'}
                </p>
                <div className="mt-auto flex items-center justify-between gap-3 border-t pt-3 text-xs font-medium">
                  <div className="rounded bg-muted/40 px-2 py-1 font-mono text-muted-foreground">
                    {channel.routing_model || channel.model || channel.name}
                  </div>
                  <div className="flex items-center text-emerald-600">
                    <span className="mr-1.5 h-2 w-2 rounded-full bg-emerald-500" />可用
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Sheet open={docVisible} onOpenChange={setDocVisible}>
        <SheetContent side="right" className="flex w-[90vw] flex-col overflow-y-auto p-0 sm:max-w-2xl">
          <SheetHeader className="shrink-0 border-b p-6 pb-2">
            <SheetTitle>
              {docMode === 'balance' ? 'API：查询账户余额' : docMode === 'task' ? 'API：查询任务结果' : docChannel?.name}
            </SheetTitle>
          </SheetHeader>
          <div className="flex-1 space-y-6 bg-muted/10 p-6">
            {docMode === 'balance' ? (
              <div className="space-y-6">
                <div className="flex items-center gap-4">
                  <div className="rounded border border-emerald-200 bg-emerald-100 px-3 py-1 text-sm font-bold tracking-wide text-emerald-800">GET</div>
                  <div className="flex flex-1 items-center rounded border bg-background px-3 py-1 font-mono">
                    <span className="flex-1">/user/balance</span>
                    <button onClick={() => copyText(`GET ${window.location.origin}/user/balance`)} className="hover:text-primary"><Copy className="h-4 w-4" /></button>
                  </div>
                </div>
                <div className="rounded-xl border border-blue-100 bg-blue-50/50 p-4 text-sm leading-relaxed text-blue-900">
                  返回当前 API Key 对应账户的余额。balance_credits 为内部精度值，balance_cny 为等值积分数。
                </div>
                <div>
                  <h4 className="mb-2 flex items-center justify-between font-semibold">
                    请求头
                    <Button variant="ghost" size="sm" className="h-7" onClick={() => copyText('Authorization: Bearer YOUR_API_KEY')}>
                      <Copy data-icon="inline-start" />
                      复制头
                    </Button>
                  </h4>
                  <pre className="overflow-auto rounded-xl bg-zinc-950 p-4 font-mono text-sm text-zinc-50">Authorization: Bearer YOUR_API_KEY</pre>
                </div>
                <div>
                  <h4 className="mb-2 flex items-center justify-between font-semibold">
                    调用示例
                    <Button variant="ghost" size="sm" className="h-7" onClick={() => copyText(getBalanceCode(langTab))}>
                      <Copy data-icon="inline-start" />
                      复制代码
                    </Button>
                  </h4>
                  <Tabs value={langTab} onValueChange={(value) => setLangTab(value as LangTab)}>
                    <TabsList className="mb-2 grid w-full grid-cols-5">
                      <TabsTrigger value="curl">cURL</TabsTrigger>
                      <TabsTrigger value="python">Python</TabsTrigger>
                      <TabsTrigger value="php">PHP</TabsTrigger>
                      <TabsTrigger value="go">Go</TabsTrigger>
                      <TabsTrigger value="java">Java</TabsTrigger>
                    </TabsList>
                  </Tabs>
                  <pre className="min-h-[140px] overflow-auto rounded-xl bg-zinc-950 p-4 font-mono text-sm whitespace-pre-wrap text-zinc-50">{getBalanceCode(langTab)}</pre>
                </div>
                <div>
                  <h4 className="mb-2 font-semibold">响应示例</h4>
                  <pre className="overflow-auto rounded-xl bg-zinc-950 p-4 font-mono text-sm text-green-400">{JSON.stringify({ balance_credits: 1971573, balance_cny: 1.971573 }, null, 2)}</pre>
                </div>
              </div>
            ) : null}

            {docMode === 'task' ? (
              <div className="space-y-6">
                <div className="flex items-center gap-4">
                  <div className="rounded border border-emerald-200 bg-emerald-100 px-3 py-1 text-sm font-bold tracking-wide text-emerald-800">GET</div>
                  <div className="flex flex-1 items-center rounded border bg-background px-3 py-1 font-mono">
                    <span className="flex-1">/v1/tasks/{'{id}'}</span>
                    <button onClick={() => copyText(`${window.location.origin}/v1/tasks/YOUR_TASK_ID`)} className="hover:text-primary"><Copy className="h-4 w-4" /></button>
                  </div>
                </div>
                <div className="rounded-xl border border-blue-100 bg-blue-50/50 p-4 text-sm leading-relaxed text-blue-900">
                  轮询图片、视频、音频、音乐任务执行结果。code=150 进行中，code=200 成功，code=500 失败。
                </div>
                <div>
                  <h4 className="mb-2 flex items-center justify-between font-semibold">
                    调用示例
                    <Button variant="ghost" size="sm" className="h-7" onClick={() => copyText(getTaskCode(langTab))}>
                      <Copy data-icon="inline-start" />
                      复制代码
                    </Button>
                  </h4>
                  <Tabs value={langTab} onValueChange={(value) => setLangTab(value as LangTab)}>
                    <TabsList className="mb-2 grid w-full grid-cols-5">
                      <TabsTrigger value="curl">cURL</TabsTrigger>
                      <TabsTrigger value="python">Python</TabsTrigger>
                      <TabsTrigger value="php">PHP</TabsTrigger>
                      <TabsTrigger value="go">Go</TabsTrigger>
                      <TabsTrigger value="java">Java</TabsTrigger>
                    </TabsList>
                  </Tabs>
                  <pre className="min-h-[140px] overflow-auto rounded-xl bg-zinc-950 p-4 font-mono text-sm whitespace-pre-wrap text-zinc-50">{getTaskCode(langTab)}</pre>
                </div>
                <div>
                  <h4 className="mb-2 font-semibold">响应示例</h4>
                  <pre className="overflow-auto rounded-xl bg-zinc-950 p-4 font-mono text-sm text-green-400">{JSON.stringify({ task_id: '12345', status: 1, code: 200, msg: 'success', url: '', credits_charged: 3600 }, null, 2)}</pre>
                </div>
              </div>
            ) : null}

            {docMode === 'channel' && docChannel ? (
              <div className="space-y-6">
                <div className="flex items-center gap-4">
                  <div className="rounded border border-blue-200 bg-blue-100 px-3 py-1 text-sm font-bold tracking-wide text-blue-800">POST</div>
                  <div className="flex flex-1 items-center rounded border bg-background px-3 py-1 font-mono">
                    <span className="flex-1">{getChannelEndpoint(docChannel)}</span>
                    <button onClick={() => copyText(`${window.location.origin}${getChannelEndpoint(docChannel)}`)} className="hover:text-primary"><Copy className="h-4 w-4" /></button>
                  </div>
                </div>

                <div className="grid gap-3 rounded-xl border bg-background p-4 sm:grid-cols-2">
                  <div>
                    <div className="text-xs text-muted-foreground">模型标识</div>
                    <div className="mt-1 font-mono text-sm font-medium">{docChannel.routing_model || docChannel.name}</div>
                  </div>
                  <div>
                    <div className="text-xs text-muted-foreground">协议</div>
                    <div className="mt-1 text-sm font-medium">{protocolLabels[docChannel.protocol || 'openai'] ?? (docChannel.protocol || 'openai')}</div>
                  </div>
                  <div>
                    <div className="text-xs text-muted-foreground">计费方式</div>
                    <div className="mt-1 text-sm font-medium">{docChannel.billing_type ? (billingTypeLabels[docChannel.billing_type] ?? docChannel.billing_type) : '—'}</div>
                  </div>
                  <div>
                    <div className="text-xs text-muted-foreground">专属价格</div>
                    <div className="mt-1 whitespace-pre-wrap text-sm font-medium text-emerald-600">{docChannel.group_price || '无差异'}</div>
                  </div>
                </div>

                <div className="rounded-xl border bg-accent/50 p-4 text-sm leading-relaxed">
                  <div className="whitespace-pre-wrap font-medium">{docChannel.price_display || '暂无默认价格说明'}</div>
                  {docChannel.description ? <div className="mt-3 whitespace-pre-wrap text-muted-foreground">{docChannel.description}</div> : null}
                </div>

                {docChannel.type === 'music' ? (
                  <Tabs value={sunoMode} onValueChange={(value) => setSunoMode(value as SunoMode)}>
                    <TabsList className="grid h-auto w-full grid-cols-5 py-1">
                      <TabsTrigger value="inspire" className="text-xs">描述模式</TabsTrigger>
                      <TabsTrigger value="custom" className="text-xs">定制模式</TabsTrigger>
                      <TabsTrigger value="extend" className="text-xs">延长模式</TabsTrigger>
                      <TabsTrigger value="overpainting" className="text-xs">重绘主歌</TabsTrigger>
                      <TabsTrigger value="underpainting" className="text-xs">重绘前奏</TabsTrigger>
                    </TabsList>
                  </Tabs>
                ) : null}

                <div>
                  <h4 className="mb-2 flex items-center justify-between font-semibold">
                    调用示例
                    <Button variant="ghost" size="sm" className="h-7" onClick={() => copyText(getChannelCode(docChannel, langTab, sunoMode))}>
                      <Copy data-icon="inline-start" />
                      复制代码
                    </Button>
                  </h4>
                  <Tabs value={langTab} onValueChange={(value) => setLangTab(value as LangTab)}>
                    <TabsList className="mb-2 grid w-full grid-cols-5">
                      <TabsTrigger value="curl">cURL</TabsTrigger>
                      <TabsTrigger value="python">Python</TabsTrigger>
                      <TabsTrigger value="php">PHP</TabsTrigger>
                      <TabsTrigger value="go">Go</TabsTrigger>
                      <TabsTrigger value="java">Java</TabsTrigger>
                    </TabsList>
                  </Tabs>
                  <pre className="min-h-[140px] overflow-auto rounded-xl bg-zinc-950 p-4 font-mono text-sm whitespace-pre-wrap text-zinc-50">{getChannelCode(docChannel, langTab, sunoMode)}</pre>
                </div>

                <div>
                  <h4 className="mb-2 font-semibold">响应示例</h4>
                  <pre className={cn('overflow-auto rounded-xl bg-zinc-950 p-4 font-mono text-sm text-green-400', docChannel.type !== 'llm' && 'opacity-80')}>
                    {getChannelResponse(docChannel)}
                  </pre>
                </div>
              </div>
            ) : null}
          </div>
        </SheetContent>
      </Sheet>
    </>
  )
}
