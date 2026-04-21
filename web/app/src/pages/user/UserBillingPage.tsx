import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { getApiErrorMessage } from '@/lib/api/http'
import { userApi, type UserTransaction } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'

export function UserBillingPage() {
  const [rows, setRows] = useState<UserTransaction[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const response = await userApi.getTransactions()
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
        title="账单与流水"
        description="账单页先建立统一数据表格密度、列宽和状态表达，后续再补完整筛选与详情交互。"
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
              <TableHead>时间</TableHead>
              <TableHead>类型</TableHead>
              <TableHead>金额</TableHead>
              <TableHead>说明</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row, index) => (
              <TableRow key={row.id ?? index}>
                <TableCell>{row.created_at ?? row.time ?? '-'}</TableCell>
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
