<template>
  <div class="invite-page">
    <!-- 邀请链接卡片 -->
    <el-card class="invite-card">
      <div class="invite-header">
        <div>
          <div class="eyebrow">Invite Link</div>
          <h3>我的邀请链接</h3>
          <p>将链接发送给客户，通过链接注册的用户将自动绑定到您的名下，您可以在下方查看他们的充值和消费明细。</p>
        </div>
      </div>

      <div class="link-row" v-if="inviteCode">
        <el-input :value="inviteLink" readonly class="link-input" />
        <el-button type="primary" @click="copyLink">复制链接</el-button>
      </div>
      <el-skeleton v-else :rows="1" animated />
    </el-card>

    <!-- 微信二维码设置 -->
    <el-card class="qr-card">
      <h4>我的微信二维码</h4>
      <p class="tip">设置您的微信二维码后，通过您的邀请链接注册并登录的用户将自动看到此二维码。</p>

      <div class="qr-area">
        <div class="qr-preview" v-if="wechatQR">
          <img :src="wechatQR" alt="微信二维码" />
        </div>
        <div class="qr-placeholder" v-else>
          <el-icon size="40"><Picture /></el-icon>
          <span>暂未设置</span>
        </div>
        <div class="qr-actions">
          <el-button @click="triggerQRUpload">从本地上传</el-button>
          <el-input v-model="qrInput" placeholder="或粘贴图片 URL" clearable style="flex:1" />
          <el-button type="primary" @click="saveQR">保存二维码</el-button>
        </div>
        <input ref="qrFileRef" type="file" accept="image/*" style="display:none" @change="onQRFile" />
      </div>
    </el-card>

    <!-- 邀请用户统计 -->
    <el-card>
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:16px">
        <h4 style="margin:0">邀请用户统计</h4>
        <span style="color:#909399;font-size:.85rem">共 {{ total }} 位用户</span>
      </div>

      <el-table :data="stats" stripe border>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="email" label="邮箱" min-width="150" />
        <el-table-column label="余额（¥）" width="120">
          <template #default="{ row }">¥{{ (row.balance / 1e6).toFixed(4) }}</template>
        </el-table-column>
        <el-table-column label="累计充值" width="130">
          <template #default="{ row }">
            <span style="color:#10b981;font-weight:600">¥{{ (row.total_recharge / 1e6).toFixed(4) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="累计消费" width="130">
          <template #default="{ row }">
            <span style="color:#f59e0b;font-weight:600">¥{{ (row.total_spend / 1e6).toFixed(4) }}</span>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        :page-size="50"
        :total="total"
        style="margin-top:16px"
        @current-change="fetchStats"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { agentInviteApi } from '@/api/agent'
import { ElMessage } from 'element-plus'
import { Picture } from '@element-plus/icons-vue'

const inviteCode = ref('')
const wechatQR = ref('')
const qrInput = ref('')
const qrFileRef = ref(null)
const stats = ref([])
const page = ref(1)
const total = ref(0)

const inviteLink = computed(() => {
  if (!inviteCode.value) return ''
  return `${location.origin}/register?ref=${inviteCode.value}`
})

onMounted(async () => {
  const res = await agentInviteApi.get()
  inviteCode.value = res.invite_code || ''
  wechatQR.value = res.wechat_qr || ''
  qrInput.value = wechatQR.value.startsWith('data:') ? '' : wechatQR.value
  fetchStats()
})

async function fetchStats() {
  const res = await agentInviteApi.stats(page.value)
  stats.value = res.stats ?? []
  total.value = res.total ?? 0
}

function copyLink() {
  navigator.clipboard.writeText(inviteLink.value)
    .then(() => ElMessage.success('链接已复制'))
    .catch(() => {
      const el = document.createElement('input')
      el.value = inviteLink.value
      document.body.appendChild(el)
      el.select()
      document.execCommand('copy')
      document.body.removeChild(el)
      ElMessage.success('链接已复制')
    })
}

function triggerQRUpload() {
  qrFileRef.value?.click()
}

function onQRFile(e) {
  const file = e.target.files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = (ev) => {
    wechatQR.value = ev.target.result
    qrInput.value = ''
  }
  reader.readAsDataURL(file)
}

async function saveQR() {
  const val = qrInput.value.trim() || wechatQR.value
  if (!val) return ElMessage.warning('请先上传或输入二维码')
  await agentInviteApi.updateWechatQR(val)
  wechatQR.value = val
  ElMessage.success('微信二维码已保存')
}
</script>

<style scoped>
.invite-page { max-width: 960px; display: flex; flex-direction: column; gap: 20px; }
.invite-card h3, .qr-card h4 { margin: 0 0 8px; }
.eyebrow { color:#10b981;font-size:.82rem;font-weight:700;text-transform:uppercase;letter-spacing:.08em;margin-bottom:6px; }
.invite-header { margin-bottom: 16px; }
.invite-header p { margin:0;color:#617086;font-size:.9rem;line-height:1.6; }
.link-row { display: flex; gap: 10px; align-items: center; }
.link-input :deep(.el-input__wrapper) { background: #f7f8fc; }
.tip { color: #909399; font-size: .85rem; margin: 4px 0 16px; }
.qr-area { display: flex; flex-direction: column; gap: 14px; }
.qr-preview { display: flex; justify-content: center; }
.qr-preview img { width: 180px; height: 180px; object-fit: contain; border: 1px solid #e5e7eb; border-radius: 10px; padding: 8px; }
.qr-placeholder {
  width: 180px; height: 180px;
  border: 2px dashed #d1d5db;
  border-radius: 10px;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  color: #9ca3af; gap: 8px;
}
.qr-actions { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; }
</style>
