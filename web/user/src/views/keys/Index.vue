<template>
  <div class="keys-page">
    <div class="page-header">
      <div class="page-header-left">
        <div class="page-title">API 密钥</div>
        <div class="page-desc">密钥用于调用 API，请妥善保管。密钥泄露时请立即撤销。</div>
      </div>
      <el-button type="primary" size="default" @click="showCreate = true" style="border-radius: 6px;">
        <el-icon style="margin-right:4px"><Plus /></el-icon>
        新建密钥
      </el-button>
    </div>

    <div class="api-base-tip" v-if="site.apiBase">
      <el-icon style="color:#165dff"><InfoFilled /></el-icon>
      <span>接口调用地址：</span>
      <code class="base-code">{{ site.apiBase }}</code>
      <el-button link size="small" @click="copyText(site.apiBase)">复制</el-button>
    </div>

    <div class="keys-list-card">
      <el-table :data="keys" stripe empty-text="暂无密钥，点击右上角创建">
        <el-table-column label="名称" min-width="140">
          <template #default="{ row }">
            <span class="key-name">{{ row.name || '未命名' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="密钥" min-width="320">
          <template #default="{ row }">
            <div class="key-val-row">
              <code class="key-code">{{ row.viewable ? row.raw_key : row.key_prefix + '...' }}</code>
              <el-button v-if="row.viewable" link size="small" style="color:#165dff" @click="copyText(row.raw_key)">复制</el-button>
              <span v-else class="key-hidden-tip">历史密钥不可查看</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.key_type === 'stable' ? 'warning' : 'info'"
              size="small"
              effect="light"
            >
              {{ row.key_type === 'stable' ? '稳定' : '低价' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <div class="status-badge" :class="row.is_active ? 'active' : 'revoked'">
              <span class="status-dot"></span>
              {{ row.is_active ? '启用' : '已撤销' }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            <span class="time-text">{{ formatTime(row.created_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" align="center">
          <template #default="{ row }">
            <el-popconfirm
              title="此操作将永久删除该密钥，无法恢复，确认继续？"
              confirm-button-text="永久删除"
              cancel-button-text="取消"
              confirm-button-type="danger"
              @confirm="revokeKey(row.id)"
            >
              <template #reference>
                <el-button type="danger" size="small" link>删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="showCreate" title="创建 API 密钥" width="480px" :close-on-click-modal="false">
      <el-form :model="createForm" label-position="top">
        <el-form-item label="备注名称">
          <el-input v-model="createForm.name" placeholder="如：我的项目 / 测试用" clearable @keyup.enter="doCreate" />
        </el-form-item>
        <el-form-item label="密钥类型">
          <div class="key-type-row">
            <div
              class="key-type-card"
              :class="{ selected: createForm.keyType === 'low_price' }"
              @click="createForm.keyType = 'low_price'"
            >
              <div class="key-type-title">💰 低价密钥</div>
              <div class="key-type-desc">按优先级和权重路由，价格更优惠</div>
            </div>
            <div
              class="key-type-card"
              :class="{ selected: createForm.keyType === 'stable' }"
              @click="createForm.keyType = 'stable'"
            >
              <div class="key-type-title">🛡️ 稳定密钥</div>
              <div class="key-type-desc">同模型先试售价最低的渠道，失败或超时自动切换到更贵的渠道重试</div>
            </div>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="doCreate">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showNew" title="密钥已创建" width="520px" :close-on-click-modal="false">
      <div class="new-key-box">
        <code class="new-key-code">{{ newKey }}</code>
        <el-button type="primary" size="small" @click="copyText(newKey)" style="border-radius:4px">复制</el-button>
      </div>
      <template #footer>
        <el-button type="primary" @click="showNew = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import { userApi } from '@/api'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import { ElMessage } from 'element-plus'
import { Plus, InfoFilled } from '@element-plus/icons-vue'

const store = useUserStore()
const site = useSiteStore()
const keys = ref([])
const showCreate = ref(false)
const showNew = ref(false)
const newKey = ref('')
const creating = ref(false)
const createForm = reactive({ name: '', keyType: 'low_price' })

onMounted(fetchKeys)

async function fetchKeys() {
  try {
    const res = await userApi.listAPIKeys()
    keys.value = res.api_keys ?? []
  } catch {}
}

async function doCreate() {
  if (!createForm.name.trim()) {
    ElMessage.warning('请输入备注名称')
    return
  }
  creating.value = true
  try {
    const res = await userApi.createAPIKey(createForm.name, createForm.keyType)
    newKey.value = res.key
    showCreate.value = false
    showNew.value = true
    createForm.name = ''
    createForm.keyType = 'low_price'
    fetchKeys()
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || '创建失败')
  } finally {
    creating.value = false
  }
}

async function revokeKey(id) {
  try {
    await userApi.deleteAPIKey(id)
    ElMessage.success('密钥已永久删除')
    fetchKeys()
  } catch {
    ElMessage.error('删除失败')
  }
}

function copyText(text) {
  navigator.clipboard.writeText(text).then(() => {
    ElMessage({ message: '已复制', type: 'success', duration: 1500 })
  })
}

function formatTime(v) {
  if (!v) return '—'
  return new Date(v).toLocaleString('zh-CN', { hour12: false })
}
</script>

<style scoped>
.keys-page {
  padding-bottom: 40px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}
.page-title {
  font-size: 24px;
  font-weight: 600;
  color: rgb(26, 27, 28);
  line-height: 32px;
}
.page-desc {
  font-size: 13px;
  color: rgb(134, 147, 171);
  margin-top: 4px;
}
.api-base-tip {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgb(240, 244, 255);
  border: 1px solid #d0e2ff;
  border-radius: 6px;
  padding: 10px 16px;
  font-size: 13px;
  color: #4e5969;
}
.base-code {
  font-family: monospace;
  background: white;
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid #d0e2ff;
  color: #165dff;
  font-size: 13px;
}
.keys-list-card {
  background: white;
  border-radius: 8px;
  border: 1px solid #e5e6eb;
  overflow: hidden;
}
.key-name { font-weight: 500; color: #1d2129; }
.key-val-row { display: flex; align-items: center; gap: 8px; }
.key-code {
  font-family: monospace;
  font-size: 12px;
  background: #f5f7fa;
  padding: 3px 8px;
  border-radius: 4px;
  color: #1d2129;
  border: 1px solid #e5e6eb;
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
}
.key-hidden-tip { font-size: 12px; color: #c9cdd4; }
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  padding: 2px 8px;
  border-radius: 999px;
}
.status-badge.active { color: #00b42a; background: #e8f7e8; }
.status-badge.revoked { color: #86909c; background: #f5f7fa; }
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  display: inline-block;
}
.time-text { font-size: 13px; color: #86909c; }
.dim { color: #c9cdd4; }
.usage-card {
  background: white;
  border-radius: 8px;
  border: 1px solid #e5e6eb;
  padding: 20px;
}
.usage-title {
  font-size: 16px;
  font-weight: 600;
  color: #1d2129;
  margin-bottom: 14px;
}
.usage-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 0;
  border-bottom: 1px dashed #f0f1f5;
  font-size: 13px;
}
.usage-item:last-child { border: none; }
.usage-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}
.usage-badge.llm { background: #f0f4ff; color: #165dff; }
.usage-badge.img { background: #e8f7e8; color: #00b42a; }
code { font-family: monospace; font-size: 12px; color: #4e5969; }
.new-key-box {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #f5f7fa;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  padding: 12px 14px;
}
.new-key-code {
  font-family: monospace;
  font-size: 13px;
  color: #165dff;
  flex: 1;
  word-break: break-all;
}

/* 密钥类型选择卡片 */
.key-type-row {
  display: flex;
  gap: 12px;
  width: 100%;
}
.key-type-card {
  flex: 1;
  border: 2px solid #e5e6eb;
  border-radius: 8px;
  padding: 14px 16px;
  cursor: pointer;
  transition: border-color .15s, background .15s;
  background: white;
}
.key-type-card:hover {
  border-color: #165dff;
  background: #f0f4ff;
}
.key-type-card.selected {
  border-color: #165dff;
  background: #f0f4ff;
}
.key-type-title {
  font-size: 15px;
  font-weight: 600;
  color: #1d2129;
  margin-bottom: 6px;
}
.key-type-desc {
  font-size: 12px;
  color: #86909c;
  line-height: 1.6;
}
</style>
