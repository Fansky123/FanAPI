import { useMemo, useState } from 'react'
import { PlusIcon, SaveIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { adminApi, type AdminChannel, type AdminKeyPool } from '@/lib/api/admin'
import { useAsync } from '@/hooks/use-async'

type ChannelForm = {
  id?: number
  name: string
  model: string
  type: string
  protocol: string
  base_url: string
  method: string
  query_url: string
  query_method: string
  timeout_ms: string
  query_timeout_ms: string
  billing_type: string
  headers_text: string
  // token billing
  billing_input_price: string
  billing_output_price: string
  billing_input_cost: string
  billing_output_cost: string
  billing_cache_read_price: string
  billing_cache_read_cost: string
  billing_input_from_response: boolean
  // image billing
  billing_base_price: string
  billing_default_size_price: string
  // video / audio billing
  billing_price_per_second: string
  // count billing
  billing_price_per_call: string
  billing_script: string
  request_script: string
  response_script: string
  query_script: string
  error_script: string
  key_pool_id: string
  auth_type: string
  auth_param_name: string
  auth_region: string
  auth_service: string
  weight: string
  priority: string
  icon_url: string
  description: string
  is_active: boolean
}

const emptyJson = '{}'

const emptyForm: ChannelForm = {
  name: '',
  model: '',
  type: 'llm',
  protocol: 'openai',
  base_url: '',
  method: 'POST',
  query_url: '',
  query_method: 'GET',
  timeout_ms: '60000',
  query_timeout_ms: '30000',
  billing_type: 'token',
  headers_text: emptyJson,
  billing_input_price: '',
  billing_output_price: '',
  billing_input_cost: '',
  billing_output_cost: '',
  billing_cache_read_price: '',
  billing_cache_read_cost: '',
  billing_input_from_response: false,
  billing_base_price: '',
  billing_default_size_price: '',
  billing_price_per_second: '',
  billing_price_per_call: '',
  billing_script: '',
  request_script: '',
  response_script: '',
  query_script: '',
  error_script: '',
  key_pool_id: '',
  auth_type: 'bearer',
  auth_param_name: '',
  auth_region: '',
  auth_service: '',
  weight: '1',
  priority: '0',
  icon_url: '',
  description: '',
  is_active: true,
}

function prettyJson(value: unknown) {
  if (!value || (typeof value === 'object' && Object.keys(value as object).length === 0)) {
    return emptyJson
  }
  return JSON.stringify(value, null, 2)
}

function parseJsonField(label: string, value: string) {
  try {
    return JSON.parse(value || emptyJson) as Record<string, unknown>
  } catch {
    throw new Error(`${label} 不是合法 JSON`)
  }
}

function getNum(cfg: Record<string, unknown>, key: string): string {
  const v = cfg[key]
  return v !== undefined && v !== null ? String(v) : ''
}

function buildBillingConfig(form: ChannelForm): Record<string, unknown> {
  switch (form.billing_type) {
    case 'token': {
      const cfg: Record<string, unknown> = {}
      if (form.billing_input_price) cfg.input_price_per_1m_tokens = Number(form.billing_input_price)
      if (form.billing_output_price) cfg.output_price_per_1m_tokens = Number(form.billing_output_price)
      if (form.billing_input_cost) cfg.input_cost_per_1m_tokens = Number(form.billing_input_cost)
      if (form.billing_output_cost) cfg.output_cost_per_1m_tokens = Number(form.billing_output_cost)
      if (form.billing_cache_read_price) cfg.cache_read_price_per_1m_tokens = Number(form.billing_cache_read_price)
      if (form.billing_cache_read_cost) cfg.cache_read_cost_per_1m_tokens = Number(form.billing_cache_read_cost)
      if (form.billing_input_from_response) cfg.input_from_response = true
      return cfg
    }
    case 'image': {
      const cfg: Record<string, unknown> = {}
      if (form.billing_base_price) cfg.base_price = Number(form.billing_base_price)
      if (form.billing_default_size_price) cfg.default_size_price = Number(form.billing_default_size_price)
      return cfg
    }
    case 'video':
    case 'audio': {
      const cfg: Record<string, unknown> = {}
      if (form.billing_price_per_second) cfg.price_per_second = Number(form.billing_price_per_second)
      return cfg
    }
    case 'count': {
      const cfg: Record<string, unknown> = {}
      if (form.billing_price_per_call) cfg.price_per_call = Number(form.billing_price_per_call)
      return cfg
    }
    default:
      return {}
  }
}

function formatBilling(channel: AdminChannel) {
  const config = channel.billing_config ?? {}
  switch (channel.billing_type) {
    case 'token':
      return `in ${config.input_price_per_1m_tokens ?? 0} / out ${config.output_price_per_1m_tokens ?? 0}`
    case 'image':
      return String(config.default_size_price ?? config.base_price ?? 0)
    case 'video':
    case 'audio':
      return `${config.price_per_second ?? 0}/s`
    case 'count':
      return `${config.price_per_call ?? 0}/call`
    default:
      return channel.billing_type ?? '-'
  }
}

export function AdminChannelsPage() {
  const { data, loading, error: loadError, reload } = useAsync(async () => {
    const [channelResponse, poolResponse] = await Promise.all([
      adminApi.listChannels(),
      adminApi.listKeyPools(),
    ])
    const rows = Array.isArray(channelResponse)
      ? channelResponse
      : channelResponse.channels ?? channelResponse.items ?? []
    const pools = Array.isArray(poolResponse) ? poolResponse : poolResponse.pools ?? []
    return { rows, pools }
  }, { rows: [] as AdminChannel[], pools: [] as AdminKeyPool[] })

  const rows = data.rows
  const pools = data.pools

  const [mutError, setMutError] = useState('')
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState<ChannelForm>(emptyForm)
  const [pendingDeleteChannel, setPendingDeleteChannel] = useState<AdminChannel | undefined>()

  const error = loadError || mutError

  const poolOptions = useMemo(
    () =>
      pools.filter((pool) =>
        form.id
          ? pool.channel_id === form.id || String(pool.channel_id) === form.key_pool_id
          : pool.channel_id === Number(form.key_pool_id || form.id || 0) || pool.channel_id === 0
      ),
    [form.id, form.key_pool_id, pools]
  )

  function openCreate() {
    setForm(emptyForm)
    setOpen(true)
    setMutError('')
  }

  function openEdit(row: AdminChannel) {
    setForm({
      id: row.id,
      name: row.name ?? '',
      model: row.model ?? row.routing_model ?? '',
      type: row.type ?? 'llm',
      protocol: row.protocol ?? 'openai',
      base_url: row.base_url ?? '',
      method: row.method ?? 'POST',
      query_url: row.query_url ?? '',
      query_method: row.query_method ?? 'GET',
      timeout_ms: String(row.timeout_ms ?? 60000),
      query_timeout_ms: String(row.query_timeout_ms ?? 30000),
      billing_type: row.billing_type ?? 'token',
      headers_text: prettyJson(row.headers),
      billing_input_price: getNum(row.billing_config ?? {}, 'input_price_per_1m_tokens'),
      billing_output_price: getNum(row.billing_config ?? {}, 'output_price_per_1m_tokens'),
      billing_input_cost: getNum(row.billing_config ?? {}, 'input_cost_per_1m_tokens'),
      billing_output_cost: getNum(row.billing_config ?? {}, 'output_cost_per_1m_tokens'),
      billing_cache_read_price: getNum(row.billing_config ?? {}, 'cache_read_price_per_1m_tokens'),
      billing_cache_read_cost: getNum(row.billing_config ?? {}, 'cache_read_cost_per_1m_tokens'),
      billing_input_from_response: Boolean(row.billing_config?.input_from_response),
      billing_base_price: getNum(row.billing_config ?? {}, 'base_price'),
      billing_default_size_price: getNum(row.billing_config ?? {}, 'default_size_price'),
      billing_price_per_second: getNum(row.billing_config ?? {}, 'price_per_second'),
      billing_price_per_call: getNum(row.billing_config ?? {}, 'price_per_call'),
      billing_script: row.billing_script ?? '',
      request_script: row.request_script ?? '',
      response_script: row.response_script ?? '',
      query_script: row.query_script ?? '',
      error_script: row.error_script ?? '',
      key_pool_id: row.key_pool_id ? String(row.key_pool_id) : '',
      auth_type: row.auth_type ?? 'bearer',
      auth_param_name: row.auth_param_name ?? '',
      auth_region: row.auth_region ?? '',
      auth_service: row.auth_service ?? '',
      weight: String(row.weight ?? 1),
      priority: String(row.priority ?? 0),
      icon_url: row.icon_url ?? '',
      description: row.description ?? '',
      is_active: row.is_active ?? true,
    })
    setOpen(true)
    setMutError('')
  }

  async function saveChannel() {
    setMutError('')
    try {
      const payload = {
        name: form.name.trim(),
        model: form.model.trim(),
        type: form.type,
        protocol: form.protocol,
        base_url: form.base_url.trim(),
        method: form.method,
        query_url: form.query_url.trim(),
        query_method: form.query_method,
        timeout_ms: Number(form.timeout_ms || '60000'),
        query_timeout_ms: Number(form.query_timeout_ms || '30000'),
        billing_type: form.billing_type,
        headers: parseJsonField('请求头', form.headers_text),
        billing_config: buildBillingConfig(form),
        billing_script: form.billing_script,
        request_script: form.request_script,
        response_script: form.response_script,
        query_script: form.query_script,
        error_script: form.error_script,
        key_pool_id: Number(form.key_pool_id || '0'),
        auth_type: form.auth_type,
        auth_param_name: form.auth_param_name.trim(),
        auth_region: form.auth_region.trim(),
        auth_service: form.auth_service.trim(),
        weight: Number(form.weight || '1'),
        priority: Number(form.priority || '0'),
        icon_url: form.icon_url.trim(),
        description: form.description.trim(),
        is_active: form.is_active,
      }
      if (form.id) {
        await adminApi.updateChannel(form.id, payload)
      } else {
        await adminApi.createChannel(payload)
      }
      setOpen(false)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  async function toggleChannel(row: AdminChannel) {
    if (!row.id) return
    setMutError('')
    try {
      await adminApi.toggleChannel(row.id, !(row.is_active ?? true))
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  async function executeDeleteChannel() {
    if (!pendingDeleteChannel?.id) return
    setMutError('')
    try {
      await adminApi.deleteChannel(pendingDeleteChannel.id)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    } finally {
      setPendingDeleteChannel(undefined)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Catalog"
        title="渠道管理"
        description="管理 API 渠道，支持认证、计费、脚本、轮询、号池和负载参数。"
        actions={
          <>
            {error ? (
              <Button size="sm" variant="outline" onClick={reload}>
                重试
              </Button>
            ) : null}
            <Button onClick={openCreate}>
              <PlusIcon data-icon="inline-start" />
              新增渠道
            </Button>
          </>
        }
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>名称</TableHead>
              <TableHead>模型</TableHead>
              <TableHead>类型</TableHead>
              <TableHead>协议</TableHead>
              <TableHead>计费</TableHead>
              <TableHead>号池</TableHead>
              <TableHead>优先级/权重</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={9} />
          ) : (
            <TableBody>
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={9} className="py-10 text-center text-muted-foreground">
                    暂无渠道数据
                  </TableCell>
                </TableRow>
              ) : (
                rows.map((row, index) => (
                  <TableRow key={row.id ?? index}>
                    <TableCell className="max-w-56">
                      <div className="font-medium">{row.name ?? '未命名渠道'}</div>
                      {row.description ? (
                        <div className="line-clamp-1 text-xs text-muted-foreground">{row.description}</div>
                      ) : null}
                    </TableCell>
                    <TableCell className="max-w-48 break-all text-xs">{row.model ?? row.routing_model ?? '-'}</TableCell>
                    <TableCell>{row.type ?? '-'}</TableCell>
                    <TableCell>{row.protocol ?? 'openai'}</TableCell>
                    <TableCell className="text-xs">{formatBilling(row)}</TableCell>
                    <TableCell>{row.key_pool_id ? `#${row.key_pool_id}` : '—'}</TableCell>
                    <TableCell className="text-xs">P{row.priority ?? 0} / W{row.weight ?? 1}</TableCell>
                    <TableCell>
                      <Badge variant={row.is_active === false ? 'secondary' : 'default'}>
                        {row.is_active === false ? '停用' : '启用'}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button size="sm" variant="outline" onClick={() => openEdit(row)}>编辑</Button>
                        <Button size="sm" variant="outline" onClick={() => toggleChannel(row)}>
                          {row.is_active === false ? '启用' : '停用'}
                        </Button>
                        <Button size="sm" onClick={() => setPendingDeleteChannel(row)}>删除</Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          )}
        </Table>
      </Card>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-5xl">
          <DialogHeader>
            <DialogTitle>{form.id ? '编辑渠道' : '新增渠道'}</DialogTitle>
            <DialogDescription>覆盖上游接入所需的核心字段。</DialogDescription>
          </DialogHeader>

          <Tabs defaultValue="basic">
            <TabsList className="w-full">
              <TabsTrigger value="basic">基本信息</TabsTrigger>
              <TabsTrigger value="auth">认证 &amp; 号池</TabsTrigger>
              <TabsTrigger value="billing">计费</TabsTrigger>
              <TabsTrigger value="scripts">脚本 &amp; 轮询</TabsTrigger>
            </TabsList>

            {/* ── 基本信息 ── */}
            <TabsContent value="basic" className="mt-5 max-h-[62vh] overflow-y-auto pr-1">
              <div className="grid gap-5 md:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-medium">路由名称</label>
                  <Input value={form.name} onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">标准模型名</label>
                  <Input value={form.model} onChange={(event) => setForm((current) => ({ ...current, model: event.target.value }))} />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">接口类型</label>
                  <NativeSelect value={form.type} onChange={(event) => setForm((current) => ({ ...current, type: event.target.value }))}>
                    <option value="llm">llm</option>
                    <option value="image">image</option>
                    <option value="video">video</option>
                    <option value="audio">audio</option>
                    <option value="music">music</option>
                  </NativeSelect>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">协议</label>
                  <NativeSelect value={form.protocol} onChange={(event) => setForm((current) => ({ ...current, protocol: event.target.value }))}>
                    <option value="openai">openai</option>
                    <option value="claude">claude</option>
                    <option value="gemini">gemini</option>
                  </NativeSelect>
                </div>
                <div className="space-y-2 md:col-span-2">
                  <label className="text-sm font-medium">上游 URL</label>
                  <Input value={form.base_url} onChange={(event) => setForm((current) => ({ ...current, base_url: event.target.value }))} placeholder="https://api.example.com/v1/chat/completions" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">请求方法</label>
                  <NativeSelect value={form.method} onChange={(event) => setForm((current) => ({ ...current, method: event.target.value }))}>
                    <option value="POST">POST</option>
                    <option value="GET">GET</option>
                  </NativeSelect>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">超时（ms）</label>
                  <Input value={form.timeout_ms} onChange={(event) => setForm((current) => ({ ...current, timeout_ms: event.target.value }))} />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">图标 URL</label>
                  <Input value={form.icon_url} onChange={(event) => setForm((current) => ({ ...current, icon_url: event.target.value }))} placeholder="https://…/icon.png" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">描述</label>
                  <Input value={form.description} onChange={(event) => setForm((current) => ({ ...current, description: event.target.value }))} placeholder="可选，显示在渠道名称下方" />
                </div>
                <div className="flex items-center gap-2 md:col-span-2 pt-1">
                  <input
                    id="channel-active"
                    type="checkbox"
                    checked={form.is_active}
                    onChange={(event) => setForm((current) => ({ ...current, is_active: event.target.checked }))}
                    className="h-4 w-4 rounded border-input"
                  />
                  <label htmlFor="channel-active" className="cursor-pointer text-sm font-medium">渠道启用</label>
                </div>
              </div>
            </TabsContent>

            {/* ── 认证 & 号池 ── */}
            <TabsContent value="auth" className="mt-5 max-h-[62vh] overflow-y-auto pr-1">
              <div className="grid gap-5 md:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-medium">认证方式</label>
                  <NativeSelect value={form.auth_type} onChange={(event) => setForm((current) => ({ ...current, auth_type: event.target.value }))}>
                    <option value="bearer">bearer</option>
                    <option value="query_param">query_param</option>
                    <option value="basic">basic</option>
                    <option value="sigv4">sigv4</option>
                  </NativeSelect>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Query Param 名</label>
                  <Input value={form.auth_param_name} onChange={(event) => setForm((current) => ({ ...current, auth_param_name: event.target.value }))} placeholder="如 key（query_param 认证用）" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">AWS Region</label>
                  <Input value={form.auth_region} onChange={(event) => setForm((current) => ({ ...current, auth_region: event.target.value }))} placeholder="us-east-1（sigv4 认证用）" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">AWS Service</label>
                  <Input value={form.auth_service} onChange={(event) => setForm((current) => ({ ...current, auth_service: event.target.value }))} placeholder="execute-api（sigv4 认证用）" />
                </div>

                <div className="border-t pt-4 md:col-span-2" />

                <div className="space-y-2">
                  <label className="text-sm font-medium">号池绑定</label>
                  <NativeSelect value={form.key_pool_id} onChange={(event) => setForm((current) => ({ ...current, key_pool_id: event.target.value }))}>
                    <option value="">不启用</option>
                    {poolOptions.map((pool) => (
                      <option key={pool.id} value={String(pool.id)}>
                        #{pool.id} {pool.name}
                      </option>
                    ))}
                  </NativeSelect>
                </div>
                <div className="space-y-2">{/* placeholder for grid alignment */}</div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">优先级</label>
                  <Input value={form.priority} onChange={(event) => setForm((current) => ({ ...current, priority: event.target.value }))} placeholder="数值越大越优先" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">权重</label>
                  <Input value={form.weight} onChange={(event) => setForm((current) => ({ ...current, weight: event.target.value }))} placeholder="加权随机，越大被选中概率越高" />
                </div>

                <div className="border-t pt-4 md:col-span-2" />

                <div className="space-y-2 md:col-span-2">
                  <label className="text-sm font-medium">请求头（JSON）</label>
                  <p className="text-xs text-muted-foreground">固定注入到每次上游请求的 HTTP 头，如 Authorization。</p>
                  <Textarea
                    value={form.headers_text}
                    onChange={(event) => setForm((current) => ({ ...current, headers_text: event.target.value }))}
                    rows={6}
                    className="font-mono text-xs"
                  />
                </div>
              </div>
            </TabsContent>

            {/* ── 计费 ── */}
            <TabsContent value="billing" className="mt-5 max-h-[62vh] overflow-y-auto pr-1">
              <div className="grid gap-5 md:grid-cols-2">
                <div className="space-y-2 md:col-span-2">
                  <label className="text-sm font-medium">计费类型</label>
                  <NativeSelect value={form.billing_type} onChange={(event) => setForm((current) => ({ ...current, billing_type: event.target.value }))}>
                    <option value="token">token — 按 token 数计费</option>
                    <option value="image">image — 按图片张数计费</option>
                    <option value="video">video — 按视频秒数计费</option>
                    <option value="audio">audio — 按音频秒数计费</option>
                    <option value="count">count — 按调用次数计费</option>
                    <option value="custom">custom — 自定义脚本计费</option>
                  </NativeSelect>
                </div>

                {form.billing_type === 'token' && (
                  <>
                    <div className="space-y-1 md:col-span-2">
                      <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">用户侧价格</p>
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">输入价格（/百万 token）</label>
                      <Input type="number" value={form.billing_input_price} onChange={(e) => setForm((c) => ({ ...c, billing_input_price: e.target.value }))} placeholder="如 740000" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">输出价格（/百万 token）</label>
                      <Input type="number" value={form.billing_output_price} onChange={(e) => setForm((c) => ({ ...c, billing_output_price: e.target.value }))} placeholder="如 5900000" />
                    </div>
                    <div className="space-y-1 md:col-span-2">
                      <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">成本侧价格</p>
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">输入成本（/百万 token）</label>
                      <Input type="number" value={form.billing_input_cost} onChange={(e) => setForm((c) => ({ ...c, billing_input_cost: e.target.value }))} placeholder="如 612000" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">输出成本（/百万 token）</label>
                      <Input type="number" value={form.billing_output_cost} onChange={(e) => setForm((c) => ({ ...c, billing_output_cost: e.target.value }))} placeholder="如 4900000" />
                    </div>
                    <div className="space-y-1 md:col-span-2">
                      <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">缓存（留空按协议默认倍率）</p>
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">缓存读取价格（/百万 token）</label>
                      <Input type="number" value={form.billing_cache_read_price} onChange={(e) => setForm((c) => ({ ...c, billing_cache_read_price: e.target.value }))} placeholder="留空按协议默认" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">缓存读取成本（/百万 token）</label>
                      <Input type="number" value={form.billing_cache_read_cost} onChange={(e) => setForm((c) => ({ ...c, billing_cache_read_cost: e.target.value }))} placeholder="留空按协议默认" />
                    </div>
                    <div className="flex items-center gap-2 md:col-span-2">
                      <input
                        id="input-from-response"
                        type="checkbox"
                        checked={form.billing_input_from_response}
                        onChange={(e) => setForm((c) => ({ ...c, billing_input_from_response: e.target.checked }))}
                        className="h-4 w-4 rounded border-input"
                      />
                      <label htmlFor="input-from-response" className="cursor-pointer text-sm font-medium">
                        从响应中获取实际输入 token 数（input_from_response）
                      </label>
                    </div>
                  </>
                )}

                {form.billing_type === 'image' && (
                  <>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">基础价格（credits）</label>
                      <Input type="number" value={form.billing_base_price} onChange={(e) => setForm((c) => ({ ...c, billing_base_price: e.target.value }))} placeholder="如 5000000" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">默认尺寸价格（credits）</label>
                      <Input type="number" value={form.billing_default_size_price} onChange={(e) => setForm((c) => ({ ...c, billing_default_size_price: e.target.value }))} placeholder="如 5000000" />
                    </div>
                  </>
                )}

                {(form.billing_type === 'video' || form.billing_type === 'audio') && (
                  <div className="space-y-2">
                    <label className="text-sm font-medium">价格（credits / 秒）</label>
                    <Input type="number" value={form.billing_price_per_second} onChange={(e) => setForm((c) => ({ ...c, billing_price_per_second: e.target.value }))} placeholder="如 10000" />
                  </div>
                )}

                {form.billing_type === 'count' && (
                  <div className="space-y-2">
                    <label className="text-sm font-medium">价格（credits / 次）</label>
                    <Input type="number" value={form.billing_price_per_call} onChange={(e) => setForm((c) => ({ ...c, billing_price_per_call: e.target.value }))} placeholder="如 1000" />
                  </div>
                )}

                <div className="space-y-2 md:col-span-2">
                  <label className="text-sm font-medium">自定义计费脚本</label>
                  <p className="text-xs text-muted-foreground">billing_type=custom 时生效，脚本需返回 credits 数值。</p>
                  <Textarea
                    value={form.billing_script}
                    onChange={(event) => setForm((current) => ({ ...current, billing_script: event.target.value }))}
                    rows={8}
                    className="font-mono text-xs"
                    placeholder="function calcBilling(request) { return 1000 }"
                  />
                </div>
              </div>
            </TabsContent>

            {/* ── 脚本 & 轮询 ── */}
            <TabsContent value="scripts" className="mt-5 max-h-[62vh] overflow-y-auto pr-1">
              <div className="grid gap-5">
                <div className="space-y-2">
                  <label className="text-sm font-medium">入参脚本</label>
                  <p className="text-xs text-muted-foreground">mapRequest(input) → 将平台请求映射为上游格式。</p>
                  <Textarea value={form.request_script} onChange={(event) => setForm((current) => ({ ...current, request_script: event.target.value }))} rows={7} className="font-mono text-xs" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">出参脚本</label>
                  <p className="text-xs text-muted-foreground">mapResponse(input) → 映射上游响应，或提取 upstream_task_id（异步）。</p>
                  <Textarea value={form.response_script} onChange={(event) => setForm((current) => ({ ...current, response_script: event.target.value }))} rows={7} className="font-mono text-xs" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">错误检测脚本</label>
                  <p className="text-xs text-muted-foreground">checkError(response) → 返回非空字符串表示错误，null/false 表示正常。</p>
                  <Textarea value={form.error_script} onChange={(event) => setForm((current) => ({ ...current, error_script: event.target.value }))} rows={5} className="font-mono text-xs" />
                </div>

                <div className="border-t pt-2">
                  <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">轮询配置（异步任务用）</p>
                </div>

                <div className="space-y-2">
                  <label className="text-sm font-medium">轮询 URL</label>
                  <Input value={form.query_url} onChange={(event) => setForm((current) => ({ ...current, query_url: event.target.value }))} placeholder="如 https://api.example.com/v1/tasks/{id}" />
                </div>
                <div className="grid grid-cols-2 gap-5">
                  <div className="space-y-2">
                    <label className="text-sm font-medium">轮询方法</label>
                    <NativeSelect value={form.query_method} onChange={(event) => setForm((current) => ({ ...current, query_method: event.target.value }))}>
                      <option value="GET">GET</option>
                      <option value="POST">POST</option>
                    </NativeSelect>
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium">轮询超时（ms）</label>
                    <Input value={form.query_timeout_ms} onChange={(event) => setForm((current) => ({ ...current, query_timeout_ms: event.target.value }))} />
                  </div>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">轮询脚本</label>
                  <p className="text-xs text-muted-foreground">mapResponse(input) → 将轮询响应映射为标准格式。</p>
                  <Textarea value={form.query_script} onChange={(event) => setForm((current) => ({ ...current, query_script: event.target.value }))} rows={7} className="font-mono text-xs" />
                </div>
              </div>
            </TabsContent>
          </Tabs>

          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>取消</Button>
            <Button onClick={saveChannel} disabled={!form.name.trim() || !form.model.trim() || !form.base_url.trim()}>
              <SaveIcon data-icon="inline-start" />
              保存
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={pendingDeleteChannel !== undefined} onOpenChange={() => setPendingDeleteChannel(undefined)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确认删除渠道"{pendingDeleteChannel?.name ?? pendingDeleteChannel?.model ?? pendingDeleteChannel?.id}"吗？此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={executeDeleteChannel}>删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
