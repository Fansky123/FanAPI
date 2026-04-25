import { useState } from 'react'

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
import { Card, CardContent } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
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
import { adminApi, type AdminOcpcPlatform } from '@/lib/api/admin'
import { useAsync } from '@/hooks/use-async'

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
  const { data, loading, error: loadError, reload } = useAsync(async () => {
    const [platformRes, scheduleRes] = await Promise.all([
      adminApi.listOcpcPlatforms(),
      adminApi.getOcpcSchedule(),
    ])
    const platforms = Array.isArray(platformRes) ? platformRes : platformRes.list ?? []
    const schedule = scheduleRes.schedule ?? {}
    return {
      platforms,
      scheduleEnabled: schedule.ocpc_schedule_enabled === 'true',
      scheduleInterval: schedule.ocpc_schedule_interval ?? '30',
    }
  }, { platforms: [] as AdminOcpcPlatform[], scheduleEnabled: false, scheduleInterval: '30' })

  const [mutError, setMutError] = useState('')
  const [form, setForm] = useState<PlatformForm>(emptyForm)
  const [open, setOpen] = useState(false)
  const [scheduleEnabled, setScheduleEnabled] = useState<boolean | null>(null)
  const [interval, setInterval] = useState<string | null>(null)
  const [uploadResult, setUploadResult] = useState('')
  const [pendingDeletePlatform, setPendingDeletePlatform] = useState<AdminOcpcPlatform | undefined>()

  // Use loaded values unless user has changed them
  const effectiveScheduleEnabled = scheduleEnabled ?? data.scheduleEnabled
  const effectiveInterval = interval ?? data.scheduleInterval

  const error = loadError || mutError

  function openCreate() {
    setForm(emptyForm)
    setOpen(true)
    setMutError('')
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
    setMutError('')
  }

  async function savePlatform() {
    const payload = {
      ...form,
      baidu_reg_type: Number(form.baidu_reg_type),
      baidu_order_type: Number(form.baidu_order_type),
    }
    setMutError('')
    try {
      if (form.id) {
        await adminApi.updateOcpcPlatform(form.id, payload)
      } else {
        await adminApi.createOcpcPlatform(payload)
      }
      setOpen(false)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  async function togglePlatform(row: AdminOcpcPlatform) {
    if (!row.id) return
    setMutError('')
    try {
      await adminApi.toggleOcpcPlatform(row.id)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  async function executeDeletePlatform() {
    if (!pendingDeletePlatform?.id) return
    setMutError('')
    try {
      await adminApi.deleteOcpcPlatform(pendingDeletePlatform.id)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    } finally {
      setPendingDeletePlatform(undefined)
    }
  }

  async function saveSchedule() {
    setMutError('')
    try {
      await adminApi.updateOcpcSchedule({ enabled: effectiveScheduleEnabled, interval: Number(effectiveInterval) })
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  async function triggerUpload() {
    setMutError('')
    try {
      const result = await adminApi.triggerOcpcUpload()
      setUploadResult(JSON.stringify(result, null, 2))
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="OCPC"
        title="推广账户管理"
        description="平台账户配置、手动上报和定时调度。"
        actions={
          <>
            {error ? (
              <Button size="sm" variant="outline" onClick={reload}>
                重试
              </Button>
            ) : null}
            <Button onClick={openCreate}>新增账户</Button>
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
              <TableHead>ID</TableHead>
              <TableHead>平台</TableHead>
              <TableHead>名称</TableHead>
              <TableHead>关键配置</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={6} />
          ) : (
            <TableBody>
              {data.platforms.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="py-10 text-center text-muted-foreground">
                    暂无推广账户
                  </TableCell>
                </TableRow>
              ) : (
                data.platforms.map((row, index) => (
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
                        <Button size="sm" variant="outline" onClick={() => openEdit(row)}>编辑</Button>
                        <Button size="sm" variant="outline" onClick={() => togglePlatform(row)}>
                          {row.enabled ? '停用' : '启用'}
                        </Button>
                        <Button size="sm" onClick={() => setPendingDeletePlatform(row)}>删除</Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          )}
        </Table>
      </Card>
      <Card>
        <CardContent className="flex flex-wrap items-center gap-4 p-6">
          <Label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={effectiveScheduleEnabled}
              onChange={(event) => setScheduleEnabled(event.target.checked)}
            />
            自动上报
          </Label>
          <div className="flex items-center gap-2">
            <Label className="text-sm text-muted-foreground">调度间隔</Label>
            <NativeSelect
              className="w-40"
              value={effectiveInterval}
              onChange={(event) => setInterval(event.target.value)}
            >
              <option value="10">每 10 分钟</option>
              <option value="15">每 15 分钟</option>
              <option value="30">每 30 分钟</option>
              <option value="60">每 1 小时</option>
              <option value="120">每 2 小时</option>
              <option value="360">每 6 小时</option>
            </NativeSelect>
          </div>
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
        <DialogContent className="max-w-lg">
          <DialogHeader><DialogTitle>{form.id ? '编辑账户' : '新增账户'}</DialogTitle></DialogHeader>
          <div className="grid gap-4">
            <div className="grid gap-1.5">
              <Label>平台</Label>
              <NativeSelect
                value={form.platform}
                onChange={(event) => setForm((current) => ({ ...current, platform: event.target.value }))}
              >
                <option value="baidu">百度</option>
                <option value="360">360</option>
              </NativeSelect>
            </div>
            <div className="grid gap-1.5">
              <Label>名称</Label>
              <Input
                value={form.name}
                onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))}
                placeholder="账户备注名"
              />
            </div>
            {form.platform === 'baidu' ? (
              <>
                <div className="grid gap-1.5">
                  <Label>百度 Token</Label>
                  <Input
                    value={form.baidu_token}
                    onChange={(event) => setForm((current) => ({ ...current, baidu_token: event.target.value }))}
                    placeholder="百度统计 / 转化 API Token"
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label>落地页 URL</Label>
                  <Input
                    value={form.baidu_page_url}
                    onChange={(event) => setForm((current) => ({ ...current, baidu_page_url: event.target.value }))}
                    placeholder="https://example.com/landing"
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label>注册转化类型</Label>
                  <NativeSelect
                    value={form.baidu_reg_type}
                    onChange={(event) => setForm((current) => ({ ...current, baidu_reg_type: event.target.value }))}
                  >
                    <option value="68">68 - 注册</option>
                    <option value="25">25 - 表单提交</option>
                    <option value="3">3 - 咨询</option>
                  </NativeSelect>
                </div>
                <div className="grid gap-1.5">
                  <Label>订单转化类型</Label>
                  <NativeSelect
                    value={form.baidu_order_type}
                    onChange={(event) => setForm((current) => ({ ...current, baidu_order_type: event.target.value }))}
                  >
                    <option value="10">10 - 付费</option>
                    <option value="5">5 - 购买</option>
                    <option value="6">6 - 下单</option>
                  </NativeSelect>
                </div>
              </>
            ) : (
              <>
                <div className="grid gap-1.5">
                  <Label>360 Key</Label>
                  <Input
                    value={form.e360_key}
                    onChange={(event) => setForm((current) => ({ ...current, e360_key: event.target.value }))}
                    placeholder="360 推广 API Key"
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label>360 Secret</Label>
                  <Input
                    value={form.e360_secret}
                    onChange={(event) => setForm((current) => ({ ...current, e360_secret: event.target.value }))}
                    placeholder="360 推广 API Secret"
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label>jzqs（站点 ID）</Label>
                  <Input
                    value={form.e360_jzqs}
                    onChange={(event) => setForm((current) => ({ ...current, e360_jzqs: event.target.value }))}
                    placeholder="360 推广站点 ID"
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label>搜索类型</Label>
                  <NativeSelect
                    value={form.e360_so_type}
                    onChange={(event) => setForm((current) => ({ ...current, e360_so_type: event.target.value }))}
                  >
                    <option value="1">PC 搜索</option>
                    <option value="2">移动搜索</option>
                  </NativeSelect>
                </div>
                <div className="grid gap-1.5">
                  <Label>注册事件 ID</Label>
                  <Input
                    value={form.e360_reg_event}
                    onChange={(event) => setForm((current) => ({ ...current, e360_reg_event: event.target.value }))}
                    placeholder="360 注册转化事件 ID"
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label>订单事件 ID</Label>
                  <Input
                    value={form.e360_order_event}
                    onChange={(event) => setForm((current) => ({ ...current, e360_order_event: event.target.value }))}
                    placeholder="360 订单转化事件 ID"
                  />
                </div>
              </>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>取消</Button>
            <Button onClick={savePlatform}>保存</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={pendingDeletePlatform !== undefined}
        onOpenChange={() => setPendingDeletePlatform(undefined)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确认删除 OCPC 账户 {pendingDeletePlatform?.name ?? pendingDeletePlatform?.id} 吗？此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={executeDeletePlatform}>删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
