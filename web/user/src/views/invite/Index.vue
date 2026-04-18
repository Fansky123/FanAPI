<template>
  <div class="invite-page">
    <el-card class="invite-card" shadow="never">
      <template #header>
        <span class="card-title">邀请中心</span>
      </template>

      <div v-if="loading" class="loading-wrap">
        <el-skeleton :rows="4" animated />
      </div>

      <template v-else>
        <!-- 邀请码 -->
        <div class="stat-row">
          <div class="stat-item">
            <div class="stat-label">我的邀请码</div>
            <div class="stat-value code-value">
              {{ info.invite_code }}
              <el-button size="small" link @click="copyCode">
                <el-icon><CopyDocument /></el-icon>
              </el-button>
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-label">已邀请人数</div>
            <div class="stat-value">{{ info.invite_count }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">冻结返佣积分</div>
            <div class="stat-value highlight">{{ formatCredits(info.frozen_balance) }}</div>
          </div>
        </div>

        <!-- 邀请链接 -->
        <el-divider />
        <div class="link-section">
          <div class="section-label">邀请链接</div>
          <div class="link-row">
            <el-input :value="inviteLink" readonly class="link-input" />
            <el-button type="primary" @click="copyLink">复制链接</el-button>
          </div>
          <div class="form-tip">将此链接分享给好友，对方通过该链接注册后即成为您的邀请用户</div>
        </div>

        <!-- 解冻操作 -->
        <el-divider />
        <div class="convert-section">
          <div class="section-label">解冻积分</div>
          <p class="form-tip">将冻结返佣积分兑换为可用积分，当前可解冻：<b>{{ formatCredits(info.frozen_balance) }}</b></p>
          <div class="convert-row">
            <el-input-number
              v-model="convertAmount"
              :min="0"
              :max="info.frozen_balance"
              :step="100"
              placeholder="0 表示全部解冻"
              style="width:200px"
            />
            <el-button type="primary" :loading="converting" @click="doConvert" :disabled="info.frozen_balance <= 0">
              解冻
            </el-button>
          </div>
        </div>

        <!-- 说明 -->
        <el-divider />
        <el-alert type="info" :closable="false" show-icon>
          <template #title>返佣规则</template>
          <template #default>
            <ul class="rule-list">
              <li>您邀请的用户每次消费，系统将按比例将返佣积分冻结至您的账户</li>
              <li>冻结积分可随时解冻为可用积分</li>
              <li>具体返佣比例以平台设置为准，请咨询客服</li>
            </ul>
          </template>
        </el-alert>
      </template>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'
import { userApi } from '@/api'

const loading = ref(true)
const converting = ref(false)
const convertAmount = ref(0)
const info = ref({ invite_code: '', invite_count: 0, frozen_balance: 0 })

const inviteLink = computed(() => {
  const base = window.location.origin
  return `${base}/register?ref=${info.value.invite_code}`
})

function formatCredits(v) {
  if (!v) return '0'
  return (v / 1e6).toFixed(4) + ' 积分'
}

async function copyCode() {
  try {
    await navigator.clipboard.writeText(info.value.invite_code)
    ElMessage.success('邀请码已复制')
  } catch {
    ElMessage.error('复制失败')
  }
}

async function copyLink() {
  try {
    await navigator.clipboard.writeText(inviteLink.value)
    ElMessage.success('邀请链接已复制')
  } catch {
    ElMessage.error('复制失败')
  }
}

async function doConvert() {
  converting.value = true
  try {
    const res = await userApi.convertFrozen(convertAmount.value)
    ElMessage.success(`成功解冻 ${formatCredits(res.converted)}`)
    info.value.frozen_balance = 0
    convertAmount.value = 0
    // 刷新
    await fetchInfo()
  } catch {
    // 错误已由拦截器展示
  } finally {
    converting.value = false
  }
}

async function fetchInfo() {
  try {
    const res = await userApi.getInviteInfo()
    info.value = res
  } catch {
    // 忽略
  }
}

onMounted(async () => {
  await fetchInfo()
  loading.value = false
})
</script>

<style scoped>
.invite-page {
  max-width: 700px;
}
.invite-card {
  border-radius: 12px;
}
.card-title {
  font-size: 16px;
  font-weight: 600;
}
.loading-wrap {
  padding: 20px 0;
}
.stat-row {
  display: flex;
  gap: 32px;
  flex-wrap: wrap;
}
.stat-item {
  flex: 1;
  min-width: 140px;
  background: #f5f7fa;
  border-radius: 10px;
  padding: 16px 20px;
}
.stat-label {
  font-size: 13px;
  color: #86909c;
  margin-bottom: 8px;
}
.stat-value {
  font-size: 22px;
  font-weight: 700;
  color: #1d2129;
  display: flex;
  align-items: center;
  gap: 6px;
}
.stat-value.highlight {
  color: #165dff;
}
.code-value {
  font-size: 18px;
  letter-spacing: 0.05em;
}
.section-label {
  font-size: 14px;
  font-weight: 600;
  color: #1d2129;
  margin-bottom: 10px;
}
.link-section,
.convert-section {
  margin: 4px 0;
}
.link-row {
  display: flex;
  gap: 10px;
  align-items: center;
}
.link-input {
  flex: 1;
}
.convert-row {
  display: flex;
  gap: 10px;
  align-items: center;
  margin-top: 8px;
}
.form-tip {
  font-size: 13px;
  color: #86909c;
  margin-top: 6px;
  line-height: 1.6;
}
.rule-list {
  margin: 6px 0 0 16px;
  padding: 0;
  font-size: 13px;
  color: #4e5969;
  line-height: 1.8;
}
</style>
