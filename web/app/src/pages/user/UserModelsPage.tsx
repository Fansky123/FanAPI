import { useState, useMemo } from 'react'
import { BlocksIcon, Search, Copy, TerminalSquare } from 'lucide-react'

import { EmptyState } from '@/components/shared/EmptyState'
import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Input } from '@/components/ui/input'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { userApi, type UserChannel } from '@/lib/api/user'
import { useAsync } from '@/hooks/use-async'
import { cn } from '@/lib/utils'
import { toast } from 'sonner'

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

export function UserModelsPage() {
  const { data: channels, loading, error, reload } = useAsync(async () => {
    const response = await userApi.listChannels()
    return Array.isArray(response) ? response : response.channels ?? []
  }, [] as UserChannel[])

  const [filterType, setFilterType] = useState('')
  const [filterName, setFilterName] = useState('')

  // Drawer state
  const [docVisible, setDocVisible] = useState(false)
  const [docMode, setDocMode] = useState<DocMode>('channel')
  const [docChannel, setDocChannel] = useState<UserChannel | null>(null)
  const [langTab, setLangTab] = useState<LangTab>('curl')
  const [sunoMode, setSunoMode] = useState<SunoMode>('inspire')

  const filteredChannels = useMemo(() => {
    return channels.filter(ch => {
      if (filterType && ch.type !== filterType) return false
      if (filterName && !ch.name?.toLowerCase().includes(filterName.toLowerCase()) && !ch.routing_model?.toLowerCase().includes(filterName.toLowerCase())) return false
      return true
    })
  }, [channels, filterType, filterName])

  const copyText = (text: string, label = '已复制') => {
    navigator.clipboard.writeText(text).then(() => {
      toast.success(label)
    })
  }

  const openDoc = (ch: UserChannel) => {
    setDocChannel(ch)
    setDocMode('channel')
    setLangTab('curl')
    setSunoMode('inspire')
    setDocVisible(true)
  }

  const openBalanceDocs = () => {
    setDocMode('balance')
    setLangTab('curl')
    setDocVisible(true)
  }

  const openTaskDocs = () => {
    setDocMode('task')
    setLangTab('curl')
    setDocVisible(true)
  }

  // ------------ API DOC GENERATORS ------------
  const endpointMap: Record<string, string> = {
    llm: '/v1/chat/completions',
    image: '/v1/image',
    video: '/v1/video',
    audio: '/v1/audio',
    music: '/v1/music',
  }

  const docEndpoint = (ch: UserChannel) => {
    if (ch.type === 'llm' && ch.protocol === 'gemini') {
      const model = ch.routing_model || ch.name
      return `/v1beta/models/${model}:generateContent`
    }
    return endpointMap[ch.type ?? 'llm'] || '/v1/chat/completions'
  }

  const docRequestBody = (ch: UserChannel) => {
    const model = ch.routing_model || ch.name
    if (ch.type === 'llm') {
      if (ch.protocol === 'gemini') {
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
    if (ch.type === 'image') {
      return JSON.stringify({ model, prompt: '一只可爱的橘猫坐在阳光下', size: '1k', aspect_ratio: '1:1', n: 1 }, null, 2)
    }
    if (ch.type === 'video') {
      return JSON.stringify({ model, prompt: '海浪拍打岸边，夕阳西下', size: '720p', aspect_ratio: '16:9', duration: '5' }, null, 2)
    }
    if (ch.type === 'audio') {
      return JSON.stringify({ model, input: '你好，欢迎使用语音合成服务', voice: 'alloy' }, null, 2)
    }
    if (ch.type === 'music') {
      if (sunoMode === 'custom') {
        return JSON.stringify({
          model,
          input_type: '20',
          prompt: '[主歌]\n周四的阳光晒脸庞\n微风轻轻吹过窗\n\n[副歌]\n周四快乐不 散场\n欢声笑语满心房',
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
      // other modes logic
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

  const docCode = (ch: UserChannel, lang: string) => {
    const origin = window.location.origin
    const endpoint = docEndpoint(ch)
    const body = docRequestBody(ch)
    if (lang === 'curl') {
      return `curl -X POST "${origin}${endpoint}" \\\n  -H "Content-Type: application/json" \\\n  -H "Authorization: Bearer YOUR_API_KEY" \\\n  -d '${body}'`
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
    if (lang === 'java') {
      const escapedBody = body.replace(/\\/g, '\\\\').replace(/"/g, '\\"').replace(/\n/g, '\\n')
      return `import java.net.http.*;\nimport java.net.URI;\n\npublic class Main {\n    public static void main(String[] args) throws Exception {\n        String body = "${escapedBody}";\n\n        var request = HttpRequest.newBuilder()\n            .uri(URI.create("${origin}${endpoint}"))\n            .header("Authorization", "Bearer YOUR_API_KEY")\n            .header("Content-Type", "application/json")\n            .POST(HttpRequest.BodyPublishers.ofString(body))\n            .build();\n\n        var response = HttpClient.newHttpClient()\n            .send(request, HttpResponse.BodyHandlers.ofString());\n        System.out.println(response.body());\n    }\n}`
    }
    return ''
  }

  const docResponse = (ch: UserChannel) => {
    if (ch.type === 'llm') {
      return JSON.stringify({
        id: 'chatcmpl-abc123',
        object: 'chat.completion',
        model: ch.routing_model || ch.name,
        choices: [{
          index: 0,
          message: { role: 'assistant', content: '你好！我是一个人工智能助手，很高兴认识你。请问有什么我可以帮助你的吗？' },
          finish_reason: 'stop',
        }],
        usage: { prompt_tokens: 12, completion_tokens: 34, total_tokens: 46 },
      }, null, 2)
    }
    return JSON.stringify({
      task_id: 'task_abc1234xyz',
      status: 'pending',
    }, null, 2)
  }

  const balanceCode = (lang: string) => {
    const origin = window.location.origin
    if (lang === 'curl') return `curl -X GET "${origin}/user/balance" \\\n  -H "Authorization: Bearer YOUR_API_KEY"`
    if (lang === 'python') return `import requests\n\nurl = "${origin}/user/balance"\nheaders = {"Authorization": "Bearer YOUR_API_KEY"}\n\nresponse = requests.get(url, headers=headers)\nprint(response.json())`
    if (lang === 'php') return `<?php\n$url = "${origin}/user/balance";\n\n$ch = curl_init($url);\ncurl_setopt_array($ch, [\n    CURLOPT_RETURNTRANSFER => true,\n    CURLOPT_HTTPHEADER     => ['Authorization: Bearer YOUR_API_KEY'],\n]);\n\necho curl_exec($ch);\ncurl_close($ch);`
    if (lang === 'go') return `package main\n\nimport (\n\t"fmt"\n\t"io"\n\t"net/http"\n)\n\nfunc main() {\n\treq, _ := http.NewRequest("GET", "${origin}/user/balance", nil)\n\treq.Header.Set("Authorization", "Bearer YOUR_API_KEY")\n\n\tresp, _ := (&http.Client{}).Do(req)\n\tdefer resp.Body.Close()\n\tdata, _ := io.ReadAll(resp.Body)\n\tfmt.Println(string(data))\n}`
    if (lang === 'java') return `import java.net.http.*;\nimport java.net.URI;\n\npublic class Main {\n    public static void main(String[] args) throws Exception {\n        var request = HttpRequest.newBuilder()\n            .uri(URI.create("${origin}/user/balance"))\n            .header("Authorization", "Bearer YOUR_API_KEY")\n            .GET()\n            .build();\n\n        var response = HttpClient.newHttpClient()\n            .send(request, HttpResponse.BodyHandlers.ofString());\n        System.out.println(response.body());\n    }\n}`
    return ''
  }
  const balanceResponse = () => JSON.stringify({ balance_credits: 1971573, balance_cny: 1.971573 }, null, 2)

  const taskCode = (lang: string) => {
    const origin = window.location.origin
    if (lang === 'curl') return `curl -X GET "${origin}/v1/tasks/YOUR_TASK_ID" \\\n  -H "Authorization: Bearer YOUR_API_KEY"`
    if (lang === 'python') return `import requests\n\nurl = "${origin}/v1/tasks/YOUR_TASK_ID"\nheaders = {"Authorization": "Bearer YOUR_API_KEY"}\n\nresponse = requests.get(url, headers=headers)\nprint(response.json())`
    if (lang === 'php') return `<?php\n$url = "${origin}/v1/tasks/YOUR_TASK_ID";\n\n$ch = curl_init($url);\ncurl_setopt_array($ch, [\n    CURLOPT_RETURNTRANSFER => true,\n    CURLOPT_HTTPHEADER     => ['Authorization: Bearer YOUR_API_KEY'],\n]);\n\necho curl_exec($ch);\ncurl_close($ch);`
    if (lang === 'go') return `package main\n\nimport (\n\t"fmt"\n\t"io"\n\t"net/http"\n)\n\nfunc main() {\n\treq, _ := http.NewRequest("GET", "${origin}/v1/tasks/YOUR_TASK_ID", nil)\n\treq.Header.Set("Authorization", "Bearer YOUR_API_KEY")\n\n\tresp, _ := (&http.Client{}).Do(req)\n\tdefer resp.Body.Close()\n\tdata, _ := io.ReadAll(resp.Body)\n\tfmt.Println(string(data))\n}`
    if (lang === 'java') return `import java.net.http.*;\nimport java.net.URI;\n\npublic class Main {\n    public static void main(String[] args) throws Exception {\n        var request = HttpRequest.newBuilder()\n            .uri(URI.create("${origin}/v1/tasks/YOUR_TASK_ID"))\n            .header("Authorization", "Bearer YOUR_API_KEY")\n            .GET()\n            .build();\n\n        var response = HttpClient.newHttpClient()\n            .send(request, HttpResponse.BodyHandlers.ofString());\n        System.out.println(response.body());\n    }\n}`
    return ''
  }
  const taskResponse = () => JSON.stringify({ task_id: '12345', status: 1, code: 200, msg: 'success', url: '', credits_charged: 3600 }, null, 2)
  // ------------ END DOC GEN ------------

  return (
    <>
      <PageHeader
        eyebrow="Catalog"
        title="模型列表"
        description="查看当前可用的 AI 模型渠道，了解参数及价格说明并获取调用示例代码。"
        actions={
          <>
            <Button variant="outline" onClick={openBalanceDocs} className="hidden sm:inline-flex">
              <TerminalSquare className="mr-2 h-4 w-4" /> 查余额API
            </Button>
            <Button variant="outline" onClick={openTaskDocs} className="hidden sm:inline-flex">
              <TerminalSquare className="mr-2 h-4 w-4" /> 异步任务API
            </Button>
            {error && <Button size="sm" variant="outline" onClick={reload}>重试</Button>}
          </>
        }
      />
      {error && (
        <Alert variant="destructive" className="mb-4">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* 搜索与过滤 */}
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-wrap items-center gap-2">
          {typeOptions.map(t => (
            <Badge 
              key={t.value} 
              variant={filterType === t.value ? 'default' : 'secondary'}
              className="cursor-pointer px-3 py-1"
              onClick={() => setFilterType(t.value)}
            >
              {t.label}
            </Badge>
          ))}
        </div>
        <div className="flex items-center gap-3">
          <span className="text-sm font-medium text-muted-foreground whitespace-nowrap">共 {filteredChannels.length} 模型</span>
          <div className="relative">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input 
              placeholder="搜索模型名称或标识" 
              className="w-[240px] pl-9" 
              value={filterName} 
              onChange={e => setFilterName(e.target.value)} 
            />
          </div>
        </div>
      </div>

      {loading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {Array.from({ length: 8 }).map((_, index) => (
            <Card key={index}><CardContent className="p-5 flex flex-col gap-3"><Skeleton className="h-12 w-full"/><Skeleton className="h-4 w-2/3"/></CardContent></Card>
          ))}
        </div>
      ) : filteredChannels.length === 0 ? (
        <EmptyState
          icon={<BlocksIcon className="size-6 text-muted-foreground" />}
          title="暂无模型数据"
          description="没有找到匹配条件的模型数据"
        />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {filteredChannels.map((channel, index) => (
            <Card key={channel.id ?? index} className="overflow-hidden hover:border-primary/50 transition-colors cursor-pointer group" onClick={() => openDoc(channel)}>
              <CardContent className="p-5 flex flex-col h-full gap-4">
                <div className="flex items-start gap-3">
                  <div className="h-10 w-10 shrink-0 rounded-lg bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center font-bold text-lg border">
                    {channel.icon_url ? (
                      <img src={channel.icon_url} alt="" className="h-full w-full rounded-lg object-cover" />
                    ) : (
                      (channel.name || '?').charAt(0).toUpperCase()
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-2 overflow-hidden">
                      <h3 className="font-semibold text-sm truncate" title={channel.name}>{channel.name}</h3>
                      <button onClick={(e) => { e.stopPropagation(); copyText(channel.routing_model || channel.name || '', '已复制模型标识') }} className="text-muted-foreground hover:text-foreground shrink-0 focus:outline-none hidden group-hover:block">
                        <Copy className="h-3.5 w-3.5" />
                      </button>
                    </div>
                    {channel.price_display ? (
                      <div className="text-xs text-muted-foreground mt-1">
                        {channel.price_display.split('\n').map((line, i) => (
                          <div key={i} className={i === 0 ? 'text-primary/80 font-medium' : 'text-[10px]'}>{line}</div>
                        ))}
                      </div>
                    ) : (
                      <div className="text-xs text-muted-foreground mt-1">按量计费</div>
                    )}
                  </div>
                </div>
                <div className="text-xs text-muted-foreground line-clamp-2 mt-auto pt-2 border-t font-mono bg-muted/30 px-2 py-1 rounded w-fit">
                  {channel.routing_model || channel.model || channel.name}
                </div>
                <div className="flex items-center justify-between text-xs font-medium">
                  <Badge variant="outline" className={cn(
                    channel.type === 'llm' && 'text-blue-600 border-blue-200 bg-blue-50 dark:bg-blue-900/20',
                    channel.type === 'image' && 'text-purple-600 border-purple-200 bg-purple-50 dark:bg-purple-900/20',
                    channel.type === 'audio' && 'text-green-600 border-green-200 bg-green-50 dark:bg-green-900/20',
                  )}>
                    {channel.type?.toUpperCase() || 'LLM'}
                  </Badge>
                  <div className="flex items-center text-green-600">
                    <span className="h-2 w-2 rounded-full bg-green-500 mr-1.5" />可用
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* 文档 Drawer */}
      <Sheet open={docVisible} onOpenChange={setDocVisible}>
        <SheetContent side="right" className="sm:max-w-2xl w-[90vw] overflow-y-auto p-0 flex flex-col">
          <SheetHeader className="p-6 pb-2 border-b shrink-0">
            <SheetTitle>
              {docMode === 'balance' ? 'API：查询账户余额' : docMode === 'task' ? 'API：查询任务结果' : docChannel?.name}
            </SheetTitle>
          </SheetHeader>
          <div className="p-6 bg-muted/10 flex-1">
            {docMode === 'balance' && (
              <div className="space-y-6">
                <div className="flex items-center gap-4">
                  <div className="flex items-center bg-green-100 text-green-800 px-3 py-1 rounded border border-green-200 text-sm font-bold tracking-wide mt-1">GET</div>
                  <div className="flex items-center font-mono bg-background border px-3 py-1 rounded flex-1">
                    <span className="flex-1">/user/balance</span>
                    <button onClick={() => copyText(`GET ${window.location.origin}/user/balance`)} className="hover:text-primary"><Copy className="h-4 w-4"/></button>
                  </div>
                </div>
                <div className="bg-blue-50/50 text-blue-900 border border-blue-100 p-4 rounded-xl text-sm leading-relaxed">
                  返回当前 API Key 对应账户的余额。<br/>
                  <code className="bg-background px-1 border rounded mx-1 text-xs">balance_credits</code> 为内部精度值（÷ 1,000,000 = 积分数）；<br/>
                  <code className="bg-background px-1 border rounded mx-1 text-xs">balance_cny</code> 为等值积分数。
                </div>
                <div>
                  <h4 className="font-semibold mb-2 flex items-center justify-between">请求体 Request <Button variant="ghost" size="sm" className="h-7" onClick={() => copyText('Authorization: Bearer YOUR_API_KEY')}><Copy className="w-3 h-3 mr-1"/> 复制头</Button></h4>
                  <pre className="bg-zinc-950 text-zinc-50 p-4 rounded-xl text-sm font-mono overflow-auto">Authorization: Bearer YOUR_API_KEY</pre>
                </div>
                <div>
                  <h4 className="font-semibold mb-2 flex items-center justify-between">调用示例 Code <Button variant="ghost" size="sm" className="h-7" onClick={() => copyText(balanceCode(langTab))}><Copy className="w-3 h-3 mr-1"/> 复制代码</Button></h4>
                  <Tabs value={langTab} onValueChange={(v) => setLangTab(v as LangTab)}>
                    <TabsList className="mb-2 grid w-full grid-cols-5">
                      <TabsTrigger value="curl">cURL</TabsTrigger>
                      <TabsTrigger value="python">Python</TabsTrigger>
                      <TabsTrigger value="php">PHP</TabsTrigger>
                      <TabsTrigger value="go">Go</TabsTrigger>
                      <TabsTrigger value="java">Java</TabsTrigger>
                    </TabsList>
                    <div className="relative">
                      <pre className="bg-zinc-950 text-zinc-50 p-4 rounded-xl text-sm font-mono overflow-auto min-h-[140px] whitespace-pre-wrap">{balanceCode(langTab)}</pre>
                    </div>
                  </Tabs>
                </div>
                <div>
                  <h4 className="font-semibold mb-2">响应 Response</h4>
                  <pre className="bg-zinc-950 text-green-400 p-4 rounded-xl text-sm font-mono overflow-auto">{balanceResponse()}</pre>
                </div>
              </div>
            )}
            
            {docMode === 'task' && (
              <div className="space-y-6">
                <div className="flex items-center gap-4">
                  <div className="flex items-center bg-green-100 text-green-800 px-3 py-1 rounded border border-green-200 text-sm font-bold tracking-wide mt-1">GET</div>
                  <div className="flex items-center font-mono bg-background border px-3 py-1 rounded flex-1">
                    <span className="flex-1">/v1/tasks/{"{id}"}</span>
                    <button onClick={() => copyText(`${window.location.origin}/v1/tasks/YOUR_TASK_ID`)} className="hover:text-primary"><Copy className="h-4 w-4"/></button>
                  </div>
                </div>
                <div className="bg-blue-50/50 text-blue-900 border border-blue-100 p-4 rounded-xl text-sm leading-relaxed">
                  轮询图片 / 视频 / 音频 / 音乐任务的执行结果。<br/>
                  <code className="bg-background px-1 border rounded mx-1 text-xs">code=150</code> 进行中，
                  <code className="bg-background px-1 border rounded mx-1 text-xs">code=200</code> 成功，
                  <code className="bg-background px-1 border rounded mx-1 text-xs">code=500</code> 失败。<br/>
                  建议间隔 2～5 秒轮询，成功后读取 <code className="bg-background px-1 border rounded mx-1 text-xs">url</code>（单结果）或 <code className="bg-background px-1 border rounded mx-1 text-xs">items</code>（多结果）。
                </div>
                <div>
                  <h4 className="font-semibold mb-2 flex items-center justify-between">调用示例 Code <Button variant="ghost" size="sm" className="h-7" onClick={() => copyText(taskCode(langTab))}><Copy className="w-3 h-3 mr-1"/> 复制代码</Button></h4>
                  <Tabs value={langTab} onValueChange={(v) => setLangTab(v as LangTab)}>
                    <TabsList className="mb-2 grid w-full grid-cols-5">
                      <TabsTrigger value="curl">cURL</TabsTrigger>
                      <TabsTrigger value="python">Python</TabsTrigger>
                      <TabsTrigger value="php">PHP</TabsTrigger>
                      <TabsTrigger value="go">Go</TabsTrigger>
                      <TabsTrigger value="java">Java</TabsTrigger>
                    </TabsList>
                    <div className="relative">
                      <pre className="bg-zinc-950 text-zinc-50 p-4 rounded-xl text-sm font-mono overflow-auto min-h-[140px] whitespace-pre-wrap">{taskCode(langTab)}</pre>
                    </div>
                  </Tabs>
                </div>
                <div>
                  <h4 className="font-semibold mb-2">响应 Response</h4>
                  <pre className="bg-zinc-950 text-green-400 p-4 rounded-xl text-sm font-mono overflow-auto">{taskResponse()}</pre>
                </div>
              </div>
            )}

            {docMode === 'channel' && docChannel && (
              <div className="space-y-6">
                <div className="flex items-center gap-4">
                  <div className="flex items-center bg-blue-100 text-blue-800 px-3 py-1 rounded border border-blue-200 text-sm font-bold tracking-wide mt-1">POST</div>
                  <div className="flex items-center font-mono bg-background border px-3 py-1 rounded flex-1">
                    <span className="flex-1">{docEndpoint(docChannel)}</span>
                    <button onClick={() => copyText(`${window.location.origin}${docEndpoint(docChannel)}`)} className="hover:text-primary"><Copy className="h-4 w-4"/></button>
                  </div>
                </div>

                <div className="bg-accent/50 border p-4 rounded-xl text-sm leading-relaxed">
                  <ul>
                    <li>模型标识：<code className="bg-background px-1 border rounded ml-1 font-mono text-xs font-semibold">{docChannel.routing_model || docChannel.name}</code></li>
                    {docChannel.description && <li className="mt-2 text-muted-foreground whitespace-pre-wrap">{docChannel.description}</li>}
                  </ul>
                </div>

                {docChannel.type === 'music' && (
                  <Tabs value={sunoMode} onValueChange={(v) => setSunoMode(v as SunoMode)}>
                    <TabsList className="grid grid-cols-5 h-auto py-1 w-full">
                      <TabsTrigger value="inspire" className="text-xs">描述模式</TabsTrigger>
                      <TabsTrigger value="custom" className="text-xs">定制模式</TabsTrigger>
                      <TabsTrigger value="extend" className="text-xs">延长模式</TabsTrigger>
                      <TabsTrigger value="overpainting" className="text-xs">重绘主歌</TabsTrigger>
                      <TabsTrigger value="underpainting" className="text-xs">重绘前奏</TabsTrigger>
                    </TabsList>
                  </Tabs>
                )}

                <div>
                  <h4 className="font-semibold mb-2 flex items-center justify-between">调用示例 Code <Button variant="ghost" size="sm" className="h-7" onClick={() => copyText(docCode(docChannel, langTab))}><Copy className="w-3 h-3 mr-1"/> 复制代码</Button></h4>
                  <Tabs value={langTab} onValueChange={(v) => setLangTab(v as LangTab)}>
                    <TabsList className="mb-2 grid w-full grid-cols-5">
                      <TabsTrigger value="curl">cURL</TabsTrigger>
                      <TabsTrigger value="python">Python</TabsTrigger>
                      <TabsTrigger value="php">PHP</TabsTrigger>
                      <TabsTrigger value="go">Go</TabsTrigger>
                      <TabsTrigger value="java">Java</TabsTrigger>
                    </TabsList>
                    <div className="relative">
                      <pre className="bg-zinc-950 text-zinc-50 p-4 rounded-xl text-sm font-mono overflow-auto min-h-[140px] whitespace-pre-wrap">{docCode(docChannel, langTab)}</pre>
                    </div>
                  </Tabs>
                </div>
                <div>
                  <h4 className="font-semibold mb-2">响应示例</h4>
                  <pre className={cn("bg-zinc-950 text-green-400 p-4 rounded-xl text-sm font-mono overflow-auto", docChannel.type !== 'llm' && "opacity-80")}>{docResponse(docChannel)}</pre>
                </div>
              </div>
            )}
          </div>
        </SheetContent>
      </Sheet>
    </>
  )
}
