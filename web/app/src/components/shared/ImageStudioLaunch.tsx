import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { createHttpClient, getApiErrorMessage } from '@/lib/api/http'
const http = createHttpClient('user')
export function ImageStudioLaunch({ prompt = '' }: { prompt?: string }) {
 const [error, setError] = useState('')
 const [attempt, setAttempt] = useState(0)
 useEffect(() => {
  const controller = new AbortController()
  http.post<{url:string}>('/user/image-studio/handoff', {all_keys:true,prompt}, {signal:controller.signal}).then(result => {
   if(controller.signal.aborted) return
   const target = new URL(result.url)
   if(target.protocol !== 'https:' && !(target.protocol === 'http:' && ['localhost','127.0.0.1'].includes(target.hostname))) throw new Error('图片工作台地址无效')
   window.location.replace(target.href)
  }).catch(err => {if(!controller.signal.aborted) setError(getApiErrorMessage(err))})
  return () => controller.abort()
 }, [prompt,attempt])
 return <div className="space-y-4 py-8">
  {error ? <><p role="alert" className="text-destructive">{error}</p><Button onClick={() => {setError('');setAttempt(value => value+1)}}>重新进入</Button><a href="/keys" className="ml-4 underline">管理 API 密钥</a></> : <p role="status">正在进入图片工作台…</p>}
 </div>
}
