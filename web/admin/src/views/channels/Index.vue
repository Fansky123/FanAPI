<template>
  <div>
    <div style="margin-bottom:16px;display:flex;justify-content:flex-end">
      <el-button type="primary" @click="openDialog()">
        <el-icon><Plus /></el-icon> 新增渠道
      </el-button>
    </div>

    <el-table :data="channels" stripe border>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="渠道名称" />
      <el-table-column prop="model" label="模型" />
      <el-table-column prop="type" label="类型" width="90">
        <template #default="{ row }">
          <el-tag size="small">{{ row.type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="billing_type" label="计费类型" width="100" />
      <el-table-column label="售价" width="160">
        <template #default="{ row }">
          <span style="font-size:12px">{{ formatPrice(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="进价" width="160">
        <template #default="{ row }">
          <span style="font-size:12px;color:#999">{{ formatCost(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="is_active" label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-switch v-model="row.is_active" @change="toggleActive(row)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" align="center">
        <template #default="{ row }">
          <el-button size="small" @click="openDialog(row)">编辑</el-button>
          <el-popconfirm title="确认删除？" @confirm="deleteRow(row.id)">
            <template #reference>
              <el-button size="small" type="danger">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="editRow ? '编辑渠道' : '新增渠道'" width="760px" top="5vh">
      <el-form :model="form" label-width="120px" style="max-height:70vh;overflow-y:auto">
        <el-form-item label="渠道名称" required>
          <el-input v-model="form.name" placeholder="如：nano-1001（用户可见）" />
        </el-form-item>
        <el-form-item label="标准模型名" required>
          <el-input v-model="form.model" placeholder="如：nano-banana-pro（用于前端分组）" />
        </el-form-item>
        <el-form-item label="接口类型" required>
          <el-select v-model="form.type" style="width:100%">
            <el-option label="LLM 对话" value="llm" />
            <el-option label="图片生成" value="image" />
            <el-option label="视频生成" value="video" />
            <el-option label="音频生成" value="audio" />
          </el-select>
        </el-form-item>
        <el-form-item label="上游 URL" required>
          <el-input v-model="form.base_url" placeholder="https://api.example.com/v1/..." />
        </el-form-item>
        <el-form-item label="请求方法">
          <el-select v-model="form.method" style="width:100px">
            <el-option label="POST" value="POST" />
            <el-option label="GET" value="GET" />
          </el-select>
        </el-form-item>
        <el-form-item label="请求头（JSON）">
          <el-input v-model="form.headersStr" type="textarea" :rows="3"
            placeholder='{"Authorization": "Bearer YOUR_KEY"}' style="font-family:monospace" />
        </el-form-item>
        <el-form-item label="超时（ms）">
          <el-input-number v-model="form.timeout_ms" :min="1000" :step="1000" />
        </el-form-item>
        <el-form-item label="计费类型" required>
          <el-select v-model="form.billing_type" style="width:100%">
            <el-option label="token 计费" value="token" />
            <el-option label="图片计费" value="image" />
            <el-option label="视频计费" value="video" />
            <el-option label="音频计费" value="audio" />
            <el-option label="按次计费" value="count" />
            <el-option label="自定义脚本" value="custom" />
          </el-select>
        </el-form-item>

        <!-- ===== Token 计费价格 ===== -->
        <template v-if="form.billing_type === 'token'">
          <el-divider content-position="left" style="margin:8px 0 12px">
            <span style="font-size:13px;color:#666">Token 价格（单位：credits / 1M tokens）</span>
          </el-divider>
          <el-form-item label="售价 · 输入">
            <el-input-number v-model="form.bp.input_price_per_1m_tokens" :min="0" :step="100000" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">用户被扣费</span>
          </el-form-item>
          <el-form-item label="售价 · 输出">
            <el-input-number v-model="form.bp.output_price_per_1m_tokens" :min="0" :step="100000" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">用户被扣费</span>
          </el-form-item>
          <el-form-item label="进价 · 输入">
            <el-input-number v-model="form.bp.input_cost_per_1m_tokens" :min="0" :step="100000" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">支付给上游（成本）</span>
          </el-form-item>
          <el-form-item label="进价 · 输出">
            <el-input-number v-model="form.bp.output_cost_per_1m_tokens" :min="0" :step="100000" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">支付给上游（成本）</span>
          </el-form-item>
          <el-form-item label="输入从响应取">
            <el-switch v-model="form.bp.input_from_response" />
            <span style="margin-left:8px;color:#999;font-size:12px">开启后输入 token 数从响应 usage 字段读取</span>
          </el-form-item>
        </template>

        <!-- ===== 图片计费价格 ===== -->
        <template v-if="form.billing_type === 'image'">
          <el-divider content-position="left" style="margin:8px 0 12px">
            <span style="font-size:13px;color:#666">图片价格（单位：credits / 张·基准分辨率）</span>
          </el-divider>
          <el-form-item label="售价 · 基础">
            <el-input-number v-model="form.bp.base_price" :min="0" :step="1000" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">用户被扣费</span>
          </el-form-item>
          <el-form-item label="进价 · 基础">
            <el-input-number v-model="form.bp.base_cost" :min="0" :step="1000" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">支付给上游（成本）</span>
          </el-form-item>
        </template>

        <!-- ===== 视频 / 音频计费价格 ===== -->
        <template v-if="form.billing_type === 'video' || form.billing_type === 'audio'">
          <el-divider content-position="left" style="margin:8px 0 12px">
            <span style="font-size:13px;color:#666">{{ form.billing_type === 'video' ? '视频' : '音频' }}价格（单位：credits / 秒）</span>
          </el-divider>
          <el-form-item label="售价 · 每秒">
            <el-input-number v-model="form.bp.price_per_second" :min="0" :step="100" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">用户被扣费</span>
          </el-form-item>
          <el-form-item label="进价 · 每秒">
            <el-input-number v-model="form.bp.cost_per_second" :min="0" :step="100" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">支付给上游（成本）</span>
          </el-form-item>
        </template>

        <!-- ===== 按次计费价格 ===== -->
        <template v-if="form.billing_type === 'count'">
          <el-divider content-position="left" style="margin:8px 0 12px">
            <span style="font-size:13px;color:#666">按次价格（单位：credits / 次）</span>
          </el-divider>
          <el-form-item label="售价 · 每次">
            <el-input-number v-model="form.bp.price_per_call" :min="0" :step="1000" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">用户被扣费</span>
          </el-form-item>
          <el-form-item label="进价 · 每次">
            <el-input-number v-model="form.bp.cost_per_call" :min="0" :step="1000" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">支付给上游（成本）</span>
          </el-form-item>
        </template>

        <el-form-item label="高级配置（JSON）">
          <el-input v-model="form.billingConfigStr" type="textarea" :rows="4"
            placeholder="其余计费参数，如 metric_paths、resolution_tiers 等" style="font-family:monospace;font-size:12px" />
          <div style="font-size:11px;color:#aaa;margin-top:4px">上方价格字段优先级更高，保存时自动合并到此 JSON</div>
        </el-form-item>
        <el-form-item label="入参映射脚本">
          <el-input v-model="form.request_script" type="textarea" :rows="8" placeholder="package main&#10;&#10;func MapRequest(input map[string]interface{}) map[string]interface{} {&#10;    return input&#10;}" style="font-family:monospace;font-size:.82rem" />
        </el-form-item>
        <el-form-item label="出参映射脚本">
          <el-input v-model="form.response_script" type="textarea" :rows="8" placeholder="package main&#10;&#10;// 同步接口：返回 {&quot;code&quot;:200,&quot;url&quot;:&quot;...&quot;,&quot;status&quot;:2}&#10;// 异步接口：返回 {&quot;upstream_task_id&quot;:&quot;xxx&quot;} 即可触发轮询&#10;func MapResponse(input map[string]interface{}) map[string]interface{} {&#10;    return input&#10;}" style="font-family:monospace;font-size:.82rem" />
        </el-form-item>

        <el-divider content-position="left" style="margin:8px 0 12px">
          <span style="font-size:13px;color:#666">异步轮询配置（视频 / 音频等异步接口使用）</span>
        </el-divider>
        <el-form-item label="轮询 URL">
          <el-input v-model="form.query_url" placeholder="https://api.example.com/v1/tasks/{id}（{id} 为第三方任务 ID 占位符）" />
        </el-form-item>
        <el-form-item label="轮询方法">
          <el-select v-model="form.query_method" style="width:100px">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
          </el-select>
        </el-form-item>
        <el-form-item label="轮询映射脚本">
          <el-input v-model="form.query_script" type="textarea" :rows="8" placeholder="package main&#10;&#10;// 将第三方轮询响应映射为标准格式&#10;// status: 2=成功 3=失败 其他=进行中&#10;func MapResponse(input map[string]interface{}) map[string]interface{} {&#10;    return input&#10;}" style="font-family:monospace;font-size:.82rem" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveChannel">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import { channelApi } from '@/api'
import { ElMessage } from 'element-plus'

const channels = ref([])
const dialogVisible = ref(false)
const editRow = ref(null)

const emptyForm = () => ({
  name: '', model: '', type: 'llm', base_url: '', method: 'POST',
  headersStr: '{}', timeout_ms: 30000,
  billing_type: 'token', billingConfigStr: '{}',
  request_script: '', response_script: '',
  query_url: '', query_method: 'GET', query_script: '',
  is_active: true,
  bp: emptyBp(),
})

function emptyBp() {
  return {
    // token
    input_price_per_1m_tokens: 0, output_price_per_1m_tokens: 0,
    input_cost_per_1m_tokens: 0, output_cost_per_1m_tokens: 0,
    input_from_response: false,
    // image
    base_price: 0, base_cost: 0,
    // video / audio
    price_per_second: 0, cost_per_second: 0,
    // count
    price_per_call: 0, cost_per_call: 0,
  }
}

// 从 billing_config JSON 中提取结构化价格字段
function extractBp(cfg) {
  const bp = emptyBp()
  const keys = Object.keys(bp)
  for (const k of keys) {
    if (cfg[k] !== undefined) bp[k] = cfg[k]
  }
  return bp
}

// 将结构化价格字段合并回 billing_config（过滤零值的进价/售价避免污染 JSON）
function mergeBpToConfig(bp, baseConfigStr) {
  let cfg = {}
  try { cfg = JSON.parse(baseConfigStr || '{}') } catch { cfg = {} }
  for (const [k, v] of Object.entries(bp)) {
    if (v !== 0 && v !== false) cfg[k] = v
    else if (cfg[k] !== undefined) cfg[k] = v // 已有字段置为 0/false 也要保留
  }
  return cfg
}
const form = reactive(emptyForm())

onMounted(fetchChannels)

async function fetchChannels() {
  const res = await channelApi.list()
  channels.value = res.channels ?? []
}

function openDialog(row = null) {
  editRow.value = row
  if (row) {
    const cfg = row.billing_config ?? {}
    // 将 billing_config 中的价格字段提到 bp，剩余字段留在 billingConfigStr
    const bp = extractBp(cfg)
    const bpKeys = new Set(Object.keys(emptyBp()))
    const rest = Object.fromEntries(Object.entries(cfg).filter(([k]) => !bpKeys.has(k)))
    Object.assign(form, {
      ...row,
      headersStr: JSON.stringify(row.headers ?? {}, null, 2),
      billingConfigStr: JSON.stringify(rest, null, 2),
      bp,
    })
  } else {
    Object.assign(form, emptyForm())
  }
  dialogVisible.value = true
}

async function saveChannel() {
  let headers, billingConfig
  try {
    headers = JSON.parse(form.headersStr || '{}')
    billingConfig = mergeBpToConfig(form.bp, form.billingConfigStr)
  } catch {
    return ElMessage.error('JSON 格式错误，请检查请求头或高级配置')
  }

  const payload = {
    name: form.name, model: form.model, type: form.type,
    base_url: form.base_url, method: form.method, headers,
    timeout_ms: form.timeout_ms, billing_type: form.billing_type,
    billing_config: billingConfig, request_script: form.request_script,
    response_script: form.response_script,
    query_url: form.query_url, query_method: form.query_method, query_script: form.query_script,
    is_active: form.is_active,
  }

  if (editRow.value) {
    await channelApi.update(editRow.value.id, payload)
  } else {
    await channelApi.create(payload)
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetchChannels()
}

async function deleteRow(id) {
  await channelApi.delete(id)
  ElMessage.success('已删除')
  fetchChannels()
}

async function toggleActive(row) {
  await channelApi.update(row.id, { is_active: row.is_active })
}

function formatPrice(row) {
  const c = row.billing_config ?? {}
  if (row.billing_type === 'token') {
    return `输入 ${c.input_price_per_1m_tokens ?? 0} / 输出 ${c.output_price_per_1m_tokens ?? 0}`
  }
  if (row.billing_type === 'image') return `基础 ${c.base_price ?? 0}`
  if (row.billing_type === 'video' || row.billing_type === 'audio') return `${c.price_per_second ?? 0} /秒`
  if (row.billing_type === 'count') return `${c.price_per_call ?? 0} /次`
  return '—'
}

function formatCost(row) {
  const c = row.billing_config ?? {}
  if (row.billing_type === 'token') {
    return `输入 ${c.input_cost_per_1m_tokens ?? 0} / 输出 ${c.output_cost_per_1m_tokens ?? 0}`
  }
  if (row.billing_type === 'image') return `基础 ${c.base_cost ?? 0}`
  if (row.billing_type === 'video' || row.billing_type === 'audio') return `${c.cost_per_second ?? 0} /秒`
  if (row.billing_type === 'count') return `${c.cost_per_call ?? 0} /次`
  return '—'
}
</script>
