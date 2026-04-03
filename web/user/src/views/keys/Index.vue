<template>
  <div class="key-page">
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">Access Keys</div>
          <h3>管理 API 密钥与调用身份</h3>
          <p>每个密钥可按项目独立管理，可长期查看密钥标识，并按需吊销。</p>
        </div>
        <el-button type="primary" @click="showCreate = true">
          <el-icon><Plus /></el-icon> 创建密钥
        </el-button>
      </div>
    </el-card>

    <el-card>
      <el-table :data="keys" stripe>
      <el-table-column label="名称" prop="name" />
      <el-table-column label="密钥标识" width="320">
        <template #default="{ row }">
          <code>kid_{{ row.key_prefix }}</code>
        </template>
      </el-table-column>
        <el-table-column label="完整密钥" min-width="360">
          <template #default="{ row }">
            <div v-if="row.viewable" class="key-cell">
              <code class="full-key">{{ row.raw_key }}</code>
              <el-button size="small" link @click="copyKey(row.raw_key)">复制</el-button>
            </div>
            <span v-else class="key-missing">历史密钥未保存明文，无法恢复，请重新生成</span>
          </template>
        </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
            {{ row.is_active ? '生效中' : '已吊销' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" prop="created_at" :formatter="fmtTime" />
      <el-table-column label="操作" width="120" align="center">
        <template #default="{ row }">
          <el-popconfirm v-if="row.is_active" title="确认吊销此密钥？" @confirm="deleteKey(row.id)">
            <template #reference>
              <el-button type="danger" size="small" link>吊销</el-button>
            </template>
          </el-popconfirm>
          <span v-else style="color:#c0c4cc">-</span>
        </template>
      </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建弹窗 -->
    <el-dialog v-model="showCreate" title="创建 API 密钥" width="400px">
      <el-form :model="createForm" label-position="top">
        <el-form-item label="密钥备注名">
          <el-input v-model="createForm.name" placeholder="如：我的项目" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" @click="doCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 显示新密钥（只显示一次） -->
    <el-dialog v-model="showNew" title="密钥已创建" width="480px" :close-on-click-modal="false">
      <el-alert type="warning" :closable="false" style="margin-bottom:16px">
        完整密钥只显示一次，请立即保存。后续可在列表查看密钥标识（不可反推完整密钥）。
      </el-alert>
      <el-input :value="newKey" readonly>
        <template #append>
          <el-button @click="copyNew">复制</el-button>
        </template>
      </el-input>
      <template #footer>
        <el-button type="primary" @click="showNew = false">我已保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import { userApi } from '@/api'
import { ElMessage } from 'element-plus'

const keys = ref([])
const showCreate = ref(false)
const showNew = ref(false)
const newKey = ref('')
const createForm = reactive({ name: '' })

onMounted(fetchKeys)

async function fetchKeys() {
  const res = await userApi.listAPIKeys()
  keys.value = res.api_keys ?? []
}

async function doCreate() {
  const res = await userApi.createAPIKey(createForm.name)
  newKey.value = res.key
  showCreate.value = false
  showNew.value = true
  createForm.name = ''
  fetchKeys()
}

async function deleteKey(id) {
  await userApi.deleteAPIKey(id)
  ElMessage.success('密钥已吊销')
  fetchKeys()
}

function copyNew() {
  navigator.clipboard.writeText(newKey.value)
  ElMessage.success('已复制')
}

function copyKey(value) {
	navigator.clipboard.writeText(value)
	ElMessage.success('密钥已复制')
}

function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}
</script>

<style scoped>
.key-page {
  max-width: 1320px;
}
.hero-card {
  margin-bottom: 16px;
}
.hero-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.eyebrow {
  color: #1e66ff;
  font-size: .82rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: .08em;
}
.hero-row h3 {
  margin: 8px 0 10px;
  font-size: 1.55rem;
}
.hero-row p {
  margin: 0;
  color: #617086;
}
code {
  background: #f1f5fb;
  border: 1px solid #dde7f6;
  padding: 3px 8px;
  border-radius: 6px;
}
.key-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.full-key {
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.key-missing {
  color: #98a2b3;
  font-size: .82rem;
}

@media (max-width: 900px) {
  .hero-row {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
