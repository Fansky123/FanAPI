import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { agentApi, type AgentUser } from '@/lib/api/agent'
import { getApiErrorMessage } from '@/lib/api/http'

export function AgentDashboardPage() {
  const [rows, setRows] = useState<AgentUser[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const response = await agentApi.listUsers()
        setRows(Array.isArray(response) ? response : response.items ?? response.users ?? [])
      } catch (err) {
        setError(getApiErrorMessage(err))
      }
    }

    void load()
  }, [])

  return (
    <>
      <PageHeader eyebrow="Agent" title="Agent 工作台" description="Agent 端已经接入基础数据列表，后续继续补充值和邀请能力。" />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <Card>
        <CardHeader>
          <CardTitle>可管理用户</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>ID</TableHead>
                <TableHead>用户名</TableHead>
                <TableHead>邮箱</TableHead>
                <TableHead>余额</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((row, index) => (
                <TableRow key={row.id ?? index}>
                  <TableCell>{row.id ?? '-'}</TableCell>
                  <TableCell>{row.username ?? '-'}</TableCell>
                  <TableCell>{row.email ?? '-'}</TableCell>
                  <TableCell>{row.balance_credits ?? '-'}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </>
  )
}
