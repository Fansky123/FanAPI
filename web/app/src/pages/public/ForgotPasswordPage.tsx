import type { FormEvent } from 'react'
import { useState } from 'react'
import { Link } from 'react-router-dom'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { getApiErrorMessage } from '@/lib/api/http'
import { authApi } from '@/lib/api/public'

export function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setMessage('')
    setError('')

    try {
      await authApi.forgotPassword(email)
      setMessage('如果邮箱已绑定账号，重置指引将发送到该邮箱。')
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
          Password recovery
        </p>
        <CardTitle className="text-3xl tracking-tight">找回密码</CardTitle>
      </CardHeader>
      <CardContent>
        <form className="space-y-4" onSubmit={handleSubmit}>
          <div className="space-y-2">
            <label className="text-sm font-medium">邮箱</label>
            <Input
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              placeholder="请输入账号绑定邮箱"
            />
          </div>
          {message ? (
            <div className="rounded-xl border border-primary/20 bg-primary/5 px-4 py-3 text-sm text-foreground">
              {message}
            </div>
          ) : null}
          {error ? (
            <div className="rounded-xl border border-destructive/20 bg-destructive/5 px-4 py-3 text-sm text-destructive">
              {error}
            </div>
          ) : null}
          <Button className="w-full" type="submit" disabled={submitting}>
            {submitting ? '提交中...' : '发送重置指引'}
          </Button>
        </form>
        <div className="mt-5 text-sm text-muted-foreground">
          <Link className="hover:text-foreground" to="/login">
            返回登录
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
