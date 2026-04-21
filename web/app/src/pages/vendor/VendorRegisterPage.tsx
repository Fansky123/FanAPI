import type { FormEvent } from 'react'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { createHttpClient, getApiErrorMessage } from '@/lib/api/http'

const vendorHttp = createHttpClient()

export function VendorRegisterPage() {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')

    try {
      await vendorHttp.post<{ id: number; username: string }>('/vendor/auth/register', {
        username,
        password,
      })
      // Backend only returns {id, username} — login separately to get token
      navigate('/vendor/login')
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  return (
    <Card className="w-full max-w-xl border-border/70 bg-card/92 shadow-lg">
      <CardHeader>
        <CardTitle>注册 Vendor 账号</CardTitle>
      </CardHeader>
      <CardContent>
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <Input value={username} onChange={(event) => setUsername(event.target.value)} placeholder="用户名（3-32 字符）" minLength={3} maxLength={32} required />
          <Input type="password" value={password} onChange={(event) => setPassword(event.target.value)} placeholder="密码（至少 6 位）" minLength={6} required />
          {error ? <div className="text-sm text-destructive">{error}</div> : null}
          <Button className="w-full" type="submit">注册并进入 Vendor 端</Button>
          <p className="text-center text-sm text-muted-foreground">
            已有账号？{' '}
            <Link to="/vendor/login" className="text-primary hover:underline">
              立即登录
            </Link>
          </p>
        </form>
      </CardContent>
    </Card>
  )
}
