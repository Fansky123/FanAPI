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
        <el-form-item label="计费配置（JSON）">
          <el-input v-model="form.billingConfigStr" type="textarea" :rows="6"
            placeholder='{"input_price_per_1k_tokens": 15000, ...}' style="font-family:monospace" />
        </el-form-item>
        <el-form-item label="入参映射脚本">
          <el-input v-model="form.request_script" type="textarea" :rows="8" placeholder="package main&#10;&#10;func MapRequest(input map[string]interface{}) map[string]interface{} {&#10;    return input&#10;}" style="font-family:monospace;font-size:.82rem" />
        </el-form-item>
        <el-form-item label="出参映射脚本">
          <el-input v-model="form.response_script" type="textarea" :rows="8" placeholder="package main&#10;&#10;func MapResponse(input map[string]interface{}) map[string]interface{} {&#10;    return input&#10;}" style="font-family:monospace;font-size:.82rem" />
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
  request_script: '', response_script: '', is_active: true
})
const form = reactive(emptyForm())

onMounted(fetchChannels)

async function fetchChannels() {
  const res = await channelApi.list()
  channels.value = res.channels ?? []
}

function openDialog(row = null) {
  editRow.value = row
  if (row) {
    Object.assign(form, {
      ...row,
      headersStr: JSON.stringify(row.headers ?? {}, null, 2),
      billingConfigStr: JSON.stringify(row.billing_config ?? {}, null, 2),
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
    billingConfig = JSON.parse(form.billingConfigStr || '{}')
  } catch {
    return ElMessage.error('JSON 格式错误，请检查请求头或计费配置')
  }

  const payload = {
    name: form.name, model: form.model, type: form.type,
    base_url: form.base_url, method: form.method, headers,
    timeout_ms: form.timeout_ms, billing_type: form.billing_type,
    billing_config: billingConfig, request_script: form.request_script,
    response_script: form.response_script, is_active: form.is_active,
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
</script>
