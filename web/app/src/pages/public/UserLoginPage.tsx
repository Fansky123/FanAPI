import type { FormEvent } from 'react'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { getApiErrorMessage } from '@/lib/api/http'
import { authApi } from '@/lib/api/public'
import { setRoleToken, setSiteModePreference } from '@/lib/auth/storage'

export function UserLoginPage() {
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
      const response = await authApi.login({ username, password })
      setRoleToken('user', response.token)
      setSiteModePreference('user')
      navigate('/dashboard')
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
          User sign in
        </p>
        <h1 className="text-3xl font-semibold tracking-tight">登录用户端</h1>
        <p className="text-sm text-muted-foreground">
          使用现有 FanAPI 账号进入新版本用户控制台。
        </p>
      </CardHeader>
      <CardContent>
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <div className="flex flex-col gap-2">
            <Label>用户名 / 邮箱</Label>
            <Input
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              placeholder="请输入用户名或邮箱"
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label>密码</Label>
            <Input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="请输入密码"
            />
          </div>
          {error ? (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : null}
          <Button className="w-full" type="submit" disabled={submitting}>
            {submitting ? '登录中...' : '登录'}
          </Button>
        </form>
        <div className="mt-5 flex items-center justify-between text-sm text-muted-foreground">
          <Link className="hover:text-foreground" to="/register">
            还没有账号？去注册
          </Link>
          <Link className="hover:text-foreground" to="/forgot-password">
            忘记密码
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
