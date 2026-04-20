import { useEffect, useState } from 'react'
import { PlusIcon, SaveIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { adminApi, type AdminChannel } from '@/lib/api/admin'
import { getApiErrorMessage } from '@/lib/api/http'

type ChannelForm = {
  id?: number
  name: string
  model: string
  type: string
  protocol: string
  base_url: string
  method: string
  timeout_ms: string
  billing_type: string
  is_active: boolean
}

const emptyForm: ChannelForm = {
  name: '',
  model: '',
  type: 'llm',
  protocol: 'openai',
  base_url: '',
  method: 'POST',
  timeout_ms: '60000',
  billing_type: 'token',
  is_active: true,
}

export function AdminChannelsPage() {
  const [rows, setRows] = useState<AdminChannel[]>([])
  const [error, setError] = useState('')
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState<ChannelForm>(emptyForm)

  async function load() {
    try {
      const response = await adminApi.listChannels()
      setRows(Array.isArray(response) ? response : response.channels ?? response.items ?? [])
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load()
  }, [])

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
      timeout_ms: String(row.timeout_ms ?? 60000),
      billing_type: row.billing_type ?? 'token',
      is_active: row.is_active ?? true,
    })
    setOpen(true)
  }

  async function saveChannel() {
    try {
      const payload = {
        name: form.name,
        model: form.model,
        type: form.type,
        protocol: form.protocol,
        base_url: form.base_url,
        method: form.method,
        timeout_ms: Number(form.timeout_ms),
        billing_type: form.billing_type,
        headers: {},
        billing_config: {},
        auth_type: 'bearer',
        is_active: form.is_active,
        weight: 1,
        priority: 0,
      }
      if (form.id) {
        await adminApi.updateChannel(form.id, payload)
      } else {
        await adminApi.createChannel(payload)
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
      await adminApi.toggleChannel(row.id, !(row.is_active ?? true))
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function deleteChannel(row: AdminChannel) {
    if (!row.id) return
    if (!window.confirm(`确认删除渠道“${row.name ?? row.model ?? row.id}”吗？`)) return
    try {
      await adminApi.deleteChannel(row.id)
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Catalog"
        title="渠道管理"
        description="已补上最小管理闭环，支持新增、编辑、启停和删除渠道。"
        actions={
          <Button onClick={openCreate}>
            <PlusIcon data-icon="inline-start" />
            新增渠道
          </Button>
        }
      />
      {error ? (
        <Card className="border-destructive/25 bg-destructive/5">
          <CardContent className="py-4 text-sm text-destructive">{error}</CardContent>
        </Card>
      ) : null}
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>名称</TableHead>
              <TableHead>模型</TableHead>
              <TableHead>类型</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row, index) => (
              <TableRow key={row.id ?? index}>
                <TableCell className="font-medium">{row.name ?? '未命名渠道'}</TableCell>
                <TableCell>{row.model ?? row.routing_model ?? '-'}</TableCell>
                <TableCell>{row.type ?? '-'}</TableCell>
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
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{form.id ? '编辑渠道' : '新增渠道'}</DialogTitle>
            <DialogDescription>先使用最小字段集合打通后台操作闭环。</DialogDescription>
          </DialogHeader>
          <div className="grid gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">名称</label>
              <Input value={form.name} onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">模型</label>
              <Input value={form.model} onChange={(event) => setForm((current) => ({ ...current, model: event.target.value }))} />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">类型</label>
                <select className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm outline-none" value={form.type} onChange={(event) => setForm((current) => ({ ...current, type: event.target.value }))}>
                  <option value="llm">llm</option>
                  <option value="image">image</option>
                  <option value="video">video</option>
                  <option value="audio">audio</option>
                </select>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">协议</label>
                <select className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm outline-none" value={form.protocol} onChange={(event) => setForm((current) => ({ ...current, protocol: event.target.value }))}>
                  <option value="openai">openai</option>
                  <option value="claude">claude</option>
                  <option value="gemini">gemini</option>
                </select>
              </div>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">上游 URL</label>
              <Input value={form.base_url} onChange={(event) => setForm((current) => ({ ...current, base_url: event.target.value }))} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>
              取消
            </Button>
            <Button onClick={saveChannel}>
              <SaveIcon data-icon="inline-start" />
              保存
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
