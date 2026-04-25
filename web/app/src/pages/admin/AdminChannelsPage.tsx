import { useMemo, useState } from 'react'
import { CopyIcon, PlusIcon, SaveIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { adminApi, type AdminChannel, type AdminKeyPool } from '@/lib/api/admin'
import { useAsync } from '@/hooks/use-async'

type ChannelForm = {
  id?: number
  name: string
  model: string
  type: string
  protocol: string
  base_url: string
  method: string
  query_url: string
  query_method: string
  timeout_ms: string
  query_timeout_ms: string
  billing_type: string
  headers_text: string
  billing_config_text: string
  pricing_groups: PricingGroupForm[]
  // token billing
  billing_input_price: string
  billing_output_price: string
  billing_input_cost: string
  billing_output_cost: string
  billing_cache_create_price: string
  billing_cache_create_cost: string
  billing_cache_read_price: string
  billing_cache_read_cost: string
  billing_input_from_response: boolean
  // image billing
  billing_base_price: string
  billing_base_cost: string
  billing_size_price_1k: string
  billing_size_price_2k: string
  billing_size_price_3k: string
  billing_size_price_4k: string
  billing_size_cost_1k: string
  billing_size_cost_2k: string
  billing_size_cost_3k: string
  billing_size_cost_4k: string
  billing_default_size_price: string
  billing_default_size_cost: string
  // video / audio billing
  billing_price_per_second: string
  billing_cost_per_second: string
  // count billing
  billing_price_per_call: string
  billing_cost_per_call: string
  billing_script: string
  request_script: string
  response_script: string
  query_script: string
  error_script: string
  key_pool_id: string
  auth_type: string
  auth_param_name: string
  auth_region: string
  auth_service: string
  passthrough_headers: boolean
  passthrough_body: boolean
  weight: string
  priority: string
  icon_url: string
  description: string
  is_active: boolean
}

type PricingGroupForm = {
  group: string
  input_price_per_1m_tokens: string
  output_price_per_1m_tokens: string
  base_price: string
  default_size_price: string
  size_price_1k: string
  size_price_2k: string
  size_price_3k: string
  size_price_4k: string
  price_per_second: string
  price_per_call: string
  extra_config: Record<string, unknown>
}

const emptyJson = '{}'
const sizeTierKeys = ['1k', '2k', '3k', '4k'] as const
const creditsPrecision = 1_000_000
const structuredBillingConfigKeys = new Set([
  'input_price_per_1m_tokens',
  'output_price_per_1m_tokens',
  'input_cost_per_1m_tokens',
  'output_cost_per_1m_tokens',
  'cache_creation_price_per_1m_tokens',
  'cache_creation_cost_per_1m_tokens',
  'cache_read_price_per_1m_tokens',
  'cache_read_cost_per_1m_tokens',
  'input_from_response',
  'base_price',
  'base_cost',
  'size_prices',
  'size_costs',
  'default_size_price',
  'default_size_cost',
  'price_per_second',
  'cost_per_second',
  'price_per_call',
  'cost_per_call',
  'pricing_groups',
])
const pricingGroupOverrideKeys = new Set([
  'input_price_per_1m_tokens',
  'output_price_per_1m_tokens',
  'base_price',
  'default_size_price',
  'size_prices',
  'price_per_second',
  'price_per_call',
])

function createEmptyPricingGroup(): PricingGroupForm {
  return {
    group: '',
    input_price_per_1m_tokens: '',
    output_price_per_1m_tokens: '',
    base_price: '',
    default_size_price: '',
    size_price_1k: '',
    size_price_2k: '',
    size_price_3k: '',
    size_price_4k: '',
    price_per_second: '',
    price_per_call: '',
    extra_config: {},
  }
}

const emptyForm: ChannelForm = {
  name: '',
  model: '',
  type: 'llm',
  protocol: 'openai',
  base_url: '',
  method: 'POST',
  query_url: '',
  query_method: 'GET',
  timeout_ms: '60000',
  query_timeout_ms: '30000',
  billing_type: 'token',
  headers_text: emptyJson,
  billing_config_text: emptyJson,
  pricing_groups: [],
  billing_input_price: '',
  billing_output_price: '',
  billing_input_cost: '',
  billing_output_cost: '',
  billing_cache_create_price: '',
  billing_cache_create_cost: '',
  billing_cache_read_price: '',
  billing_cache_read_cost: '',
  billing_input_from_response: false,
  billing_base_price: '',
  billing_base_cost: '',
  billing_size_price_1k: '',
  billing_size_price_2k: '',
  billing_size_price_3k: '',
  billing_size_price_4k: '',
  billing_size_cost_1k: '',
  billing_size_cost_2k: '',
  billing_size_cost_3k: '',
  billing_size_cost_4k: '',
  billing_default_size_price: '',
  billing_default_size_cost: '',
  billing_price_per_second: '',
  billing_cost_per_second: '',
  billing_price_per_call: '',
  billing_cost_per_call: '',
  billing_script: '',
  request_script: '',
  response_script: '',
  query_script: '',
  error_script: '',
  key_pool_id: '',
  auth_type: 'bearer',
  auth_param_name: '',
  auth_region: '',
  auth_service: '',
  passthrough_headers: false,
  passthrough_body: false,
  weight: '1',
  priority: '0',
  icon_url: '',
  description: '',
  is_active: true,
}

function prettyJson(value: unknown) {
  if (!value || (typeof value === 'object' && Object.keys(value as object).length === 0)) {
    return emptyJson
  }
  return JSON.stringify(value, null, 2)
}

function parseJsonField(label: string, value: string) {
  try {
    return JSON.parse(value || emptyJson) as Record<string, unknown>
  } catch {
    throw new Error(`${label} 不是合法 JSON`)
  }
}

function parseAmount(value: unknown) {
  if (value === undefined || value === null || value === '') {
    return undefined
  }
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : undefined
}

function formatCnyValue(value: number) {
  return value.toFixed(6).replace(/\.0+$/, '').replace(/(\.\d*?[1-9])0+$/, '$1')
}

function toCnyInput(value: unknown): string {
  const amount = parseAmount(value)
  if (amount === undefined) {
    return ''
  }
  return formatCnyValue(amount / creditsPrecision)
}

function formatCnyDisplay(value: unknown): string {
  const amount = parseAmount(value)
  if (amount === undefined) {
    return 'CNY 0'
  }
  return `CNY ${formatCnyValue(amount / creditsPrecision)}`
}

function fromCnyInput(value: string) {
  const trimmed = value.trim()
  if (!trimmed) {
    return undefined
  }
  const parsed = Number(trimmed)
  if (!Number.isFinite(parsed)) {
    return undefined
  }
  return Math.round(parsed * creditsPrecision)
}

function getNum(cfg: Record<string, unknown>, key: string): string {
  return toCnyInput(cfg[key])
}

function getTierNum(cfg: Record<string, unknown>, key: 'size_prices' | 'size_costs', tier: (typeof sizeTierKeys)[number]): string {
  const raw = cfg[key]
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) {
    return ''
  }
  const value = (raw as Record<string, unknown>)[tier]
  return toCnyInput(value)
}

