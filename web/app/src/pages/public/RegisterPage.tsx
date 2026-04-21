import type { FormEvent } from 'react'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { getApiErrorMessage } from '@/lib/api/http'
import { authApi } from '@/lib/api/public'

export function RegisterPage() {
  const navigate = useNavigate()
  const [form, setForm] = useState({
    username: '',
    email: '',
    password: '',
  })
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setError('')

    try {
      await authApi.register(form)
      navigate('/login')
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
          User sign up
        </p>
        <CardTitle className="text-3xl tracking-tight">创建用户账号</CardTitle>
        <p className="text-sm text-muted-foreground">
          第一阶段保持兼容现有后端注册逻辑，界面按新规范重构。
        </p>
      </CardHeader>
      <CardContent>
        <form className="space-y-4" onSubmit={handleSubmit}>
          <div className="space-y-2">
            <label className="text-sm font-medium">用户名</label>
            <Input
              value={form.username}
              onChange={(event) =>
                setForm((current) => ({ ...current, username: event.target.value }))
              }
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">邮箱</label>
            <Input
              type="email"
              value={form.email}
              onChange={(event) =>
                setForm((current) => ({ ...current, email: event.target.value }))
              }
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">密码</label>
            <Input
              type="password"
              value={form.password}
              onChange={(event) =>
                setForm((current) => ({ ...current, password: event.target.value }))
              }
            />
          </div>
          {error ? (
            <div className="rounded-xl border border-destructive/20 bg-destructive/5 px-4 py-3 text-sm text-destructive">
              {error}
            </div>
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
