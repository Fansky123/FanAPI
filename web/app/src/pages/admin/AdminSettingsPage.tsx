import { useRef, useState } from 'react'
import { PlusIcon, SaveIcon, Trash2Icon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { adminApi } from '@/lib/api/admin'
import { useAsync } from '@/hooks/use-async'

type SettingsMap = Record<string, string>
type PlanRow = { credits: number; bonus: number; amount: number; origin_amount: number; desc: string }

function totalPlanCredits(plan: PlanRow) {
  return Number(plan.credits || 0) + Number(plan.bonus || 0)
}

function planDiscountLabel(plan: PlanRow) {
  const originAmount = Number(plan.origin_amount || 0)
  const amount = Number(plan.amount || 0)
  if (originAmount <= 0 || originAmount <= amount) return '—'
  return `${(((originAmount - amount) / originAmount) * 100).toFixed(1)}%`
}

function Tip({ children }: { children: React.ReactNode }) {
  return <p className="mt-1 text-xs text-muted-foreground leading-relaxed">{children}</p>
}

function FieldRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-[200px_1fr] items-start gap-4 py-3">
      <label className="pt-2 text-sm font-medium text-right">{label}</label>
      <div>{children}</div>
    </div>
  )
}

function ToggleField({ checked, onChange, children }: { checked: boolean; onChange: (v: boolean) => void; children?: React.ReactNode }) {
  return (
    <div>
      <label className="flex cursor-pointer items-center gap-2">
        <input
          type="checkbox"
          checked={checked}
          onChange={(e) => onChange(e.target.checked)}
          className="h-4 w-4 rounded border-border"
        />
        <span className="text-sm">{checked ? '已开启' : '已关闭'}</span>
      </label>
      {children}
    </div>
  )
}