function buildAdvancedBillingConfigText(cfg: Record<string, unknown>) {
  const rest = Object.fromEntries(
    Object.entries(cfg).filter(([key]) => !structuredBillingConfigKeys.has(key))
  )
  return prettyJson(rest)
}

function buildSizeMap(entries: Array<[string, string]>) {
  const mapped = Object.fromEntries(
    entries
      .filter(([, value]) => value.trim() !== '')
      .flatMap(([key, value]) => {
        const parsed = fromCnyInput(value)
        return parsed === undefined ? [] : [[key, parsed]]
      })
  )
  return Object.keys(mapped).length > 0 ? mapped : undefined
}

function extractPricingGroups(cfg: Record<string, unknown>): PricingGroupForm[] {
  const raw = cfg.pricing_groups
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) {
    return []
  }

  return Object.entries(raw)
    .filter(([, value]) => value && typeof value === 'object' && !Array.isArray(value))
    .map(([group, value]) => {
      const override = value as Record<string, unknown>
      return {
        group,
        input_price_per_1m_tokens: getNum(override, 'input_price_per_1m_tokens'),
        output_price_per_1m_tokens: getNum(override, 'output_price_per_1m_tokens'),
        base_price: getNum(override, 'base_price'),
        default_size_price: getNum(override, 'default_size_price'),
        size_price_1k: getTierNum(override, 'size_prices', '1k'),
        size_price_2k: getTierNum(override, 'size_prices', '2k'),
        size_price_3k: getTierNum(override, 'size_prices', '3k'),
        size_price_4k: getTierNum(override, 'size_prices', '4k'),
        price_per_second: getNum(override, 'price_per_second'),
        price_per_call: getNum(override, 'price_per_call'),
        extra_config: Object.fromEntries(
          Object.entries(override).filter(([key]) => !pricingGroupOverrideKeys.has(key))
        ),
      }
    })
}

function buildPricingGroups(groups: PricingGroupForm[], billingType: string): Record<string, unknown> | undefined {
  const entries = groups.flatMap((group) => {
    const name = group.group.trim()
    if (!name) {
      return []
    }

    const override: Record<string, unknown> = { ...group.extra_config }
    for (const key of pricingGroupOverrideKeys) {
      delete override[key]
    }

    const setNumber = (key: string, value: string) => {
      const parsed = fromCnyInput(value)
      if (parsed === undefined) {
        return
      }
      override[key] = parsed
    }

    switch (billingType) {
      case 'token':
        setNumber('input_price_per_1m_tokens', group.input_price_per_1m_tokens)
        setNumber('output_price_per_1m_tokens', group.output_price_per_1m_tokens)
        break
      case 'image': {
        setNumber('base_price', group.base_price)
        setNumber('default_size_price', group.default_size_price)
        const sizePrices = buildSizeMap([
          ['1k', group.size_price_1k],
          ['2k', group.size_price_2k],
          ['3k', group.size_price_3k],
          ['4k', group.size_price_4k],
        ])
        if (sizePrices) {
          override.size_prices = sizePrices
        }
        break
      }
      case 'video':
      case 'audio':
        setNumber('price_per_second', group.price_per_second)
        break
      case 'count':
        setNumber('price_per_call', group.price_per_call)
        break
    }

    if (Object.keys(override).length === 0) {
      return []
    }

    return [[name, override] as const]
  })

  if (entries.length === 0) {
    return undefined
  }

  return Object.fromEntries(entries)
}

function formatBillingSummary(billingType: string | undefined, cfg: Record<string, unknown>, mode: 'price' | 'cost' = 'price') {
  const sizeMapKey = mode === 'price' ? 'size_prices' : 'size_costs'
  const defaultSizeKey = mode === 'price' ? 'default_size_price' : 'default_size_cost'
  const baseKey = mode === 'price' ? 'base_price' : 'base_cost'
  const pricePerSecondKey = mode === 'price' ? 'price_per_second' : 'cost_per_second'
  const pricePerCallKey = mode === 'price' ? 'price_per_call' : 'cost_per_call'
  const inputKey = mode === 'price' ? 'input_price_per_1m_tokens' : 'input_cost_per_1m_tokens'
  const outputKey = mode === 'price' ? 'output_price_per_1m_tokens' : 'output_cost_per_1m_tokens'

  switch (billingType) {
    case 'token':
      return `输入 ${formatCnyDisplay(cfg[inputKey])} / 输出 ${formatCnyDisplay(cfg[outputKey])}`
    case 'image': {
      const sizePrices = cfg[sizeMapKey]
      if (sizePrices && typeof sizePrices === 'object' && !Array.isArray(sizePrices)) {
        const parts = sizeTierKeys
          .map((key) => {
            const value = (sizePrices as Record<string, unknown>)[key]
            return value !== undefined && value !== null && value !== 0 ? `${key}:${formatCnyDisplay(value)}` : null
          })
          .filter(Boolean)
        if (parts.length > 0) {
          return parts.join(' / ')
        }
      }
      return `基础 ${formatCnyDisplay(cfg[defaultSizeKey] ?? cfg[baseKey] ?? 0)}`
    }
    case 'video':
    case 'audio':
      return `${formatCnyDisplay(cfg[pricePerSecondKey] ?? 0)} /秒`
    case 'count':
      return `${formatCnyDisplay(cfg[pricePerCallKey] ?? 0)} /次`
    default:
      return '—'
  }
}

function formatGroupPricing(channel: AdminChannel) {
  if (channel.billing_type === 'custom') {
    return '—'
  }

  const pricingGroups = channel.billing_config?.pricing_groups
  if (!pricingGroups || typeof pricingGroups !== 'object' || Array.isArray(pricingGroups)) {
    return '—'
  }

  const entries = Object.entries(pricingGroups)
    .filter(([, value]) => value && typeof value === 'object' && !Array.isArray(value))
    .map(([group, value]) => `${group}: ${formatBillingSummary(channel.billing_type, value as Record<string, unknown>, 'price')}`)

  if (entries.length === 0) {
    return '—'
  }

  if (entries.length <= 2) {
    return entries.join(' | ')
  }

  return `${entries.slice(0, 2).join(' | ')} | +${entries.length - 2}组`
}

