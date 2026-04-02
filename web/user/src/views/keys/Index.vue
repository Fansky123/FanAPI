<template>
  <div>
    <div style="margin-bottom:16px;display:flex;justify-content:space-between;align-items:center">
      <span style="color:#606266">每个密钥可独立使用，支持按需吊销</span>
      <el-button type="primary" @click="showCreate = true">
        <el-icon><Plus /></el-icon> 创建密钥
      </el-button>
    </div>

    <el-table :data="keys" stripe>
      <el-table-column label="名称" prop="name" />
      <el-table-column label="密钥（前缀）" width="280">
        <template #default="{ row }">
          <code>sk-{{ row.key_prefix }}...</code>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" prop="created_at" :formatter="fmtTime" />
      <el-table-column label="操作" width="120" align="center">
        <template #default="{ row }">
          <el-popconfirm title="确认吊销此密钥？" @confirm="deleteKey(row.id)">
            <template #reference>
              <el-button type="danger" size="small" link>吊销</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

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
        此密钥只显示一次，请立即保存！
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
  keys.value = res.keys ?? []
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

function fmtTime(row, col, val) {
  return val ? new Date(val).toLocaleString('zh-CN') : '-'
}
</script>
