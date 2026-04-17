<template>
  <div class="auth-page">
    <div class="auth-card">
      <!-- Logo -->
      <div class="auth-logo">
        <div class="logo-icon">{{ site.siteName.charAt(0).toUpperCase() }}</div>
        <span class="logo-name">{{ site.siteName }}</span>
      </div>

      <!-- 模式切换 -->
      <div v-if="site.wechatMPLoginEnabled" class="mode-toggle">
        <button :class="['mode-btn', mode === 'password' ? 'active' : '']" @click="switchMode('password')">
          账号密码
        </button>
        <button :class="['mode-btn', mode === 'qr' ? 'active' : '']" @click="switchMode('qr')">
          <span class="wechat-dot">●</span> 微信扫码
        </button>
      </div>

      <!-- 账号密码注册 -->
      <template v-if="mode === 'password'">
        <h2 class="auth-title">创建账户</h2>
        <p class="auth-sub">填写以下信息完成注册</p>

        <el-form :model="form" @submit.prevent="handleRegister" label-position="top" class="auth-form">
          <el-form-item label="用户名">
            <el-input v-model="form.username" placeholder="3-32 个字符" clearable size="large" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="form.password" type="password" show-password placeholder="至少 8 位" size="large" />
          </el-form-item>
          <el-form-item v-if="form.invite_code" label="邀请码">
            <el-input v-model="form.invite_code" size="large" readonly>
              <template #prefix><el-icon><Link /></el-icon></template>
            </el-input>
            <div class="form-hint">通过邀请链接注册，注册后可获得专属客服支持</div>
          </el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" class="submit-btn">
            完成注册
          </el-button>
        </el-form>

        <div class="hint-tip">
          <el-icon><InfoFilled /></el-icon>
          注册后可在账户设置中绑定邮箱，以便通过邮箱找回密码。
        </div>
      </template>

      <!-- 微信公众号扫码注册/登录 -->
      <template v-else>
        <h2 class="auth-title">微信扫码注册</h2>
        <p class="auth-sub">扫描二维码，关注公众号即可完成注册并登录</p>

        <div class="qr-section">
          <div v-if="qrLoading" class="qr-placeholder">
            <el-skeleton :rows="0" animated style="width:200px;height:200px;border-radius:8px" />
            <p class="qr-status">正在生成二维码…</p>
          </div>

          <template v-else-if="qrImg && qrStatus !== 'expired'">
            <div class="qr-img-wrap">
              <img :src="qrImg" class="qr-img" alt="微信注册二维码" />
              <div v-if="qrStatus === 'scanned'" class="qr-scanned-overlay">
                <el-icon class="check-icon"><CircleCheckFilled /></el-icon>
                <span>注册成功，正在跳转…</span>
              </div>
            </div>
            <p class="qr-status">
              <span v-if="qrStatus === 'pending'">
                <span class="pulse-dot" />请用微信扫描二维码（{{ countdown }}s 后过期）
              </span>
              <span v-else style="color:#07c160">注册成功，正在跳转…</span>
            </p>
          </template>

          <template v-else>
            <div class="qr-expired-wrap">
              <el-button type="primary" plain round @click="loadQRCode">
                <el-icon><RefreshRight /></el-icon> 刷新二维码
              </el-button>
            </div>
            <p class="qr-status" style="color:#f56c6c">二维码已过期，请刷新</p>
          </template>
        </div>

        <div class="qr-tip">
          <el-icon><InfoFilled /></el-icon>
          首次扫码将自动创建账户，已有账号直接登录
        </div>
      </template>

      <div class="link-row">
        已有账号？<router-link to="/login" class="link-a">立即登录</router-link>
      </div>
    </div>
  </div>

  <!-- 注册成功客服二维码弹窗 -->
  <el-dialog v-model="showQR" title="专属客服" width="320px" :close-on-click-modal="false" align-center>
    <div class="qr-dialog">
      <p>🎉 注册成功！扫码添加您的专属客服微信获取使用帮助。</p>
      <img :src="inviterQR" alt="客服微信" class="qr-dialog-img" />
    </div>
    <template #footer>
      <el-button type="primary" @click="closeQRAndGo">进入首页</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive, ref, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import { authApi } from '@/api'
