<template>
  <div>
    <!-- 操作栏 -->
    <div class="action-bar">
      <span class="action-hint">密钥用于调用 API，请妥善保管。</span>
      <el-button type="primary" size="small" @click="showCreate = true">
        <el-icon><Plus /></el-icon> 创建密钥
      </el-button>
    </div>

    <el-card>
      <el-table :data="keys" stripe>
        <el-table-column label="名称" prop="name" min-width="120" />
        <el-table-column label="密钥标识" width="220">
          <template #default="{ row }">
            <code class="code-tag">{{ row.key_prefix }}…</code>
          </template>
        </el-table-column>
        <el-table-column label="完整密钥" min-width="300">
          <template #default="{ row }">
            <div v-if="row.viewable" class="key-cell">
              <code class="code-tag key-full">{{ row.raw_key }}</code>
              <el-button size="small" link @click="copyKey(row.raw_key)">复制</el-button>
            </div>
            <span v-else class="dim-text">历史密钥不可查看，请重新生成</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
              {{ row.is_active ? '启用' : '已撤销' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170" :formatter="(r,c,v) => v ? new Date(v).toLocaleString('zh-CN',{hour12:false}) : '-'" prop="created_at" />
        <el-table-column label="操作" width="90" align="center">
          <template #default="{ row }">
            <el-popconfirm v-if="row.is_active" title="确认撤销此密钥？" @confirm="deleteKey(row.id)">
              <template #reference>
                <el-button type="danger" size="small" link>撤销</el-button>
              </template>
            </el-popconfirm>
            <span v-else class="dim-text">—</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建弹窗 -->
    <el-dialog v-model="showCreate" title="创建 API 密钥" width="400px">
      <el-form :model="createForm" label-position="top">
        <el-form-item label="备注名称">
          <el-input v-model="createForm.name" placeholder="如：我的项目" @keyup.enter="doCreate" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" @click="doCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 显示新密钥（只显示一次）-->
    <el-dialog v-model="showNew" title="密钥已创建" width="480px" :close-on-click-modal="false">
      <el-alert type="warning" :closable="false" style="margin-bottom:14px" title="完整密钥只显示一次，请立即保存！" />
      <el-input :value="newKey" readonly>
        <template #append>
          <el-button @click="copyNew">复制</el-button>
        </template>
      </el-input>
      <template #footer>
        <el-button type="primary" @click="showNew = false">已保存</el-button>
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
  if (!createForm.name.trim()) return
  const res = await userApi.createAPIKey(createForm.name)
  newKey.value = res.key
  showCreate.value = false
  showNew.value = true
  createForm.name = ''
  fetchKeys()
}

async function deleteKey(id) {
  await userApi.deleteAPIKey(id)
  ElMessage.success('密钥已撤销')
  fetchKeys()
}

function copyNew() {
  navigator.clipboard.writeText(newKey.value)
  ElMessage.success('已复制')
}

function copyKey(value) {
  navigator.clipboard.writeText(value)
  ElMessage.success('已复制')
}
</script>

<style scoped>
.action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.action-hint { color: #8a94a8; font-size: .82rem; }
.code-tag {
  background: #f0f2f7;
  padding: 2px 8px;
  border-radius: 6px;
  font-family: monospace;
  font-size: .82rem;
}
.key-cell { display: flex; align-items: center; gap: 8px; }
.key-full { max-width: 240px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: inline-block; }
.dim-text { color: #b0b8cc; font-size: .8rem; }
</style>
