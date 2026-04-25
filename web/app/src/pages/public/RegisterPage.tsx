import type { FormEvent } from 'react'
import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { getApiErrorMessage } from '@/lib/api/http'
import { authApi } from '@/lib/api/public'

export function RegisterPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const inviteCode = searchParams.get('invite') ?? searchParams.get('code') ?? searchParams.get('ref') ?? ''
  const [form, setForm] = useState({
    username: '',
    password: '',
  })
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setError('')

    try {
      await authApi.register({ ...form, ...(inviteCode ? { code: inviteCode } : {}) })
      navigate('/login')
    } catch (err) {
      setError(getApiErrorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Card className="w-full max-w-xl border-border/70 bg-card/92 shadow-lg">
      <CardHeader className="flex flex-col gap-3">
        <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
          User sign up
        </p>
        <CardTitle className="text-3xl tracking-tight">创建用户账号</CardTitle>
        <p className="text-sm text-muted-foreground">
          第一阶段保持兼容现有后端注册逻辑，界面按新规范重构。
        </p>
      </CardHeader>
      <CardContent>
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <div className="flex flex-col gap-2">
            <Label>用户名</Label>
            <Input
              value={form.username}
              onChange={(event) =>
                setForm((current) => ({ ...current, username: event.target.value }))
              }
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label>密码</Label>
            <Input
              type="password"
              value={form.password}
              onChange={(event) =>
                setForm((current) => ({ ...current, password: event.target.value }))
              }
            />
          </div>
          {error ? (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : null}
          <Button className="w-full" type="submit" disabled={submitting}>
            {submitting ? '创建中...' : '创建账号'}
          </Button>
        </form>
        <div className="mt-5 text-sm text-muted-foreground">
          已有账号？{' '}
          <Link className="font-medium text-foreground hover:underline" to="/login">
            去登录
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