function buildBillingConfig(form: ChannelForm): Record<string, unknown> {
  const cfg = parseJsonField('高级计费配置', form.billing_config_text)

  for (const key of structuredBillingConfigKeys) {
    delete cfg[key]
  }

  const setNumber = (key: string, value: string) => {
    const parsed = fromCnyInput(value)
    if (parsed === undefined) {
      return
    }
    cfg[key] = parsed
  }

  switch (form.billing_type) {
    case 'token':
      setNumber('input_price_per_1m_tokens', form.billing_input_price)
      setNumber('output_price_per_1m_tokens', form.billing_output_price)
      setNumber('input_cost_per_1m_tokens', form.billing_input_cost)
      setNumber('output_cost_per_1m_tokens', form.billing_output_cost)
      setNumber('cache_creation_price_per_1m_tokens', form.billing_cache_create_price)
      setNumber('cache_creation_cost_per_1m_tokens', form.billing_cache_create_cost)
      setNumber('cache_read_price_per_1m_tokens', form.billing_cache_read_price)
      setNumber('cache_read_cost_per_1m_tokens', form.billing_cache_read_cost)
      if (form.billing_input_from_response) {
        cfg.input_from_response = true
      }
      break
    case 'image': {
      setNumber('base_price', form.billing_base_price)
      setNumber('base_cost', form.billing_base_cost)
      setNumber('default_size_price', form.billing_default_size_price)
      setNumber('default_size_cost', form.billing_default_size_cost)

      const sizePrices = buildSizeMap([
        ['1k', form.billing_size_price_1k],
        ['2k', form.billing_size_price_2k],
        ['3k', form.billing_size_price_3k],
        ['4k', form.billing_size_price_4k],
      ])
      if (sizePrices) {
        cfg.size_prices = sizePrices
      }

      const sizeCosts = buildSizeMap([
        ['1k', form.billing_size_cost_1k],
        ['2k', form.billing_size_cost_2k],
        ['3k', form.billing_size_cost_3k],
        ['4k', form.billing_size_cost_4k],
      ])
      if (sizeCosts) {
        cfg.size_costs = sizeCosts
      }
      break
    }
    case 'video':
    case 'audio':
      setNumber('price_per_second', form.billing_price_per_second)
      setNumber('cost_per_second', form.billing_cost_per_second)
      break
    case 'count':
      setNumber('price_per_call', form.billing_price_per_call)
      setNumber('cost_per_call', form.billing_cost_per_call)
      break
  }

  const pricingGroups = buildPricingGroups(form.pricing_groups, form.billing_type)
  if (pricingGroups) {
    cfg.pricing_groups = pricingGroups
  }

  return cfg
}

function buildFormFromChannel(row: AdminChannel, isCopy = false): ChannelForm {
  const billingConfig = row.billing_config ?? {}

  return {
    ...emptyForm,
    id: isCopy ? undefined : row.id,
    name: isCopy ? `${row.name ?? ''} - 副本` : row.name ?? '',
    model: row.model ?? row.routing_model ?? '',
    type: row.type ?? 'llm',
    protocol: row.protocol ?? 'openai',
    base_url: row.base_url ?? '',
    method: row.method ?? 'POST',
    query_url: row.query_url ?? '',
    query_method: row.query_method ?? 'GET',
    timeout_ms: String(row.timeout_ms ?? 60000),
    query_timeout_ms: String(row.query_timeout_ms ?? 30000),
    billing_type: row.billing_type ?? 'token',
    headers_text: prettyJson(row.headers),
    billing_config_text: buildAdvancedBillingConfigText(billingConfig),
    pricing_groups: extractPricingGroups(billingConfig),
    billing_input_price: getNum(billingConfig, 'input_price_per_1m_tokens'),
    billing_output_price: getNum(billingConfig, 'output_price_per_1m_tokens'),
    billing_input_cost: getNum(billingConfig, 'input_cost_per_1m_tokens'),
    billing_output_cost: getNum(billingConfig, 'output_cost_per_1m_tokens'),
    billing_cache_create_price: getNum(billingConfig, 'cache_creation_price_per_1m_tokens'),
    billing_cache_create_cost: getNum(billingConfig, 'cache_creation_cost_per_1m_tokens'),
    billing_cache_read_price: getNum(billingConfig, 'cache_read_price_per_1m_tokens'),
    billing_cache_read_cost: getNum(billingConfig, 'cache_read_cost_per_1m_tokens'),
    billing_input_from_response: Boolean(billingConfig.input_from_response),
    billing_base_price: getNum(billingConfig, 'base_price'),
    billing_base_cost: getNum(billingConfig, 'base_cost'),
    billing_size_price_1k: getTierNum(billingConfig, 'size_prices', '1k'),
    billing_size_price_2k: getTierNum(billingConfig, 'size_prices', '2k'),
    billing_size_price_3k: getTierNum(billingConfig, 'size_prices', '3k'),
    billing_size_price_4k: getTierNum(billingConfig, 'size_prices', '4k'),
    billing_size_cost_1k: getTierNum(billingConfig, 'size_costs', '1k'),
    billing_size_cost_2k: getTierNum(billingConfig, 'size_costs', '2k'),
    billing_size_cost_3k: getTierNum(billingConfig, 'size_costs', '3k'),
    billing_size_cost_4k: getTierNum(billingConfig, 'size_costs', '4k'),
    billing_default_size_price: getNum(billingConfig, 'default_size_price'),
    billing_default_size_cost: getNum(billingConfig, 'default_size_cost'),
    billing_price_per_second: getNum(billingConfig, 'price_per_second'),
    billing_cost_per_second: getNum(billingConfig, 'cost_per_second'),
    billing_price_per_call: getNum(billingConfig, 'price_per_call'),
    billing_cost_per_call: getNum(billingConfig, 'cost_per_call'),
    billing_script: row.billing_script ?? '',
    request_script: row.request_script ?? '',
    response_script: row.response_script ?? '',
    query_script: row.query_script ?? '',
    error_script: row.error_script ?? '',
    key_pool_id: row.key_pool_id ? String(row.key_pool_id) : '',
    auth_type: row.auth_type ?? 'bearer',
    auth_param_name: row.auth_param_name ?? '',
    auth_region: row.auth_region ?? '',
    auth_service: row.auth_service ?? '',
    passthrough_headers: row.passthrough_headers ?? false,
    passthrough_body: row.passthrough_body ?? false,
    weight: String(row.weight ?? 1),
    priority: String(row.priority ?? 0),
    icon_url: row.icon_url ?? '',
    description: row.description ?? '',
    is_active: row.is_active ?? true,
  }
}

function formatBilling(channel: AdminChannel) {
  return formatBillingSummary(channel.billing_type, channel.billing_config ?? {}, 'price')
}

