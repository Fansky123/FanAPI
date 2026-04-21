import type { FormEvent } from 'react'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { createHttpClient, getApiErrorMessage } from '@/lib/api/http'
import { setRoleToken, setSiteModePreference } from '@/lib/auth/storage'

const vendorHttp = createHttpClient()

export function VendorLoginPage() {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')

    try {
      const response = await vendorHttp.post<{ token: string }>('/vendor/auth/login', {
        username,
        password,
      })
      setRoleToken('vendor', response.token)
      setSiteModePreference('vendor')
      navigate('/vendor/dashboard')
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  return (
    <Card className="w-full max-w-xl border-border/70 bg-card/92 shadow-lg">
      <CardHeader>
        <CardTitle>登录 Vendor 端</CardTitle>
      </CardHeader>
      <CardContent>
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <Input value={username} onChange={(event) => setUsername(event.target.value)} placeholder="用户名" />
          <Input type="password" value={password} onChange={(event) => setPassword(event.target.value)} placeholder="密码" />
          {error ? <div className="text-sm text-destructive">{error}</div> : null}
          <Button className="w-full" type="submit">进入 Vendor 端</Button>
        </form>
      </CardContent>
    </Card>
  )
}