import { ElMessage } from 'element-plus'
import { InfoFilled, Link, CircleCheckFilled, RefreshRight } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const store = useUserStore()
const site = useSiteStore()
const loading = ref(false)
const showQR = ref(false)
const inviterQR = ref('')
const form = reactive({ username: '', password: '', invite_code: '' })

onMounted(() => {
  if (route.query.ref) form.invite_code = route.query.ref
})

// 广告追踪参数（从落地页 URL 读取，随注册请求一起发送）
function getTrackParams() {
  const q = route.query
  const p = {}
  if (q.bd_vid) p.bd_vid = q.bd_vid
  if (q.qh_click_id) p.qh_click_id = q.qh_click_id
  if (q.source_id) p.source_id = q.source_id
  if (q.ocpc_id) p.platform_id = parseInt(q.ocpc_id, 10)
  return p
}

// ── 模式切换 ──────────────────────────────────
const mode = ref('password')

function switchMode(m) {
  if (mode.value === m) return
  stopPoll()
  mode.value = m
  if (m === 'qr') loadQRCode()
}

// ── 账号密码注册 ──────────────────────────────
async function handleRegister() {
  if (!form.username || form.username.length < 3) return ElMessage.warning('用户名至少 3 个字符')
  if (!form.password || form.password.length < 8) return ElMessage.warning('密码至少 8 位')
  loading.value = true
  try {
    const payload = { username: form.username, password: form.password, ...getTrackParams() }
    if (form.invite_code) payload.invite_code = form.invite_code
    const res = await authApi.register(payload)
    store.setToken(res.token)
    store.setUsername(res.user?.username || form.username)
    ElMessage.success('注册成功，欢迎！')
    if (res.inviter_wechat_qr) {
      inviterQR.value = res.inviter_wechat_qr
      showQR.value = true
    } else {
      router.push('/models')
    }
  } finally {
    loading.value = false
  }
}

function closeQRAndGo() {
  showQR.value = false
  router.push('/models')
}

// ── 公众号扫码注册 ────────────────────────────
const qrLoading = ref(false)
const qrImg = ref('')
const qrUUID = ref('')
const qrStatus = ref('pending')
const countdown = ref(600)
let pollTimer = null
let countdownTimer = null

async function loadQRCode() {
  stopPoll()
  qrLoading.value = true
  qrImg.value = ''
  qrStatus.value = 'pending'
  countdown.value = 600
  try {
    // 将广告追踪参数传给 QR 接口，扫码后关联平台账户
    const params = { ...getTrackParams() }
    if (form.invite_code) params.source_id = form.invite_code
    const res = await authApi.wechatMPQRCode(params)
    qrUUID.value = res.uuid
    qrImg.value = 'data:image/png;base64,' + res.qr_img
    startPoll()
    startCountdown()
  } catch {
    ElMessage.error('获取二维码失败，请稍后重试')
    qrStatus.value = 'expired'
  } finally {
    qrLoading.value = false
  }
}

function startPoll() {
  pollTimer = setInterval(async () => {
    if (!qrUUID.value) return
    try {
      const res = await authApi.wechatMPPoll(qrUUID.value)
      if (res.status === 'success') {
        stopPoll()
        qrStatus.value = 'scanned'
        store.setToken(res.token)
        await store.fetchProfile()
        ElMessage.success('微信注册/登录成功，欢迎！')
        router.push('/models')
      } else if (res.status === 'expired') {
        stopPoll()
        qrStatus.value = 'expired'
      }
    } catch {
      stopPoll()
      qrStatus.value = 'expired'
    }
  }, 2000)
}

function startCountdown() {
  countdownTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(countdownTimer)
      qrStatus.value = 'expired'
      stopPoll()
    }
  }, 1000)
}

function stopPoll() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null }
}