function formatBillingCost(channel: AdminChannel) {
  return formatBillingSummary(channel.billing_type, channel.billing_config ?? {}, 'cost')
}

export function AdminChannelsPage() {
  const { data, loading, error: loadError, reload } = useAsync(async () => {
    const [channelResponse, poolResponse] = await Promise.all([
      adminApi.listChannels(),
      adminApi.listKeyPools(),
    ])
    const rows = Array.isArray(channelResponse)
      ? channelResponse
      : channelResponse.channels ?? channelResponse.items ?? []
    const pools = Array.isArray(poolResponse) ? poolResponse : poolResponse.pools ?? []
    return { rows, pools }
  }, { rows: [] as AdminChannel[], pools: [] as AdminKeyPool[] })

  const rows = data.rows
  const pools = data.pools

  const [mutError, setMutError] = useState('')
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState<ChannelForm>(emptyForm)
  const [pendingDeleteChannel, setPendingDeleteChannel] = useState<AdminChannel | undefined>()

  const error = loadError || mutError

  const poolOptions = useMemo(
    () =>
      pools.filter((pool) =>
        form.id
          ? pool.channel_id === form.id || String(pool.channel_id) === form.key_pool_id
          : pool.channel_id === Number(form.key_pool_id || form.id || 0) || pool.channel_id === 0
      ),
    [form.id, form.key_pool_id, pools]
  )

  function openCreate() {
    setForm(emptyForm)
    setOpen(true)
    setMutError('')
  }

  function openEdit(row: AdminChannel) {
    setForm(buildFormFromChannel(row))
    setOpen(true)
    setMutError('')
  }

  function openCopy(row: AdminChannel) {
    setForm(buildFormFromChannel(row, true))
    setOpen(true)
    setMutError('')
  }

  function updatePricingGroup(index: number, patch: Partial<PricingGroupForm>) {
    setForm((current) => ({
      ...current,
      pricing_groups: current.pricing_groups.map((group, groupIndex) =>
        groupIndex === index ? { ...group, ...patch } : group
      ),
    }))
  }

  function addPricingGroup() {
    setForm((current) => ({
      ...current,
      pricing_groups: [...current.pricing_groups, createEmptyPricingGroup()],
    }))
  }

  function removePricingGroup(index: number) {
    setForm((current) => ({
      ...current,
      pricing_groups: current.pricing_groups.filter((_, groupIndex) => groupIndex !== index),
    }))
  }

  async function saveChannel() {
    setMutError('')
    try {
      const payload = {
        name: form.name.trim(),
        model: form.model.trim(),
        type: form.type,
        protocol: form.protocol,
        base_url: form.base_url.trim(),
        method: form.method,
        query_url: form.query_url.trim(),
        query_method: form.query_method,
        timeout_ms: Number(form.timeout_ms || '60000'),
        query_timeout_ms: Number(form.query_timeout_ms || '30000'),
        billing_type: form.billing_type,
        headers: parseJsonField('请求头', form.headers_text),
        billing_config: buildBillingConfig(form),
        billing_script: form.billing_script,
        request_script: form.request_script,
        response_script: form.response_script,
        query_script: form.query_script,
        error_script: form.error_script,
        key_pool_id: Number(form.key_pool_id || '0'),
        auth_type: form.auth_type,
        auth_param_name: form.auth_param_name.trim(),
        auth_region: form.auth_region.trim(),
        auth_service: form.auth_service.trim(),
        passthrough_headers: form.passthrough_headers,
        passthrough_body: form.passthrough_body,
        weight: Number(form.weight || '1'),
        priority: Number(form.priority || '0'),
        icon_url: form.icon_url.trim(),
        description: form.description.trim(),
        is_active: form.is_active,
      }
      if (form.id) {
        await adminApi.updateChannel(form.id, payload)
      } else {
        await adminApi.createChannel(payload)
      }
      setOpen(false)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  async function toggleChannel(row: AdminChannel) {
    if (!row.id) return
    setMutError('')
    try {
      await adminApi.toggleChannel(row.id, !(row.is_active ?? true))
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  async function executeDeleteChannel() {
    if (!pendingDeleteChannel?.id) return
    setMutError('')
    try {
      await adminApi.deleteChannel(pendingDeleteChannel.id)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    } finally {
      setPendingDeleteChannel(undefined)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Catalog"
        title="渠道管理"
        description="管理 API 渠道，支持认证、计费、脚本、轮询、号池和负载参数。"
        actions={
          <>
            {error ? (
              <Button size="sm" variant="outline" onClick={reload}>
                重试
              </Button>
            ) : null}
            <Button onClick={openCreate}>
              <PlusIcon data-icon="inline-start" />
              新增渠道
            </Button>
          </>
        }
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>名称</TableHead>
              <TableHead>模型</TableHead>
              <TableHead>类型</TableHead>
              <TableHead>协议</TableHead>
              <TableHead>价格摘要</TableHead>
              <TableHead>号池</TableHead>
              <TableHead>优先级/权重</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={9} />
          ) : (
            <TableBody>
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={9} className="py-10 text-center text-muted-foreground">
                    暂无渠道数据
                  </TableCell>
                </TableRow>
              ) : (
                rows.map((row, index) => (
                  <TableRow key={row.id ?? index}>
                    <TableCell className="max-w-56">
                      <div className="font-medium">{row.name ?? '未命名渠道'}</div>
                      {row.description ? (
                        <div className="line-clamp-1 text-xs text-muted-foreground">{row.description}</div>
                      ) : null}
                    </TableCell>
                    <TableCell className="max-w-48 break-all text-xs">{row.model ?? row.routing_model ?? '-'}</TableCell>
                    <TableCell>{row.type ?? '-'}</TableCell>
                    <TableCell>{row.protocol ?? 'openai'}</TableCell>
                    <TableCell className="max-w-80 align-top text-xs">
                      <div className="flex flex-col gap-1">
                        <div>
                          <span className="text-muted-foreground">售价：</span>
                          <span>{formatBilling(row)}</span>
                        </div>
                        <div>
                          <span className="text-muted-foreground">进价：</span>
                          <span>{formatBillingCost(row)}</span>
                        </div>
                        <div>
                          <span className="text-muted-foreground">分组价：</span>
                          <span>{formatGroupPricing(row)}</span>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>{row.key_pool_id ? `#${row.key_pool_id}` : '—'}</TableCell>
                    <TableCell className="text-xs">P{row.priority ?? 0} / W{row.weight ?? 1}</TableCell>
                    <TableCell>
                      <Badge variant={row.is_active === false ? 'secondary' : 'default'}>
                        {row.is_active === false ? '停用' : '启用'}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button size="sm" variant="outline" onClick={() => openEdit(row)}>编辑</Button>
                        <Button size="sm" variant="outline" onClick={() => openCopy(row)}>
                          <CopyIcon data-icon="inline-start" />
                          复制
                        </Button>
                        <Button size="sm" variant="outline" onClick={() => toggleChannel(row)}>
                          {row.is_active === false ? '启用' : '停用'}
                        </Button>
                        <Button size="sm" onClick={() => setPendingDeleteChannel(row)}>删除</Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          )}
        </Table>
        </div>
      </Card>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-[1224px]">
          <DialogHeader>
            <DialogTitle>{form.id ? '编辑渠道' : '新增渠道'}</DialogTitle>
            <DialogDescription>覆盖上游接入所需的核心字段。</DialogDescription>
          </DialogHeader>

          <Tabs defaultValue="basic">
            <TabsList className="w-full">
              <TabsTrigger value="basic">基本信息</TabsTrigger>
              <TabsTrigger value="auth">认证 &amp; 号池</TabsTrigger>
              <TabsTrigger value="billing">计费</TabsTrigger>
              <TabsTrigger value="scripts">脚本 &amp; 轮询</TabsTrigger>
            </TabsList>

            {/* ── 基本信息 ── */}
            <TabsContent value="basic" className="mt-5 max-h-[62vh] overflow-y-auto pr-1">
              <div className="grid gap-5 md:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-medium">路由名称</label>
                  <Input value={form.name} onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">标准模型名</label>
                  <Input value={form.model} onChange={(event) => setForm((current) => ({ ...current, model: event.target.value }))} />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">接口类型</label>
                  <NativeSelect value={form.type} onChange={(event) => setForm((current) => ({ ...current, type: event.target.value }))}>
                    <option value="llm">llm</option>
                    <option value="image">image</option>
                    <option value="video">video</option>
                    <option value="audio">audio</option>
                    <option value="music">music</option>
                  </NativeSelect>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">协议</label>
                  <NativeSelect value={form.protocol} onChange={(event) => setForm((current) => ({ ...current, protocol: event.target.value }))}>
                    <option value="openai">openai</option>
                    <option value="claude">claude</option>
                    <option value="gemini">gemini</option>
                  </NativeSelect>
                </div>
                <div className="space-y-2 md:col-span-2">
                  <label className="text-sm font-medium">上游 URL</label>
                  <Input value={form.base_url} onChange={(event) => setForm((current) => ({ ...current, base_url: event.target.value }))} placeholder="https://api.example.com/v1/chat/completions" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">请求方法</label>
                  <NativeSelect value={form.method} onChange={(event) => setForm((current) => ({ ...current, method: event.target.value }))}>
                    <option value="POST">POST</option>
                    <option value="GET">GET</option>
                  </NativeSelect>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">超时（ms）</label>
                  <Input value={form.timeout_ms} onChange={(event) => setForm((current) => ({ ...current, timeout_ms: event.target.value }))} />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">图标 URL</label>
                  <Input value={form.icon_url} onChange={(event) => setForm((current) => ({ ...current, icon_url: event.target.value }))} placeholder="https://…/icon.png" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">描述</label>
                  <Input value={form.description} onChange={(event) => setForm((current) => ({ ...current, description: event.target.value }))} placeholder="可选，显示在渠道名称下方" />
                </div>
                <div className="flex items-center gap-2 md:col-span-2 pt-1">
                  <input
                    id="channel-active"
                    type="checkbox"
                    checked={form.is_active}
                    onChange={(event) => setForm((current) => ({ ...current, is_active: event.target.checked }))}
                    className="h-4 w-4 rounded border-input"
                  />
                  <label htmlFor="channel-active" className="cursor-pointer text-sm font-medium">渠道启用</label>
                </div>
              </div>
            </TabsContent>

            {/* ── 认证 & 号池 ── */}
            <TabsContent value="auth" className="mt-5 max-h-[62vh] overflow-y-auto pr-1">
              <div className="grid gap-5 md:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-medium">认证方式</label>
                  <NativeSelect value={form.auth_type} onChange={(event) => setForm((current) => ({ ...current, auth_type: event.target.value }))}>
                    <option value="bearer">bearer</option>
                    <option value="query_param">query_param</option>
                    <option value="basic">basic</option>
                    <option value="sigv4">sigv4</option>
                  </NativeSelect>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Query Param 名</label>
                  <Input value={form.auth_param_name} onChange={(event) => setForm((current) => ({ ...current, auth_param_name: event.target.value }))} placeholder="如 key（query_param 认证用）" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">AWS Region</label>
                  <Input value={form.auth_region} onChange={(event) => setForm((current) => ({ ...current, auth_region: event.target.value }))} placeholder="us-east-1（sigv4 认证用）" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">AWS Service</label>
                  <Input value={form.auth_service} onChange={(event) => setForm((current) => ({ ...current, auth_service: event.target.value }))} placeholder="execute-api（sigv4 认证用）" />
                </div>

                <div className="border-t pt-4 md:col-span-2" />

                <div className="flex items-center gap-2">
                  <input
                    id="channel-passthrough-headers"
                    type="checkbox"
                    checked={form.passthrough_headers}
                    onChange={(event) => setForm((current) => ({ ...current, passthrough_headers: event.target.checked }))}
                    className="h-4 w-4 rounded border-input"
                  />
                  <label htmlFor="channel-passthrough-headers" className="cursor-pointer text-sm font-medium">透传请求头（passthrough_headers）</label>
                  <span className="text-xs text-muted-foreground">将客户端 User-Agent、Anthropic-Version 等头原样转发给上游</span>
                </div>
                <div className="flex items-center gap-2">
                  <input
                    id="channel-passthrough-body"
                    type="checkbox"
                    checked={form.passthrough_body}
                    onChange={(event) => setForm((current) => ({ ...current, passthrough_body: event.target.checked }))}
                    className="h-4 w-4 rounded border-input"
                  />
                  <label htmlFor="channel-passthrough-body" className="cursor-pointer text-sm font-medium">透传请求体（passthrough_body）</label>
                  <span className="text-xs text-muted-foreground">跳过协议转换和脚本，原样转发原始请求体（适用于签名校验场景）</span>
                </div>

                <div className="border-t pt-4 md:col-span-2" />

                <div className="space-y-2">
                  <label className="text-sm font-medium">号池绑定</label>
                  <NativeSelect value={form.key_pool_id} onChange={(event) => setForm((current) => ({ ...current, key_pool_id: event.target.value }))}>
                    <option value="">不启用</option>
                    {poolOptions.map((pool) => (
                      <option key={pool.id} value={String(pool.id)}>
                        #{pool.id} {pool.name}
                      </option>
                    ))}
                  </NativeSelect>
                </div>
                <div className="space-y-2">{/* placeholder for grid alignment */}</div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">优先级</label>
                  <Input value={form.priority} onChange={(event) => setForm((current) => ({ ...current, priority: event.target.value }))} placeholder="数值越大越优先" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">权重</label>
                  <Input value={form.weight} onChange={(event) => setForm((current) => ({ ...current, weight: event.target.value }))} placeholder="加权随机，越大被选中概率越高" />
                </div>

                <div className="border-t pt-4 md:col-span-2" />

                <div className="space-y-2 md:col-span-2">
                  <label className="text-sm font-medium">请求头（JSON）</label>
                  <p className="text-xs text-muted-foreground">固定注入到每次上游请求的 HTTP 头，如 Authorization。</p>
                  <Textarea
                    value={form.headers_text}
                    onChange={(event) => setForm((current) => ({ ...current, headers_text: event.target.value }))}
                    rows={6}
                    className="font-mono text-xs"
                  />
                </div>
              </div>
            </TabsContent>

            {/* ── 计费 ── */}
            <TabsContent value="billing" className="mt-5 max-h-[62vh] overflow-y-auto pr-1">
              <div className="grid gap-5 md:grid-cols-2">
                <div className="space-y-2 md:col-span-2">
                  <label className="text-sm font-medium">计费类型</label>
                  <NativeSelect value={form.billing_type} onChange={(event) => setForm((current) => ({ ...current, billing_type: event.target.value }))}>
                    <option value="token">token — 按 token 数计费</option>
                    <option value="image">image — 按图片张数计费</option>
                    <option value="video">video — 按视频秒数计费</option>
                    <option value="audio">audio — 按音频秒数计费</option>
                    <option value="count">count — 按调用次数计费</option>
                    <option value="custom">custom — 自定义脚本计费</option>
                  </NativeSelect>
                </div>

                {form.billing_type === 'token' && (
                  <>
                    <div className="space-y-1 md:col-span-2">
                      <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">用户侧价格</p>
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">输入价格（CNY / 百万 token）</label>
                      <Input type="number" value={form.billing_input_price} onChange={(e) => setForm((c) => ({ ...c, billing_input_price: e.target.value }))} placeholder="如 0.74" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">输出价格（CNY / 百万 token）</label>
                      <Input type="number" value={form.billing_output_price} onChange={(e) => setForm((c) => ({ ...c, billing_output_price: e.target.value }))} placeholder="如 5.9" />
                    </div>
                    <div className="space-y-1 md:col-span-2">
                      <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">成本侧价格</p>
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">输入成本（CNY / 百万 token）</label>
                      <Input type="number" value={form.billing_input_cost} onChange={(e) => setForm((c) => ({ ...c, billing_input_cost: e.target.value }))} placeholder="如 0.612" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">输出成本（CNY / 百万 token）</label>
                      <Input type="number" value={form.billing_output_cost} onChange={(e) => setForm((c) => ({ ...c, billing_output_cost: e.target.value }))} placeholder="如 4.9" />
                    </div>
                    <div className="space-y-1 md:col-span-2">
                      <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">缓存写入（留空按协议默认倍率）</p>
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">缓存写入价格（CNY / 百万 token）</label>
                      <Input type="number" value={form.billing_cache_create_price} onChange={(e) => setForm((c) => ({ ...c, billing_cache_create_price: e.target.value }))} placeholder="留空按协议默认" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">缓存写入成本（CNY / 百万 token）</label>
                      <Input type="number" value={form.billing_cache_create_cost} onChange={(e) => setForm((c) => ({ ...c, billing_cache_create_cost: e.target.value }))} placeholder="留空按协议默认" />
                    </div>
                    <div className="space-y-1 md:col-span-2">
                      <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">缓存（留空按协议默认倍率）</p>
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">缓存读取价格（CNY / 百万 token）</label>
                      <Input type="number" value={form.billing_cache_read_price} onChange={(e) => setForm((c) => ({ ...c, billing_cache_read_price: e.target.value }))} placeholder="留空按协议默认" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">缓存读取成本（CNY / 百万 token）</label>
                      <Input type="number" value={form.billing_cache_read_cost} onChange={(e) => setForm((c) => ({ ...c, billing_cache_read_cost: e.target.value }))} placeholder="留空按协议默认" />
                    </div>
                    <div className="flex items-center gap-2 md:col-span-2">
                      <input
                        id="input-from-response"
                        type="checkbox"
                        checked={form.billing_input_from_response}
                        onChange={(e) => setForm((c) => ({ ...c, billing_input_from_response: e.target.checked }))}
                        className="h-4 w-4 rounded border-input"
                      />
                      <label htmlFor="input-from-response" className="cursor-pointer text-sm font-medium">
                        从响应中获取实际输入 token 数（input_from_response）
                      </label>
                    </div>
                  </>
                )}

                {form.billing_type === 'image' && (
                  <>
                    <div className="space-y-1 md:col-span-2">
                      <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">按档位定价（size_prices / size_costs）</p>
                      <p className="text-xs text-muted-foreground">填写后按 1k/2k/3k/4k 档位优先计费；留空则使用基础价格 + resolution_tiers。</p>
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">1k 售价（CNY / 张）</label>
                      <Input type="number" value={form.billing_size_price_1k} onChange={(e) => setForm((c) => ({ ...c, billing_size_price_1k: e.target.value }))} placeholder="如 3" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">1k 进价（CNY / 张）</label>
                      <Input type="number" value={form.billing_size_cost_1k} onChange={(e) => setForm((c) => ({ ...c, billing_size_cost_1k: e.target.value }))} placeholder="如 2.5" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">2k 售价（CNY / 张）</label>
                      <Input type="number" value={form.billing_size_price_2k} onChange={(e) => setForm((c) => ({ ...c, billing_size_price_2k: e.target.value }))} placeholder="如 5" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">2k 进价（CNY / 张）</label>
                      <Input type="number" value={form.billing_size_cost_2k} onChange={(e) => setForm((c) => ({ ...c, billing_size_cost_2k: e.target.value }))} placeholder="如 4.2" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">3k 售价（CNY / 张）</label>
                      <Input type="number" value={form.billing_size_price_3k} onChange={(e) => setForm((c) => ({ ...c, billing_size_price_3k: e.target.value }))} placeholder="如 7" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">3k 进价（CNY / 张）</label>
                      <Input type="number" value={form.billing_size_cost_3k} onChange={(e) => setForm((c) => ({ ...c, billing_size_cost_3k: e.target.value }))} placeholder="如 6.2" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">4k 售价（CNY / 张）</label>
                      <Input type="number" value={form.billing_size_price_4k} onChange={(e) => setForm((c) => ({ ...c, billing_size_price_4k: e.target.value }))} placeholder="如 9" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">4k 进价（CNY / 张）</label>
                      <Input type="number" value={form.billing_size_cost_4k} onChange={(e) => setForm((c) => ({ ...c, billing_size_cost_4k: e.target.value }))} placeholder="如 7.8" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">兜底尺寸售价（CNY）</label>
                      <Input type="number" value={form.billing_default_size_price} onChange={(e) => setForm((c) => ({ ...c, billing_default_size_price: e.target.value }))} placeholder="size 不在表中时使用" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">兜底尺寸进价（CNY）</label>
                      <Input type="number" value={form.billing_default_size_cost} onChange={(e) => setForm((c) => ({ ...c, billing_default_size_cost: e.target.value }))} placeholder="size 不在表中时使用" />
                    </div>
                    <div className="space-y-1 md:col-span-2">
                      <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">基础价格（档位留空时生效）</p>
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">基础价格（CNY）</label>
                      <Input type="number" value={form.billing_base_price} onChange={(e) => setForm((c) => ({ ...c, billing_base_price: e.target.value }))} placeholder="如 5" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">基础进价（CNY）</label>
                      <Input type="number" value={form.billing_base_cost} onChange={(e) => setForm((c) => ({ ...c, billing_base_cost: e.target.value }))} placeholder="如 4.2" />
                    </div>
                  </>
                )}

                {(form.billing_type === 'video' || form.billing_type === 'audio') && (
                  <>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">售价（CNY / 秒）</label>
                      <Input type="number" value={form.billing_price_per_second} onChange={(e) => setForm((c) => ({ ...c, billing_price_per_second: e.target.value }))} placeholder="如 0.01" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">进价（CNY / 秒）</label>
                      <Input type="number" value={form.billing_cost_per_second} onChange={(e) => setForm((c) => ({ ...c, billing_cost_per_second: e.target.value }))} placeholder="如 0.008" />
                    </div>
                  </>
                )}

                {form.billing_type === 'count' && (
                  <>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">售价（CNY / 次）</label>
                      <Input type="number" value={form.billing_price_per_call} onChange={(e) => setForm((c) => ({ ...c, billing_price_per_call: e.target.value }))} placeholder="如 0.001" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-medium">进价（CNY / 次）</label>
                      <Input type="number" value={form.billing_cost_per_call} onChange={(e) => setForm((c) => ({ ...c, billing_cost_per_call: e.target.value }))} placeholder="如 0.0008" />
                    </div>
                  </>
                )}

                {form.billing_type !== 'custom' ? (
                  <div className="space-y-3 md:col-span-2">
                    <div className="flex items-center justify-between gap-3">
                      <div className="space-y-1">
                        <label className="text-sm font-medium">分组定价（pricing_groups）</label>
                        <p className="text-xs text-muted-foreground">按用户分组覆盖默认售价，不影响上游进价。组名必须与用户 group 完全一致，例如 vip、premium。</p>
                      </div>
                      <Button type="button" size="sm" variant="outline" onClick={addPricingGroup}>
                        <PlusIcon data-icon="inline-start" />
                        新增分组
                      </Button>
                    </div>

                    {form.pricing_groups.length === 0 ? (
                      <div className="rounded-lg border border-dashed px-4 py-5 text-xs text-muted-foreground">
                        暂无分组定价。需要 VIP 或其他用户组差异化售价时，点击“新增分组”进行配置。
                      </div>
                    ) : (
                      <div className="flex flex-col gap-4">
                        {form.pricing_groups.map((group, index) => (
                          <Card key={`${group.group || 'group'}-${index}`} size="sm">
                            <CardHeader>
                              <div className="flex items-start justify-between gap-3">
                                <div className="space-y-1">
                                  <CardTitle className="text-base">分组覆盖 #{index + 1}</CardTitle>
                                  <CardDescription>仅覆盖用户侧售价字段，保存后会写入 billing_config.pricing_groups。</CardDescription>
                                </div>
                                <Button type="button" size="sm" variant="outline" onClick={() => removePricingGroup(index)}>
                                  移除
                                </Button>
                              </div>
                            </CardHeader>
                            <CardContent className="grid gap-5 md:grid-cols-2">
                              <div className="space-y-2 md:col-span-2">
                                <label className="text-sm font-medium">用户分组名</label>
                                <Input
                                  value={group.group}
                                  onChange={(event) => updatePricingGroup(index, { group: event.target.value })}
                                  placeholder="如 vip"
                                />
                              </div>

                              {form.billing_type === 'token' ? (
                                <>
                                  <div className="space-y-2">
                                    <label className="text-sm font-medium">输入售价（CNY / 百万 token）</label>
                                    <Input
                                      type="number"
                                      value={group.input_price_per_1m_tokens}
                                      onChange={(event) => updatePricingGroup(index, { input_price_per_1m_tokens: event.target.value })}
                                      placeholder="如 8"
                                    />
                                  </div>
                                  <div className="space-y-2">
                                    <label className="text-sm font-medium">输出售价（CNY / 百万 token）</label>
                                    <Input
                                      type="number"
                                      value={group.output_price_per_1m_tokens}
                                      onChange={(event) => updatePricingGroup(index, { output_price_per_1m_tokens: event.target.value })}
                                      placeholder="如 32"
                                    />
                                  </div>
                                </>
                              ) : null}

                              {form.billing_type === 'image' ? (
                                <>
                                  <div className="space-y-2">
                                    <label className="text-sm font-medium">基础售价（CNY）</label>
                                    <Input
                                      type="number"
                                      value={group.base_price}
                                      onChange={(event) => updatePricingGroup(index, { base_price: event.target.value })}
                                      placeholder="如 5"
                                    />
                                  </div>
                                  <div className="space-y-2">
                                    <label className="text-sm font-medium">兜底尺寸售价（CNY）</label>
                                    <Input
                                      type="number"
                                      value={group.default_size_price}
                                      onChange={(event) => updatePricingGroup(index, { default_size_price: event.target.value })}
                                      placeholder="如 5"
                                    />
                                  </div>
                                  <div className="space-y-2">
                                    <label className="text-sm font-medium">1k 售价（CNY / 张）</label>
                                    <Input
                                      type="number"
                                      value={group.size_price_1k}
                                      onChange={(event) => updatePricingGroup(index, { size_price_1k: event.target.value })}
                                      placeholder="如 3"
                                    />
                                  </div>
                                  <div className="space-y-2">
                                    <label className="text-sm font-medium">2k 售价（CNY / 张）</label>
                                    <Input
                                      type="number"
                                      value={group.size_price_2k}
                                      onChange={(event) => updatePricingGroup(index, { size_price_2k: event.target.value })}
                                      placeholder="如 5"
                                    />
                                  </div>
                                  <div className="space-y-2">
                                    <label className="text-sm font-medium">3k 售价（CNY / 张）</label>
                                    <Input
                                      type="number"
                                      value={group.size_price_3k}
                                      onChange={(event) => updatePricingGroup(index, { size_price_3k: event.target.value })}
                                      placeholder="如 7"
                                    />
                                  </div>
                                  <div className="space-y-2">
                                    <label className="text-sm font-medium">4k 售价（CNY / 张）</label>
                                    <Input
                                      type="number"
                                      value={group.size_price_4k}
                                      onChange={(event) => updatePricingGroup(index, { size_price_4k: event.target.value })}
                                      placeholder="如 9"
                                    />
                                  </div>
                                </>
                              ) : null}

                              {(form.billing_type === 'video' || form.billing_type === 'audio') ? (
                                <div className="space-y-2 md:col-span-2">
                                  <label className="text-sm font-medium">售价（CNY / 秒）</label>
                                  <Input
                                    type="number"
                                    value={group.price_per_second}
                                    onChange={(event) => updatePricingGroup(index, { price_per_second: event.target.value })}
                                    placeholder="如 0.006"
                                  />
                                </div>
                              ) : null}

                              {form.billing_type === 'count' ? (
                                <div className="space-y-2 md:col-span-2">
                                  <label className="text-sm font-medium">售价（CNY / 次）</label>
                                  <Input
                                    type="number"
                                    value={group.price_per_call}
                                    onChange={(event) => updatePricingGroup(index, { price_per_call: event.target.value })}
                                    placeholder="如 0.001"
                                  />
                                </div>
                              ) : null}
                            </CardContent>
                          </Card>
                        ))}
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="rounded-lg border border-dashed px-4 py-5 text-xs text-muted-foreground md:col-span-2">
                    `custom` 计费依赖脚本返回 credits 数值，当前不支持通过 pricing_groups 做可视化分组定价覆盖。
                  </div>
                )}

                <div className="space-y-2 md:col-span-2">
                  <label className="text-sm font-medium">高级配置（JSON）</label>
                  <p className="text-xs text-muted-foreground">用于配置 metric_paths、resolution_tiers 等高级参数；结构化价格字段和分组定价会在保存时覆盖同名键。</p>
                  <Textarea
                    value={form.billing_config_text}
                    onChange={(event) => setForm((current) => ({ ...current, billing_config_text: event.target.value }))}
                    rows={8}
                    className="font-mono text-xs"
                    placeholder={'{\n  "metric_paths": {\n    "input_tokens": "response.usage.prompt_tokens",\n    "output_tokens": "response.usage.completion_tokens"\n  },\n  "resolution_tiers": [\n    { "max_pixels": 1048576, "multiplier": 1.0 }\n  ]\n}'}
                  />
                  <p className="text-xs text-muted-foreground">分组定价已改为上方可视化编辑；这里建议只放 metric_paths、resolution_tiers 等非价格字段。</p>
                </div>

                <div className="space-y-2 md:col-span-2">
                  <label className="text-sm font-medium">自定义计费脚本</label>
                  <p className="text-xs text-muted-foreground">billing_type=custom 时生效，脚本需返回 credits 数值。</p>
                  <Textarea
                    value={form.billing_script}
                    onChange={(event) => setForm((current) => ({ ...current, billing_script: event.target.value }))}
                    rows={8}
                    className="font-mono text-xs"
                    placeholder="function calcBilling(request) { return 1000 }"
                  />
                </div>
              </div>
            </TabsContent>

            {/* ── 脚本 & 轮询 ── */}
            <TabsContent value="scripts" className="mt-5 max-h-[62vh] overflow-y-auto pr-1">
              <div className="grid gap-5">
                <div className="space-y-2">
                  <label className="text-sm font-medium">入参脚本</label>
                  <p className="text-xs text-muted-foreground">mapRequest(input) → 将平台请求映射为上游格式。</p>
                  <Textarea value={form.request_script} onChange={(event) => setForm((current) => ({ ...current, request_script: event.target.value }))} rows={7} className="font-mono text-xs" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">出参脚本</label>
                  <p className="text-xs text-muted-foreground">mapResponse(input) → 映射上游响应，或提取 upstream_task_id（异步）。</p>
                  <Textarea value={form.response_script} onChange={(event) => setForm((current) => ({ ...current, response_script: event.target.value }))} rows={7} className="font-mono text-xs" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">错误检测脚本</label>
                  <p className="text-xs text-muted-foreground">checkError(response) → 返回非空字符串表示错误，null/false 表示正常。</p>
                  <Textarea value={form.error_script} onChange={(event) => setForm((current) => ({ ...current, error_script: event.target.value }))} rows={5} className="font-mono text-xs" />
                </div>

                <div className="border-t pt-2">
                  <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">轮询配置（异步任务用）</p>
                </div>

                <div className="space-y-2">
                  <label className="text-sm font-medium">轮询 URL</label>
                  <Input value={form.query_url} onChange={(event) => setForm((current) => ({ ...current, query_url: event.target.value }))} placeholder="如 https://api.example.com/v1/tasks/{id}" />
                </div>
                <div className="grid grid-cols-2 gap-5">
                  <div className="space-y-2">
                    <label className="text-sm font-medium">轮询方法</label>
                    <NativeSelect value={form.query_method} onChange={(event) => setForm((current) => ({ ...current, query_method: event.target.value }))}>
                      <option value="GET">GET</option>
                      <option value="POST">POST</option>
                    </NativeSelect>
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium">轮询超时（ms）</label>
                    <Input value={form.query_timeout_ms} onChange={(event) => setForm((current) => ({ ...current, query_timeout_ms: event.target.value }))} />
                  </div>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">轮询脚本</label>
                  <p className="text-xs text-muted-foreground">mapResponse(input) → 将轮询响应映射为标准格式。</p>
                  <Textarea value={form.query_script} onChange={(event) => setForm((current) => ({ ...current, query_script: event.target.value }))} rows={7} className="font-mono text-xs" />
                </div>
              </div>
            </TabsContent>
          </Tabs>

          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>取消</Button>
            <Button onClick={saveChannel} disabled={!form.name.trim() || !form.model.trim() || !form.base_url.trim()}>
              <SaveIcon data-icon="inline-start" />
              保存
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={pendingDeleteChannel !== undefined} onOpenChange={() => setPendingDeleteChannel(undefined)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确认删除渠道"{pendingDeleteChannel?.name ?? pendingDeleteChannel?.model ?? pendingDeleteChannel?.id}"吗？此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={executeDeleteChannel}>删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