export function AdminSettingsPage() {
  const { data: rawSettings, loading, error: loadError } = useAsync(async () => {
    const res = await adminApi.getSettings()
    const s = (res as { settings?: SettingsMap }).settings ?? (res as SettingsMap)
    return s as SettingsMap
  }, {} as SettingsMap)

  // Form state
  const [form, setForm] = useState<SettingsMap>({})
  const [formReady, setFormReady] = useState(false)
  const [planRows, setPlanRows] = useState<PlanRow[]>([])
  const [allowCustom, setAllowCustom] = useState(true)
  const [rebatePercent, setRebatePercent] = useState('')
  const [vendorCommPercent, setVendorCommPercent] = useState('')

  // Sync from loaded data once
  if (!loading && !formReady && Object.keys(rawSettings).length > 0) {
    setForm({ ...rawSettings })
    setAllowCustom(rawSettings.recharge_allow_custom !== 'false')
    setRebatePercent(String(parseFloat((parseFloat(rawSettings.default_rebate_ratio || '0') * 100).toFixed(2)) || ''))
    setVendorCommPercent(String(parseFloat((parseFloat(rawSettings.default_vendor_commission || '0') * 100).toFixed(2)) || ''))
    try { setPlanRows(JSON.parse(rawSettings.recharge_plans || '[]')) } catch { setPlanRows([]) }
    setFormReady(true)
  }

  const [mutError, setMutError] = useState('')
  const [saving, setSaving] = useState(false)
  const error = loadError || mutError

  const qrRef = useRef<HTMLInputElement>(null)
  const qqRef = useRef<HTMLInputElement>(null)
  const wechatRef = useRef<HTMLInputElement>(null)

  function set(key: string, value: string) {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  function readFile(file: File, onResult: (base64: string) => void) {
    const reader = new FileReader()
    reader.onload = (e) => { if (e.target?.result) onResult(e.target.result as string) }
    reader.readAsDataURL(file)
  }

  function addPlan() {
    setPlanRows((p) => [...p, { credits: 100, bonus: 0, amount: 10, origin_amount: 0, desc: '' }])
  }
  function removePlan(i: number) {
    setPlanRows((p) => p.filter((_, idx) => idx !== i))
  }
  function updatePlan<K extends keyof PlanRow>(i: number, key: K, value: PlanRow[K]) {
    setPlanRows((p) => p.map((row, idx) => idx === i ? { ...row, [key]: value } : row))
  }

  const epayEnabled = form.epay_enabled === 'true'
  const payApplyEnabled = form.pay_apply_enabled === 'true'
  const payApplyNotifyUrl = `${window.location.origin.replace(':3001', '')}/pay/apply/notify`

  async function save() {
    setSaving(true)
    setMutError('')
    try {
      const payload: SettingsMap = {
        ...form,
        recharge_allow_custom: allowCustom ? 'true' : 'false',
        recharge_plans: JSON.stringify(planRows),
        default_rebate_ratio: (parseFloat(rebatePercent || '0') / 100).toFixed(4),
        default_vendor_commission: (parseFloat(vendorCommPercent || '0') / 100).toFixed(4),
      }
      delete payload.pay_apply_notify_url
      await adminApi.updateSettings(payload)
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Configuration"
        title="系统设置"
        description="配置平台基本信息、支付、公告及充值套餐等全局参数。"
        actions={
          <Button onClick={save} disabled={saving || loading}>
            <SaveIcon data-icon="inline-start" />
            {saving ? '保存中...' : '保存设置'}
          </Button>
        }
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <Card>
        <CardContent className="p-0">
          {loading ? (
            <div className="space-y-4 p-6">
              <Skeleton className="h-10 w-full" />
              <div className="grid gap-3">
                <Skeleton className="h-9 w-1/3" />
                <Skeleton className="h-24 w-full" />
                <Skeleton className="h-9 w-1/2" />
                <Skeleton className="h-24 w-full" />
              </div>
            </div>
          ) : (
            <Tabs defaultValue="basic">
              <TabsList className="w-full rounded-none border-b bg-transparent justify-start gap-0 p-0">
                {(['basic', 'appearance', 'payment', 'notice', 'plans', 'rebate', 'vendor'] as const).map((tab) => (
                  <TabsTrigger
                    key={tab}
                    value={tab}
                    className="rounded-none border-b-2 border-transparent px-4 py-3 text-sm data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:shadow-none"
                  >
                    {{ basic: '基本设置', appearance: '页眉&页脚', payment: '支付设置', notice: '公告&联系', plans: '充值套餐', rebate: '邀请返佣', vendor: '号商设置' }[tab]}
                  </TabsTrigger>
                ))}
              </TabsList>

              {/* 基本设置 */}
              <TabsContent value="basic" className="px-6 pb-6">
                <div className="max-w-2xl divide-y">
                  <FieldRow label="站点名称">
                    <Input value={form.site_name ?? ''} onChange={(e) => set('site_name', e.target.value)} placeholder="例如：FanAPI" />
                    <Tip>显示在浏览器标题栏和页面 Logo 旁</Tip>
                  </FieldRow>
                  <FieldRow label="Logo 图片 URL">
                    <Input value={form.logo_url ?? ''} onChange={(e) => set('logo_url', e.target.value)} placeholder="https://example.com/logo.png（留空则显示文字）" />
                    <Tip>支持 PNG / SVG，建议尺寸 32×32 或 64×64，留空则使用首字母</Tip>
                  </FieldRow>
                  {form.logo_url ? (
                    <FieldRow label="Logo 预览">
                      <div className="flex h-20 w-20 items-center justify-center rounded-xl border bg-muted/30 overflow-hidden p-1">
                        <img src={form.logo_url} alt="Logo" className="max-h-full max-w-full object-contain" />
                      </div>
                    </FieldRow>
                  ) : null}
                </div>
              </TabsContent>

              {/* 页眉&页脚 */}
              <TabsContent value="appearance" className="px-6 pb-6">
                <Alert className="mb-4 mt-2 border-amber-200 bg-amber-50 text-amber-800">
                  <AlertDescription>
                    安全提示：页眉/页脚内容直接通过 <code>dangerouslySetInnerHTML</code> 渲染，请勿插入不可信的第三方脚本，避免 XSS 风险。
                  </AlertDescription>
                </Alert>
                <div className="max-w-2xl divide-y">
                  <FieldRow label="页眉 HTML">
                    <Textarea
                      value={form.header_html ?? ''}
                      onChange={(e) => set('header_html', e.target.value)}
                      rows={6}
                      className="font-mono text-xs"
                      placeholder='<div style="text-align:center;padding:8px;background:#1677ff;color:#fff">公告：系统维护中</div>'
                    />
                    <Tip>留空则不显示页眉；支持 HTML 和内联样式</Tip>
                  </FieldRow>
                  <FieldRow label="页脚 HTML">
                    <Textarea
                      value={form.footer_html ?? ''}
                      onChange={(e) => set('footer_html', e.target.value)}
                      rows={6}
                      className="font-mono text-xs"
                      placeholder='<div style="text-align:center;padding:16px;color:#888">© 2025 FanAPI</div>'
                    />
                    <Tip>留空则不显示页脚；支持 HTML 和内联样式</Tip>
                  </FieldRow>
                  {(form.header_html || form.footer_html) ? (
                    <FieldRow label="预览">
                      <div className="rounded-lg border overflow-hidden">
                        <div className="bg-muted/40 px-3 py-1.5 text-xs text-muted-foreground">页眉预览</div>
                        {/* eslint-disable-next-line react/no-danger */}
                        <div dangerouslySetInnerHTML={{ __html: form.header_html || '<span style="color:#aaa">（空）</span>' }} />
                        <Separator />
                        <div className="bg-muted/40 px-3 py-1.5 text-xs text-muted-foreground">页脚预览</div>
                        {/* eslint-disable-next-line react/no-danger */}
                        <div dangerouslySetInnerHTML={{ __html: form.footer_html || '<span style="color:#aaa">（空）</span>' }} />
                      </div>
                    </FieldRow>
                  ) : null}
                </div>
              </TabsContent>

              {/* 支付设置 */}
              <TabsContent value="payment" className="px-6 pb-6">
                <div className="max-w-2xl divide-y">
                  <FieldRow label="启用易支付">
                    <ToggleField
                      checked={epayEnabled}
                      onChange={(v) => {
                        set('epay_enabled', v ? 'true' : 'false')
                        if (v) set('pay_apply_enabled', 'false')
                      }}
                    >
                      <Tip>开启后用户可以通过易支付（支付宝/微信）充值余额</Tip>
                    </ToggleField>
                  </FieldRow>
                  {epayEnabled ? (
                    <>
                      <FieldRow label="易支付地址">
                        <Input value={form.epay_url ?? ''} onChange={(e) => set('epay_url', e.target.value)} placeholder="https://your-epay.com" />
                        <Tip>易支付平台的域名（不含末尾斜杠）</Tip>
                      </FieldRow>
                      <FieldRow label="商户 PID">
                        <Input value={form.epay_pid ?? ''} onChange={(e) => set('epay_pid', e.target.value)} placeholder="您的易支付商户 PID" />
                      </FieldRow>
                      <FieldRow label="商户密钥">
                        <Input type="password" value={form.epay_key ?? ''} onChange={(e) => set('epay_key', e.target.value)} placeholder="您的易支付商户密钥" />
                      </FieldRow>
                      <FieldRow label="异步通知地址">
                        <Input value={form.epay_notify_url ?? ''} onChange={(e) => set('epay_notify_url', e.target.value)} placeholder="https://api.yoursite.com/pay/epay/callback" />
                        <Tip>易支付回调到本系统的地址，必须可从公网访问</Tip>
                      </FieldRow>
                      <FieldRow label="同步跳转地址">
                        <Input value={form.epay_return_url ?? ''} onChange={(e) => set('epay_return_url', e.target.value)} placeholder="https://yoursite.com/billing" />
                        <Tip>用户支付成功后跳回的前端页面地址</Tip>
                      </FieldRow>
                    </>
                  ) : null}
                  <FieldRow label="">
                    <Separator className="my-2" />
                  </FieldRow>
                  <FieldRow label="启用中台支付">
                    <ToggleField
                      checked={payApplyEnabled}
                      onChange={(v) => {
                        set('pay_apply_enabled', v ? 'true' : 'false')
                        if (v) set('epay_enabled', 'false')
                      }}
                    >
                      <Tip>开启后用户可通过支付中台（微信/支付宝）充值余额</Tip>
                    </ToggleField>
                  </FieldRow>
                  {payApplyEnabled ? (
                    <>
                      <FieldRow label="中台根地址">
                        <Input value={form.pay_apply_urlroot ?? ''} onChange={(e) => set('pay_apply_urlroot', e.target.value)} placeholder="https://pay.example.com" />
                        <Tip>支付中台的域名（不含末尾斜杠）</Tip>
                      </FieldRow>
                      <FieldRow label="中台商品 Key">
                        <Input type="password" value={form.pay_apply_key ?? ''} onChange={(e) => set('pay_apply_key', e.target.value)} placeholder="支付中台分配的商品 key" />
                        <Tip>中台回调时会携带此 key 用于验签，请妥善保管</Tip>
                      </FieldRow>
                      <FieldRow label="回调地址">
                        <Input readOnly value={payApplyNotifyUrl} className="bg-muted/30 font-mono text-xs" />
                        <Tip>将此地址填写到支付中台的回调配置中，必须可从公网访问</Tip>
                      </FieldRow>
                    </>
                  ) : null}
                </div>
              </TabsContent>

              {/* 公告&联系方式 */}
              <TabsContent value="notice" className="px-6 pb-6">
                <div className="max-w-2xl divide-y">
                  <FieldRow label="公告标题">
                    <Input value={form.notice_title ?? ''} onChange={(e) => set('notice_title', e.target.value)} placeholder="例如：📢 最新公告" />
                    <Tip>显示在用户数据看板右侧，留空则不显示公告模块</Tip>
                  </FieldRow>
                  <FieldRow label="公告内容">
                    <Textarea value={form.notice_content ?? ''} onChange={(e) => set('notice_content', e.target.value)} rows={5} placeholder="支持换行，每行一条公告内容" />
                    <Tip>纯文本，每行作为一条，不支持 HTML</Tip>
                  </FieldRow>
                  <FieldRow label="联系方式">
                    <Textarea
                      value={form.contact_info ?? ''}
                      onChange={(e) => set('contact_info', e.target.value)}
                      rows={4}
                      placeholder={`微信：fanapi\nQQ群：123456789\n邮箱：support@example.com`}
                    />
                    <Tip>纯文本，每行一条联系方式，显示在数据看板公告区域</Tip>
                  </FieldRow>
                  <FieldRow label="QQ 交流群二维码">
                    <div className="flex gap-2">
                      <Input value={form.qq_group_url ?? ''} onChange={(e) => set('qq_group_url', e.target.value)} placeholder="图片 URL 或粘贴 base64" className="flex-1" />
                      <Button type="button" variant="outline" size="sm" onClick={() => qqRef.current?.click()}>上传</Button>
                    </div>
                    <input ref={qqRef} type="file" accept="image/*" className="hidden" onChange={(e) => { const f = e.target.files?.[0]; if (f) readFile(f, (b64) => set('qq_group_url', b64)); e.target.value = '' }} />
                    {form.qq_group_url ? <img src={form.qq_group_url} alt="QQ群二维码" className="mt-2 h-28 w-28 rounded-xl border object-contain p-1" /> : null}
                    <Tip>填写后用户页面头部将显示「QQ交流群」按钮，留空不显示</Tip>
                  </FieldRow>
                  <FieldRow label="微信客服二维码">
                    <div className="flex gap-2">
                      <Input value={form.wechat_cs_url ?? ''} onChange={(e) => set('wechat_cs_url', e.target.value)} placeholder="图片 URL 或粘贴 base64" className="flex-1" />
                      <Button type="button" variant="outline" size="sm" onClick={() => wechatRef.current?.click()}>上传</Button>
                    </div>
                    <input ref={wechatRef} type="file" accept="image/*" className="hidden" onChange={(e) => { const f = e.target.files?.[0]; if (f) readFile(f, (b64) => set('wechat_cs_url', b64)); e.target.value = '' }} />
                    {form.wechat_cs_url ? <img src={form.wechat_cs_url} alt="微信客服" className="mt-2 h-28 w-28 rounded-xl border object-contain p-1" /> : null}
                    <Tip>填写后用户页面头部将显示「微信客服」按钮，留空不显示</Tip>
                  </FieldRow>
                  <FieldRow label="二维码图片">
                    <div className="flex gap-2">
                      <Input value={form.qrcode_url ?? ''} onChange={(e) => set('qrcode_url', e.target.value)} placeholder="图片 URL 或粘贴 base64" className="flex-1" />
                      <Button type="button" variant="outline" size="sm" onClick={() => qrRef.current?.click()}>上传</Button>
                    </div>
                    <input ref={qrRef} type="file" accept="image/*" className="hidden" onChange={(e) => { const f = e.target.files?.[0]; if (f) readFile(f, (b64) => set('qrcode_url', b64)); e.target.value = '' }} />
                    {form.qrcode_url ? <img src={form.qrcode_url} alt="二维码" className="mt-2 h-28 w-28 rounded-xl border object-contain p-1" /> : null}
                    <Tip>支持图片 URL 或本地上传（自动转 base64）；留空则不显示</Tip>
                  </FieldRow>
                </div>
              </TabsContent>

              {/* 充值套餐 */}
              <TabsContent value="plans" className="px-6 pb-6">
                <div className="max-w-3xl">
                  <FieldRow label="允许自定义金额">
                    <ToggleField checked={allowCustom} onChange={setAllowCustom}>
                      <Tip>开启后用户可在套餐之外自由输入充值金额；关闭则只能选套餐</Tip>
                    </ToggleField>
                  </FieldRow>
                  <div className="mt-4">
                    <div className="mb-2 text-sm font-medium">套餐列表</div>
                    {planRows.length > 0 ? (
                      <div className="rounded-lg border overflow-hidden mb-3">
                        <table className="w-full text-sm">
                          <thead className="bg-muted/30">
                            <tr className="border-b">
                              <th className="px-3 py-2 text-left font-medium">积分数</th>
                              <th className="px-3 py-2 text-left font-medium">赠送积分</th>
                              <th className="px-3 py-2 text-left font-medium">到账积分</th>
                              <th className="px-3 py-2 text-left font-medium">金额（元）</th>
                              <th className="px-3 py-2 text-left font-medium">原价（元）</th>
                              <th className="px-3 py-2 text-left font-medium">折扣</th>
                              <th className="px-3 py-2 text-left font-medium">描述</th>
                              <th className="px-3 py-2 w-10"></th>
                            </tr>
                          </thead>
                          <tbody className="divide-y">
                            {planRows.map((row, i) => (
                              <tr key={i}>
                                <td className="px-2 py-1.5">
                                  <Input type="number" value={row.credits} min={1} onChange={(e) => updatePlan(i, 'credits', Number(e.target.value))} className="h-8" />
                                </td>
                                <td className="px-2 py-1.5">
                                  <Input type="number" value={row.bonus} min={0} onChange={(e) => updatePlan(i, 'bonus', Number(e.target.value))} className="h-8" />
                                </td>
                                <td className="px-3 py-1.5 text-sm font-medium text-emerald-600">
                                  {totalPlanCredits(row).toLocaleString()}
                                </td>
                                <td className="px-2 py-1.5">
                                  <Input type="number" value={row.amount} min={0.01} step={0.01} onChange={(e) => updatePlan(i, 'amount', Number(e.target.value))} className="h-8" />
                                </td>
                                <td className="px-2 py-1.5">
                                  <Input type="number" value={row.origin_amount} min={0} step={0.01} onChange={(e) => updatePlan(i, 'origin_amount', Number(e.target.value))} className="h-8" />
                                </td>
                                <td className="px-3 py-1.5 text-sm text-muted-foreground">
                                  {planDiscountLabel(row)}
                                </td>
                                <td className="px-2 py-1.5">
                                  <Input value={row.desc} onChange={(e) => updatePlan(i, 'desc', e.target.value)} placeholder="购买可获得xx积分" className="h-8" />
                                </td>
                                <td className="px-2 py-1.5">
                                  <Button type="button" variant="ghost" size="sm" className="h-8 w-8 p-0 text-red-500 hover:text-red-600" onClick={() => removePlan(i)}>
                                    <Trash2Icon className="h-4 w-4" />
                                  </Button>
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    ) : null}
                    <Button type="button" variant="outline" size="sm" onClick={addPlan}>
                      <PlusIcon className="mr-1 h-4 w-4" />
                      添加套餐
                    </Button>
                    <Tip>套餐顺序即展示顺序；不设置套餐则只显示自定义金额输入框</Tip>
                    {planRows.length > 0 ? (
                      <div className="mt-4 grid gap-3 lg:grid-cols-3">
                        {planRows.map((row, i) => (
                          <Card key={`plan-preview-${i}`}>
                            <CardHeader className="pb-3">
                              <CardTitle className="text-base">套餐 {i + 1}</CardTitle>
                              <CardDescription>用户端将按当前顺序展示</CardDescription>
                            </CardHeader>
                            <CardContent className="flex flex-col gap-2 text-sm">
                              <div className="font-semibold text-blue-600">￥{Number(row.amount || 0).toFixed(2)}</div>
                              <div>到账积分：<span className="font-medium text-emerald-600">{totalPlanCredits(row).toLocaleString()}</span></div>
                              <div>基础积分：{Number(row.credits || 0).toLocaleString()}</div>
                              <div>赠送积分：{Number(row.bonus || 0).toLocaleString()}</div>
                              <div>原价：{Number(row.origin_amount || 0) > 0 ? `￥${Number(row.origin_amount).toFixed(2)}` : '—'}</div>
                              <div>折扣：{planDiscountLabel(row)}</div>
                              <div className="text-muted-foreground">{row.desc || '未填写描述，前台将只展示金额与积分。'}</div>
                            </CardContent>
                          </Card>
                        ))}
                      </div>
                    ) : null}
                  </div>
                </div>
              </TabsContent>

              {/* 邀请返佣 */}
              <TabsContent value="rebate" className="px-6 pb-6">
                <Alert className="mb-4 mt-2">
                  <AlertDescription>
                    用户邀请新用户注册后，被邀请人每次消费将按比例给邀请人增加冻结积分，冻结积分可手动解冻为可用积分。
                  </AlertDescription>
                </Alert>
                <div className="max-w-2xl divide-y">
                  <FieldRow label="全局返佣比例（%）">
                    <div className="flex items-center gap-2">
                      <Input
                        type="number"
                        value={rebatePercent}
                        min={0}
                        max={100}
                        step={0.01}
                        onChange={(e) => setRebatePercent(e.target.value)}
                        className="w-32"
                      />
                      <span className="text-sm text-muted-foreground">%</span>
                    </div>
                    <Tip>被邀请人消费金额的该比例将冻结给邀请人（例：5 表示消费 100 积分返 5 冻结积分）。管理员可为单个用户单独设置比例以覆盖此全局值。</Tip>
                  </FieldRow>
                </div>
              </TabsContent>

              {/* 号商设置 */}
              <TabsContent value="vendor" className="px-6 pb-6">
                <Alert className="mb-4 mt-2">
                  <AlertDescription>
                    号商向号池提供 API Key，每次 Key 被使用时，平台按比例抽成，剩余部分计入号商余额。
                  </AlertDescription>
                </Alert>
                <div className="max-w-2xl divide-y">
                  <FieldRow label="全局平台抽成比例（%）">
                    <div className="flex items-center gap-2">
                      <Input
                        type="number"
                        value={vendorCommPercent}
                        min={0}
                        max={100}
                        step={0.01}
                        onChange={(e) => setVendorCommPercent(e.target.value)}
                        className="w-32"
                      />
                      <span className="text-sm text-muted-foreground">%</span>
                    </div>
                    <Tip>平台从号商收益中抽取的比例（例：2 表示号商实得 98%）。可为单个号商单独设置以覆盖此全局值。</Tip>
                  </FieldRow>
                </div>
              </TabsContent>
            </Tabs>
          )}
        </CardContent>
      </Card>
    </>
  )
}