onUnmounted(stopPoll)
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex; align-items: center; justify-content: center;
  background: linear-gradient(126.82deg, rgba(236,243,255,.8) 0.58%, rgba(232,247,251,.8) 86.28%), #f2f3f5;
  padding: 24px;
}
.auth-card {
  width: 100%; max-width: 400px;
  background: #fff; border: 1px solid #e5e6eb; border-radius: 8px;
  padding: 36px 32px; box-shadow: 0 4px 20px rgba(0,0,0,.06);
}
.auth-logo {
  display: flex; align-items: center; gap: 10px;
  margin-bottom: 20px; justify-content: center;
}
.logo-icon {
  width: 32px; height: 32px; border-radius: 8px; background: #165dff; color: #fff;
  display: grid; place-items: center; font-weight: 700; font-size: 15px;
}
.logo-name { font-size: 17px; font-weight: 700; color: #1d2129; }

.mode-toggle {
  display: flex; margin-bottom: 20px;
  border: 1px solid #dde1e7; border-radius: 8px; overflow: hidden;
}
.mode-btn {
  flex: 1; padding: 9px 0; border: none; background: #f5f7fa;
  cursor: pointer; font-size: 13px; color: #4e5969; transition: all .18s;
  display: flex; align-items: center; justify-content: center; gap: 5px;
}
.mode-btn.active { background: #165dff; color: #fff; font-weight: 600; }
.mode-btn:not(.active):hover { background: #eef2ff; }
.wechat-dot { color: #07c160; font-size: 9px; }
.mode-btn.active .wechat-dot { color: #a7f3d0; }

.auth-title { margin: 0 0 6px; font-size: 20px; font-weight: 700; color: #1d2129; text-align: center; }
.auth-sub { margin: 0 0 24px; color: #86909c; font-size: 13px; text-align: center; }
.auth-form :deep(.el-form-item__label) { font-weight: 500; color: #1d2129; font-size: 13px; }
.submit-btn { width: 100%; height: 40px; font-size: 14px; letter-spacing: .04em; margin-top: 4px; }
.hint-tip {
  display: flex; align-items: flex-start; gap: 6px;
  margin-top: 16px; padding: 10px 12px;
  background: #f0f4ff; border-radius: 4px; color: #165dff; font-size: 12px; line-height: 1.6;
}

/* 二维码区域 */
.qr-section { display: flex; flex-direction: column; align-items: center; padding: 8px 0 8px; }
.qr-placeholder { display: flex; flex-direction: column; align-items: center; gap: 12px; }
.qr-img-wrap { position: relative; width: 200px; height: 200px; }
.qr-img { width: 200px; height: 200px; border: 1px solid #e5e6eb; border-radius: 10px; display: block; }
.qr-scanned-overlay {
  position: absolute; inset: 0; background: rgba(255,255,255,.88); border-radius: 10px;
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px;
  color: #07c160; font-size: 13px; font-weight: 600;
}
.check-icon { font-size: 36px; }
.qr-expired-wrap {
  width: 200px; height: 200px; border: 1px solid #e5e6eb; border-radius: 10px;
  background: #f5f7fa; display: flex; align-items: center; justify-content: center;
}
.qr-status {
  margin: 10px 0 0; font-size: 12px; color: #86909c;
  display: flex; align-items: center; gap: 6px;
}
.pulse-dot {
  display: inline-block; width: 7px; height: 7px; border-radius: 50%;
  background: #07c160; animation: pulse 1.4s ease-in-out infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: .4; transform: scale(.7); }
}
.qr-tip {
  display: flex; align-items: flex-start; gap: 6px;
  margin-top: 12px; padding: 8px 12px;
  background: #f0f4ff; border-radius: 4px; color: #165dff; font-size: 12px;
}

.link-row { margin-top: 20px; text-align: center; color: #86909c; font-size: 13px; }
.link-a { color: #165dff; text-decoration: none; font-weight: 500; }
.link-a:hover { text-decoration: underline; }
.form-hint { color: #10b981; font-size: 12px; margin-top: 4px; }
.qr-dialog { text-align: center; }
.qr-dialog p { margin: 0 0 16px; color: #374151; }
.qr-dialog-img { width: 220px; height: 220px; object-fit: contain; border: 1px solid #e5e7eb; border-radius: 8px; }
</style>

