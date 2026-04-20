import type { FormEvent } from 'react'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { adminAuthApi } from '@/lib/api/admin'
import { getApiErrorMessage } from '@/lib/api/http'
import { setRoleToken, setSiteModePreference } from '@/lib/auth/storage'

export function AdminLoginPage() {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setError('')

    try {
      const response = await adminAuthApi.login({ username, password })
      setRoleToken('admin', response.token)
      setSiteModePreference('admin')
      navigate('/admin/dashboard')
    } catch (err) {
      setError(getApiErrorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Card className="w-full max-w-xl border-border/70 bg-card/92 shadow-lg">
      <CardHeader className="space-y-3">
        <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
          Admin sign in
        </p>
        <h1 className="text-3xl font-semibold tracking-tight">登录管理后台</h1>
        <p className="text-sm text-muted-foreground">
          管理后台优先保证高密度、可扫描和稳定的操作体验。
        </p>
      </CardHeader>
      <CardContent>
        <form className="space-y-4" onSubmit={handleSubmit}>
          <div className="space-y-2">
            <label className="text-sm font-medium">用户名</label>
            <Input value={username} onChange={(event) => setUsername(event.target.value)} />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">密码</label>
            <Input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />
          </div>
          {error ? (
            <div className="rounded-xl border border-destructive/20 bg-destructive/5 px-4 py-3 text-sm text-destructive">
              {error}
            </div>
          ) : null}
          <Button className="w-full" type="submit" disabled={submitting}>
            {submitting ? '登录中...' : '进入后台'}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
