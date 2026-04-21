import { useEffect, useState } from 'react'
import { SaveIcon } from 'lucide-react'

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
import { adminApi, type AdminVendor } from '@/lib/api/admin'
import { getApiErrorMessage } from '@/lib/api/http'

export function AdminVendorsPage() {
  const [rows, setRows] = useState<AdminVendor[]>([])
  const [error, setError] = useState('')
  const [editing, setEditing] = useState<AdminVendor | null>(null)
  const [commission, setCommission] = useState('')

  async function load() {
    try {
      const response = await adminApi.listVendors()
      setRows(Array.isArray(response) ? response : response.items ?? response.vendors ?? [])
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load()
  }, [])

  async function toggleActive(row: AdminVendor) {
    if (!row.id) return
    try {
      await adminApi.updateVendor(row.id, {
        is_active: !(row.is_active ?? row.enabled ?? true),
      })
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  function openEdit(row: AdminVendor) {
    setEditing(row)
    setCommission(
      row.commission_ratio !== undefined && row.commission_ratio !== null
        ? String(row.commission_ratio)
        : ''
    )
  }

  async function saveVendor() {
    if (!editing?.id) return
    try {
      await adminApi.updateVendor(editing.id, {
        commission_ratio: commission === '' ? undefined : Number(commission),
      })
      setEditing(null)
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Supply"
        title="号商管理"
        description="号商页已支持启停和手续费比例编辑，满足最小运营要求。"
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
              <TableHead>名称</TableHead>
              <TableHead>邮箱</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>佣金/费率</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row, index) => (
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
            ))}
          </TableBody>
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
          <div className="space-y-2">
            <label className="text-sm font-medium">手续费比例</label>
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
