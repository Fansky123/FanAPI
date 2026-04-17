<template>
  <div class="dashboard">
    <!-- 邀请链接卡片 -->
    <el-card class="invite-card" shadow="never">
      <div class="invite-header">
        <div>
          <h3 class="card-title">我的邀请链接</h3>
          <p class="card-sub">新用户通过此链接注册后将归入您名下，您可对其进行充值管理。</p>
        </div>
        <el-tag type="success" size="large" class="user-count-tag">
          已邀请 {{ total }} 人
        </el-tag>
      </div>

      <div class="link-row">
        <el-input v-model="inviteUrl" readonly class="link-input">
          <template #prepend><el-icon><Link /></el-icon></template>
        </el-input>
        <el-button type="primary" :icon="CopyDocument" @click="copyLink">复制链接</el-button>
      </div>
    </el-card>

    <!-- 已邀请用户列表 -->
    <el-card shadow="never" style="margin-top:20px">
      <div class="section-header">
        <h3 class="card-title">我邀请的用户</h3>
        <span class="hint-text">按余额从低到高排列，余额不足 ¥1 的用户以红色标出</span>
      </div>

      <el-table
        :data="users"
        stripe
        :row-class-name="rowClass"
        empty-text="暂无邀请用户"
        style="width:100%;margin-top:12px"
      >
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="username" label="用户名" min-width="130" />
        <el-table-column label="邮箱" min-width="180">
          <template #default="{ row }">{{ row.email || '—' }}</template>
        </el-table-column>
        <el-table-column label="当前余额" width="140" align="right">
          <template #default="{ row }">
            <span :class="{ 'low': row.balance < 1000000 }">
              ¥{{ (row.balance / 1e6).toFixed(4) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="累计充值" width="130" align="right">
          <template #default="{ row }">
            ¥{{ (row.total_recharge / 1e6).toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column label="累计消费" width="130" align="right">
          <template #default="{ row }">
            ¥{{ (row.total_spend / 1e6).toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column label="充值" width="90" align="center">
          <template #default="{ row }">
            <el-button size="small" type="success" @click="openRecharge(row)">充值</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="total > pageSize"
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="prev, pager, next"
        style="margin-top:16px;justify-content:flex-end"
        @current-change="fetchUsers"
      />
    </el-card>

    <!-- 充值弹窗 -->
    <el-dialog v-model="showRecharge" title="为用户充值" width="400px" :close-on-click-modal="false">
      <div class="recharge-body">
        <div class="recharge-who">
          <el-icon><User /></el-icon>
          <span>{{ rechargeUser?.username }}</span>
          <el-tag size="small" :type="(rechargeUser?.balance ?? 0) < 1000000 ? 'danger' : 'info'">
            余额 ¥{{ ((rechargeUser?.balance ?? 0) / 1e6).toFixed(4) }}
          </el-tag>
        </div>

        <el-form label-position="top" style="margin-top:16px">
          <el-form-item label="充值金额（积分 credits）">
            <el-input-number
              v-model="rechargeAmount"
              :min="1"
              :step="1000000"
              :precision="0"
              style="width:100%"
            />
            <div class="credit-hint">
              {{ rechargeAmount.toLocaleString() }} credits ≈ ¥{{ (rechargeAmount / 1e6).toFixed(4) }}
            </div>
          </el-form-item>
          <div class="quick-btns">
            <el-button size="small" @click="rechargeAmount = 1000000">¥1</el-button>
            <el-button size="small" @click="rechargeAmount = 10000000">¥10</el-button>
            <el-button size="small" @click="rechargeAmount = 100000000">¥100</el-button>
          </div>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="showRecharge = false">取消</el-button>
        <el-button type="primary" :loading="recharging" @click="doRecharge">确认充值</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Link, CopyDocument, User } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { agentUserApi, agentInviteApi } from '@/api/agent'

const inviteCode = ref('')
const users = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 50

const inviteUrl = computed(() =>
  inviteCode.value
    ? `${location.origin}/register?ref=${inviteCode.value}`
    : '加载中...'
)

const showRecharge = ref(false)
const rechargeUser = ref(null)
const rechargeAmount = ref(1000000)
const recharging = ref(false)

onMounted(async () => {
  await Promise.all([fetchInvite(), fetchUsers()])
})

async function fetchInvite() {
  try {
    const res = await agentInviteApi.get()
    inviteCode.value = res.invite_code ?? ''
  } catch (e) {
    ElMessage.error('获取邀请码失败')
  }
}

async function fetchUsers() {
  try {
    const res = await agentUserApi.list(page.value, pageSize)
    users.value = res.users ?? []
    total.value = res.total ?? 0
  } catch (e) {
    ElMessage.error('获取用户列表失败')
  }
}

async function copyLink() {
  try {
    await navigator.clipboard.writeText(inviteUrl.value)
    ElMessage.success('链接已复制')
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}

function openRecharge(user) {
  rechargeUser.value = user
  rechargeAmount.value = 1000000
  showRecharge.value = true
}

async function doRecharge() {
  recharging.value = true
  try {
    await agentUserApi.recharge(rechargeUser.value.id, rechargeAmount.value)
    ElMessage.success(`已为 ${rechargeUser.value.username} 充值 ${rechargeAmount.value.toLocaleString()} credits`)
    showRecharge.value = false
    fetchUsers()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error ?? '充值失败')
  } finally {
    recharging.value = false
  }
}

function rowClass({ row }) {
  return row.balance < 1000000 ? 'row-low' : ''
}
</script>

<style scoped>
.dashboard {
  animation: fadeIn .3s ease;
}
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to   { opacity: 1; transform: translateY(0); }
}

.invite-card { border-radius: 12px; }
.invite-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
}
.card-title {
  margin: 0 0 4px;
  font-size: 16px;
  font-weight: 700;
  color: #1d2129;
}
.card-sub {
  margin: 0;
  font-size: 13px;
  color: #8a8f9d;
}
.user-count-tag { font-size: 14px; padding: 6px 14px; }

.link-row {
  display: flex;
  gap: 10px;
}
.link-input { flex: 1; }

.section-header {
  display: flex;
  align-items: baseline;
  gap: 14px;
}
.hint-text {
  font-size: 12px;
  color: #adb0bc;
}

.low { color: #f56c6c; font-weight: 600; }
:deep(.row-low td) { background: #fff5f5 !important; }

.recharge-body { padding: 4px 0 0; }
.recharge-who {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #1d2129;
}
.credit-hint {
  margin-top: 6px;
  font-size: 12px;
  color: #8a8f9d;
}
.quick-btns { display: flex; gap: 8px; }
</style>
