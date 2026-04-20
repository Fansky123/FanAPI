import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
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
import { adminApi, type AdminOcpcPlatform } from '@/lib/api/admin'
import { getApiErrorMessage } from '@/lib/api/http'

type PlatformForm = {
  id?: number
  platform: string
  name: string
  enabled: boolean
  baidu_token: string
  baidu_page_url: string
  baidu_reg_type: string
  baidu_order_type: string
  e360_key: string
  e360_secret: string
  e360_jzqs: string
  e360_so_type: string
  e360_reg_event: string
  e360_order_event: string
}

const emptyForm: PlatformForm = {
  platform: 'baidu',
  name: '',
  enabled: true,
  baidu_token: '',
  baidu_page_url: '',
  baidu_reg_type: '68',
  baidu_order_type: '10',
  e360_key: '',
  e360_secret: '',
  e360_jzqs: '',
  e360_so_type: '1',
  e360_reg_event: '',
  e360_order_event: '',
}

export function AdminOcpcPage() {
  const [platforms, setPlatforms] = useState<AdminOcpcPlatform[]>([])
  const [error, setError] = useState('')
  const [form, setForm] = useState<PlatformForm>(emptyForm)
  const [open, setOpen] = useState(false)
  const [scheduleEnabled, setScheduleEnabled] = useState(false)
  const [interval, setInterval] = useState('30')
  const [uploadResult, setUploadResult] = useState('')

  async function load() {
    try {
      const [platformRes, scheduleRes] = await Promise.all([
        adminApi.listOcpcPlatforms(),
        adminApi.getOcpcSchedule(),
      ])
      setPlatforms(Array.isArray(platformRes) ? platformRes : platformRes.list ?? [])
      const schedule = scheduleRes.schedule ?? {}
      setScheduleEnabled(schedule.ocpc_schedule_enabled === 'true')
      setInterval(schedule.ocpc_schedule_interval ?? '30')
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

  function openEdit(row: AdminOcpcPlatform) {
    setForm({
      id: row.id,
      platform: row.platform ?? 'baidu',
      name: row.name ?? '',
      enabled: row.enabled ?? true,
      baidu_token: row.baidu_token ?? '',
      baidu_page_url: row.baidu_page_url ?? '',
      baidu_reg_type: String(row.baidu_reg_type ?? 68),
      baidu_order_type: String(row.baidu_order_type ?? 10),
      e360_key: row.e360_key ?? '',
      e360_secret: row.e360_secret ?? '',
      e360_jzqs: row.e360_jzqs ?? '',
      e360_so_type: row.e360_so_type ?? '1',
      e360_reg_event: row.e360_reg_event ?? '',
      e360_order_event: row.e360_order_event ?? '',
    })
    setOpen(true)
  }

  async function savePlatform() {
    const payload = {
      ...form,
      baidu_reg_type: Number(form.baidu_reg_type),
      baidu_order_type: Number(form.baidu_order_type),
    }
    try {
      if (form.id) {
        await adminApi.updateOcpcPlatform(form.id, payload)
      } else {
        await adminApi.createOcpcPlatform(payload)
      }
      setOpen(false)
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function togglePlatform(row: AdminOcpcPlatform) {
    if (!row.id) return
    try {
      await adminApi.toggleOcpcPlatform(row.id)
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function deletePlatform(row: AdminOcpcPlatform) {
    if (!row.id) return
    if (!window.confirm(`确认删除 OCPC 账户 ${row.name ?? row.id} 吗？`)) return
    try {
      await adminApi.deleteOcpcPlatform(row.id)
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function saveSchedule() {
    try {
      await adminApi.updateOcpcSchedule({ enabled: scheduleEnabled, interval: Number(interval) })
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function triggerUpload() {
    try {
      const result = await adminApi.triggerOcpcUpload()
      setUploadResult(JSON.stringify(result, null, 2))
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="OCPC"
        title="推广账户管理"
        description="支持平台账户配置、手动上报和定时调度。"
        actions={<Button onClick={openCreate}>新增账户</Button>}
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
              <TableHead>ID</TableHead>
              <TableHead>平台</TableHead>
              <TableHead>名称</TableHead>
              <TableHead>关键配置</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {platforms.map((row, index) => (
              <TableRow key={row.id ?? index}>
                <TableCell>{row.id ?? '-'}</TableCell>
                <TableCell>{row.platform ?? '-'}</TableCell>
                <TableCell>{row.name ?? '-'}</TableCell>
                <TableCell>{row.baidu_page_url ?? row.e360_key ?? '-'}</TableCell>
                <TableCell>
                  <Badge variant={row.enabled ? 'default' : 'secondary'}>
                    {row.enabled ? '启用' : '停用'}
                  </Badge>
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-2">
                    <Button size="sm" variant="outline" onClick={() => openEdit(row)}>
                      编辑
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => togglePlatform(row)}>
                      {row.enabled ? '停用' : '启用'}
                    </Button>
                    <Button size="sm" onClick={() => deletePlatform(row)}>删除</Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>
      <Card>
        <CardContent className="flex flex-wrap items-center gap-4 p-6">
          <label className="text-sm font-medium">自动上报</label>
          <input type="checkbox" checked={scheduleEnabled} onChange={(event) => setScheduleEnabled(event.target.checked)} />
          <Input className="w-32" value={interval} onChange={(event) => setInterval(event.target.value)} placeholder="间隔分钟" />
          <Button onClick={saveSchedule}>保存调度</Button>
          <Button variant="outline" onClick={triggerUpload}>立即上报</Button>
        </CardContent>
        {uploadResult ? (
          <CardContent>
            <pre className="overflow-auto rounded-xl border border-border/70 bg-muted/25 p-4 text-xs">
              {uploadResult}
            </pre>
          </CardContent>
        ) : null}
      </Card>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader><DialogTitle>{form.id ? '编辑账户' : '新增账户'}</DialogTitle></DialogHeader>
          <div className="grid gap-4">
            <select className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm outline-none" value={form.platform} onChange={(event) => setForm((current) => ({ ...current, platform: event.target.value }))}>
              <option value="baidu">百度</option>
              <option value="360">360</option>
            </select>
            <Input value={form.name} onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} placeholder="名称" />
            {form.platform === 'baidu' ? (
              <>
                <Input value={form.baidu_token} onChange={(event) => setForm((current) => ({ ...current, baidu_token: event.target.value }))} placeholder="百度 Token" />
                <Input value={form.baidu_page_url} onChange={(event) => setForm((current) => ({ ...current, baidu_page_url: event.target.value }))} placeholder="落地页 URL" />
              </>
            ) : (
              <>
                <Input value={form.e360_key} onChange={(event) => setForm((current) => ({ ...current, e360_key: event.target.value }))} placeholder="360 Key" />
                <Input value={form.e360_secret} onChange={(event) => setForm((current) => ({ ...current, e360_secret: event.target.value }))} placeholder="360 Secret" />
              </>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>取消</Button>
            <Button onClick={savePlatform}>保存</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
