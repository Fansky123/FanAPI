import { useEffect, useMemo, useState } from 'react'
import { PlusIcon, SaveIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
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
import { Card, CardContent } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { NativeSelect } from '@/components/ui/select'
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
import { getApiErrorMessage } from '@/lib/api/http'

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
  billing_config_text: string
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
  billing_config_text: emptyJson,
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
  const [rows, setRows] = useState<AdminChannel[]>([])
  const [pools, setPools] = useState<AdminKeyPool[]>([])
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState<ChannelForm>(emptyForm)
  const [pendingDeleteChannel, setPendingDeleteChannel] = useState<AdminChannel | undefined>()

  async function load() {
    try {
      setError('')
      const [channelResponse, poolResponse] = await Promise.all([
        adminApi.listChannels(),
        adminApi.listKeyPools(),
      ])
      setRows(Array.isArray(channelResponse) ? channelResponse : channelResponse.channels ?? channelResponse.items ?? [])
      setPools(Array.isArray(poolResponse) ? poolResponse : poolResponse.pools ?? [])
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load()
  }, [])

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
      billing_config_text: prettyJson(row.billing_config),
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
  }

  async function saveChannel() {
    try {
      setError('')
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
        billing_config: parseJsonField('计费配置', form.billing_config_text),
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
        setSuccess('渠道已更新')
      } else {
        await adminApi.createChannel(payload)
        setSuccess('渠道已创建')
      }
      setOpen(false)
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function toggleChannel(row: AdminChannel) {
    if (!row.id) return
    try {
      setError('')
      await adminApi.toggleChannel(row.id, !(row.is_active ?? true))
      setSuccess(`渠道已${row.is_active === false ? '启用' : '停用'}`)
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function deleteChannel(row: AdminChannel) {
    if (!row.id) return
    setPendingDeleteChannel(row)
  }

  async function executeDeleteChannel() {
    if (!pendingDeleteChannel?.id) return
    try {
      setError('')
      await adminApi.deleteChannel(pendingDeleteChannel.id)
      setSuccess('渠道已删除')
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    } finally {
      setPendingDeleteChannel(undefined)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Catalog"
        title="渠道管理"
        description="已补到可维护真实渠道配置，支持认证、计费、脚本、轮询、号池和负载参数。"
        actions={
          <Button onClick={openCreate}>
            <PlusIcon data-icon="inline-start" />
            新增渠道
          </Button>
        }
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      {success ? (
        <Card className="border-emerald-500/20 bg-emerald-500/5">
          <CardContent className="py-4 text-sm text-emerald-700">{success}</CardContent>
        </Card>
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
          <TableBody>
            {rows.map((row, index) => (
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
                <TableCell className="text-xs">
                  P{row.priority ?? 0} / W{row.weight ?? 1}
                </TableCell>
                <TableCell>
                  <Badge variant={row.is_active === false ? 'secondary' : 'default'}>
                    {row.is_active === false ? '停用' : '启用'}
                  </Badge>
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-2">
                    <Button size="sm" variant="outline" onClick={() => openEdit(row)}>
                      编辑
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => toggleChannel(row)}>
                      {row.is_active === false ? '启用' : '停用'}
                    </Button>
                    <Button size="sm" onClick={() => deleteChannel(row)}>
                      删除
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-5xl">
          <DialogHeader>
            <DialogTitle>{form.id ? '编辑渠道' : '新增渠道'}</DialogTitle>
            <DialogDescription>这套表单覆盖真实上游接入所需的核心字段。</DialogDescription>
          </DialogHeader>
          <div className="grid max-h-[75vh] gap-4 overflow-y-auto pr-2 md:grid-cols-2">
            <div className="flex flex-col gap-2">
              <Label>路由名称</Label>
              <Input value={form.name} onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} />
            </div>
            <div className="flex flex-col gap-2">
              <Label>标准模型名</Label>
              <Input value={form.model} onChange={(event) => setForm((current) => ({ ...current, model: event.target.value }))} />
            </div>
            <div className="flex flex-col gap-2">
              <Label>接口类型</Label>
              <NativeSelect value={form.type} onChange={(event) => setForm((current) => ({ ...current, type: event.target.value }))}>
                <option value="llm">llm</option>
                <option value="image">image</option>
                <option value="video">video</option>
                <option value="audio">audio</option>
                <option value="music">music</option>
              </NativeSelect>
            </div>
            <div className="flex flex-col gap-2">
              <Label>协议</Label>
              <NativeSelect value={form.protocol} onChange={(event) => setForm((current) => ({ ...current, protocol: event.target.value }))}>
                <option value="openai">openai</option>
                <option value="claude">claude</option>
                <option value="gemini">gemini</option>
              </NativeSelect>
            </div>
            <div className="flex flex-col gap-2 md:col-span-2">
              <Label>上游 URL</Label>
              <Input value={form.base_url} onChange={(event) => setForm((current) => ({ ...current, base_url: event.target.value }))} />
            </div>
            <div className="flex flex-col gap-2">
              <Label>请求方法</Label>
              <NativeSelect value={form.method} onChange={(event) => setForm((current) => ({ ...current, method: event.target.value }))}>
                <option value="POST">POST</option>
                <option value="GET">GET</option>
              </NativeSelect>
            </div>
            <div className="flex flex-col gap-2">
              <Label>超时（ms）</Label>
              <Input value={form.timeout_ms} onChange={(event) => setForm((current) => ({ ...current, timeout_ms: event.target.value }))} />
            </div>
            <div className="flex flex-col gap-2">
              <Label>认证方式</Label>
              <NativeSelect value={form.auth_type} onChange={(event) => setForm((current) => ({ ...current, auth_type: event.target.value }))}>
                <option value="bearer">bearer</option>
                <option value="query_param">query_param</option>
                <option value="basic">basic</option>
                <option value="sigv4">sigv4</option>
              </NativeSelect>
            </div>
            <div className="flex flex-col gap-2">
              <Label>计费类型</Label>
              <NativeSelect value={form.billing_type} onChange={(event) => setForm((current) => ({ ...current, billing_type: event.target.value }))}>
                <option value="token">token</option>
                <option value="image">image</option>
                <option value="video">video</option>
                <option value="audio">audio</option>
                <option value="count">count</option>
                <option value="custom">custom</option>
              </NativeSelect>
            </div>
            <div className="flex flex-col gap-2">
              <Label>Query Param 名</Label>
              <Input value={form.auth_param_name} onChange={(event) => setForm((current) => ({ ...current, auth_param_name: event.target.value }))} placeholder="如 key" />
            </div>
            <div className="flex flex-col gap-2">
              <Label>AWS Region</Label>
              <Input value={form.auth_region} onChange={(event) => setForm((current) => ({ ...current, auth_region: event.target.value }))} placeholder="us-east-1" />
            </div>
            <div className="flex flex-col gap-2">
              <Label>AWS Service</Label>
              <Input value={form.auth_service} onChange={(event) => setForm((current) => ({ ...current, auth_service: event.target.value }))} placeholder="execute-api" />
            </div>
            <div className="flex flex-col gap-2">
              <Label>号池绑定</Label>
              <NativeSelect value={form.key_pool_id} onChange={(event) => setForm((current) => ({ ...current, key_pool_id: event.target.value }))}>
                <option value="">不启用</option>
                {poolOptions.map((pool) => (
                  <option key={pool.id} value={String(pool.id)}>
                    #{pool.id} {pool.name}
                  </option>
                ))}
              </NativeSelect>
            </div>
            <div className="flex flex-col gap-2">
              <Label>权重</Label>
              <Input value={form.weight} onChange={(event) => setForm((current) => ({ ...current, weight: event.target.value }))} />
            </div>
            <div className="flex flex-col gap-2">
              <Label>优先级</Label>
              <Input value={form.priority} onChange={(event) => setForm((current) => ({ ...current, priority: event.target.value }))} />
            </div>
            <div className="flex flex-col gap-2">
              <Label>图标 URL</Label>
              <Input value={form.icon_url} onChange={(event) => setForm((current) => ({ ...current, icon_url: event.target.value }))} />
            </div>
            <div className="flex flex-col gap-2 md:col-span-2">
              <Label>描述</Label>
              <Textarea value={form.description} onChange={(event) => setForm((current) => ({ ...current, description: event.target.value }))} rows={2} />
            </div>
            <div className="flex flex-col gap-2 md:col-span-2">
              <Label>请求头（JSON）</Label>
              <Textarea value={form.headers_text} onChange={(event) => setForm((current) => ({ ...current, headers_text: event.target.value }))} rows={4} className="font-mono text-xs" />
            </div>
            <div className="flex flex-col gap-2 md:col-span-2">
              <Label>计费配置（JSON）</Label>
              <Textarea value={form.billing_config_text} onChange={(event) => setForm((current) => ({ ...current, billing_config_text: event.target.value }))} rows={6} className="font-mono text-xs" />
            </div>
            <div className="flex flex-col gap-2 md:col-span-2">
              <Label>入参脚本</Label>
              <Textarea value={form.request_script} onChange={(event) => setForm((current) => ({ ...current, request_script: event.target.value }))} rows={6} className="font-mono text-xs" />
            </div>
            <div className="flex flex-col gap-2 md:col-span-2">
              <Label>出参脚本</Label>
              <Textarea value={form.response_script} onChange={(event) => setForm((current) => ({ ...current, response_script: event.target.value }))} rows={6} className="font-mono text-xs" />
            </div>
            <div className="flex flex-col gap-2 md:col-span-2">
              <Label>自定义计费脚本</Label>
              <Textarea value={form.billing_script} onChange={(event) => setForm((current) => ({ ...current, billing_script: event.target.value }))} rows={5} className="font-mono text-xs" />
            </div>
            <div className="flex flex-col gap-2 md:col-span-2">
              <Label>轮询 URL</Label>
              <Input value={form.query_url} onChange={(event) => setForm((current) => ({ ...current, query_url: event.target.value }))} placeholder="异步任务用，支持 {id}" />
            </div>
            <div className="flex flex-col gap-2">
              <Label>轮询方法</Label>
              <NativeSelect value={form.query_method} onChange={(event) => setForm((current) => ({ ...current, query_method: event.target.value }))}>
                <option value="GET">GET</option>
                <option value="POST">POST</option>
              </NativeSelect>
            </div>
            <div className="flex flex-col gap-2">
              <Label>轮询超时（ms）</Label>
              <Input value={form.query_timeout_ms} onChange={(event) => setForm((current) => ({ ...current, query_timeout_ms: event.target.value }))} />
            </div>
            <div className="flex flex-col gap-2 md:col-span-2">
              <Label>轮询脚本</Label>
              <Textarea value={form.query_script} onChange={(event) => setForm((current) => ({ ...current, query_script: event.target.value }))} rows={5} className="font-mono text-xs" />
            </div>
            <div className="flex flex-col gap-2 md:col-span-2">
              <Label>错误检测脚本</Label>
              <Textarea value={form.error_script} onChange={(event) => setForm((current) => ({ ...current, error_script: event.target.value }))} rows={5} className="font-mono text-xs" />
            </div>
            <Label className="flex items-center gap-2 text-sm md:col-span-2">
              <input type="checkbox" checked={form.is_active} onChange={(event) => setForm((current) => ({ ...current, is_active: event.target.checked }))} />
              渠道启用
            </Label>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>
              取消
            </Button>
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
