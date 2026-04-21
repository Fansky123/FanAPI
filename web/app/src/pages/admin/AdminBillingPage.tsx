import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Card } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { adminApi, type AdminTransaction } from '@/lib/api/admin'
import { getApiErrorMessage } from '@/lib/api/http'
import { formatCredits } from '@/lib/formatters/credits'

export function AdminBillingPage() {
  const [rows, setRows] = useState<AdminTransaction[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const response = await adminApi.listTransactions()
        setRows(
          Array.isArray(response) ? response : response.items ?? response.transactions ?? []
        )
      } catch (err) {
        setError(getApiErrorMessage(err))
      }
    }

    void load()
  }, [])

  return (
    <>
      <PageHeader
        eyebrow="Finance"
        title="后台账单流水"
        description="已接入真实流水列表，后续继续补筛选和详情视图。"
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
              <TableHead>时间</TableHead>
              <TableHead>类型</TableHead>
              <TableHead>金额</TableHead>
              <TableHead>说明</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row, index) => (
              <TableRow key={row.id ?? index}>
                <TableCell>{row.created_at ?? '-'}</TableCell>
                <TableCell>{row.type ?? '-'}</TableCell>
                <TableCell>{formatCredits(row.amount ?? row.credits ?? 0)}</TableCell>
                <TableCell className="text-muted-foreground">
                  {row.remark ?? row.description ?? '-'}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>
    </>
  )
}
