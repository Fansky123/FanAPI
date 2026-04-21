import { useState } from 'react'
import { SaveIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
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
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { adminApi, type AdminVendor } from '@/lib/api/admin'
import { useAsync } from '@/hooks/use-async'

export function AdminVendorsPage() {
  const { data: rows, loading, error: loadError, reload } = useAsync(async () => {
    const response = await adminApi.listVendors()
    return Array.isArray(response) ? response : response.items ?? response.vendors ?? []
  }, [] as AdminVendor[])

  const [mutError, setMutError] = useState('')
  const [editing, setEditing] = useState<AdminVendor | null>(null)
  const [commission, setCommission] = useState('')

  const error = loadError || mutError

  async function toggleActive(row: AdminVendor) {
    if (!row.id) return
    setMutError('')
    try {
      await adminApi.updateVendor(row.id, {
        is_active: !(row.is_active ?? row.enabled ?? true),
      })
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  function openEdit(row: AdminVendor) {
    setEditing(row)
    setCommission(
      row.commission_ratio !== undefined && row.commission_ratio !== null
        ? String(row.commission_ratio)
        : ''
    )
    setMutError('')
  }

  async function saveVendor() {
    if (!editing?.id) return
    setMutError('')
    try {
      await adminApi.updateVendor(editing.id, {
        commission_ratio: commission === '' ? undefined : Number(commission),
      })
      setEditing(null)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Supply"
        title="号商管理"
        description="管理平台号商账号，支持启停和手续费比例编辑。"
        actions={
          error ? (
            <Button size="sm" variant="outline" onClick={reload}>
              重试
            </Button>
          ) : null
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
              <TableHead>名称</TableHead>
              <TableHead>邮箱</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>佣金/费率</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={6} />
          ) : (
            <TableBody>
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="py-10 text-center text-muted-foreground">
                    暂无号商数据
                  </TableCell>
                </TableRow>
              ) : (
                rows.map((row, index) => (
                  <TableRow key={row.id ?? index}>
                    <TableCell>{row.id ?? '-'}</TableCell>
                    <TableCell>{row.username ?? row.name ?? '-'}</TableCell>
                    <TableCell>{row.email ?? '-'}</TableCell>
                    <TableCell>
                      <Badge variant={(row.is_active ?? row.enabled ?? true) ? 'default' : 'secondary'}>
                        {(row.is_active ?? row.enabled ?? true) ? '启用' : '停用'}
                      </Badge>
                    </TableCell>
                    <TableCell>{row.commission_ratio ?? row.fee_ratio ?? '-'}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button size="sm" variant="outline" onClick={() => openEdit(row)}>
                          编辑比例
                        </Button>
                        <Button size="sm" onClick={() => toggleActive(row)}>
                          {(row.is_active ?? row.enabled ?? true) ? '禁用' : '启用'}
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          )}
        </Table>
      </Card>

      <Dialog open={Boolean(editing)} onOpenChange={() => setEditing(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>编辑号商</DialogTitle>
            <DialogDescription>
              当前号商：{editing?.username ?? editing?.email ?? '-'}
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-2">
            <Label>手续费比例</Label>
            <Input
              value={commission}
              onChange={(event) => setCommission(event.target.value)}
              placeholder="例如：0.15"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditing(null)}>
              取消
            </Button>
            <Button onClick={saveVendor}>
              <SaveIcon data-icon="inline-start" />
              保存
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
