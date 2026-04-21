import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { getApiErrorMessage } from '@/lib/api/http'
import { userApi, type RedeemRecord } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'

export function UserExchangePage() {
  const [code, setCode] = useState('')
  const [history, setHistory] = useState<RedeemRecord[]>([])
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  async function load() {
    try {
      const response = await userApi.getRedeemHistory()
      setHistory(Array.isArray(response) ? response : response.records ?? response.list ?? [])
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load()
  }, [])

  async function redeem() {
    if (!code.trim()) return
    setSubmitting(true)
    try {
      await userApi.redeemCard(code.trim())
      setCode('')
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Redeem"
        title="兑换中心"
        description="支持输入卡密兑换，并查看最近兑换记录。"
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <Card>
        <CardContent className="flex gap-3 p-6">
          <Input
            value={code}
            onChange={(event) => setCode(event.target.value)}
            placeholder="请输入兑换码"
          />
          <Button onClick={redeem} disabled={submitting}>
            {submitting ? '兑换中...' : '立即兑换'}
          </Button>
        </CardContent>
      </Card>
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>兑换码</TableHead>
              <TableHead>积分数量</TableHead>
              <TableHead>兑换时间</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {history.map((row, index) => (
              <TableRow key={row.code ?? index}>
                <TableCell className="font-mono text-xs">{row.code ?? '-'}</TableCell>
                <TableCell>{formatCredits(row.credits ?? row.amount ?? 0)}</TableCell>
                <TableCell>{row.created_at ?? row.redeemed_at ?? '-'}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>
    </>
  )
}
