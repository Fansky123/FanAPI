<template>
  <div class="auth-page">
    <div class="auth-card">
      <!-- Logo -->
      <div class="auth-logo">
        <div class="logo-icon">{{ site.siteName.charAt(0).toUpperCase() }}</div>
        <span class="logo-name">{{ site.siteName }}</span>
      </div>

      <!-- 模式切换（仅在公众号扫码登录开启时显示） -->
      <div v-if="site.wechatMPLoginEnabled" class="mode-toggle">
        <button :class="['mode-btn', mode === 'password' ? 'active' : '']" @click="switchMode('password')">
          账号密码
        </button>
        <button :class="['mode-btn', mode === 'qr' ? 'active' : '']" @click="switchMode('qr')">
          <span class="wechat-dot">●</span> 微信扫码
        </button>
      </div>

      <!-- 账号密码登录 -->
      <template v-if="mode === 'password'">
        <h2 class="auth-title">登录账户</h2>
        <p class="auth-sub">欢迎回来，请输入登录信息</p>

        <el-form :model="form" @submit.prevent="handleLogin" label-position="top" class="auth-form">
          <el-form-item label="用户名 / 邮箱">
            <el-input v-model="form.username" placeholder="用户名或绑定邮箱" clearable size="large" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="form.password" type="password" show-password size="large" placeholder="请输入密码" />
          </el-form-item>
          <div class="forgot-row">
            <router-link to="/forgot-password" class="forgot-link">忘记密码？</router-link>
          </div>
          <el-button type="primary" native-type="submit" :loading="loading" class="submit-btn">
            登 录
          </el-button>
        </el-form>

        <!-- H5 OAuth 微信登录（与公众号扫码互不干扰） -->
        <template v-if="site.wechatLoginEnabled && !site.wechatMPLoginEnabled">
          <div class="divider"><span>或</span></div>
          <el-button class="wechat-btn" @click="openWechat" :loading="wechatLoading">
            <img src="https://res.wx.qq.com/a/wx_fed/assets/res/OTE0YTAw.png" alt="" class="wechat-icon" />
            微信扫码登录
          </el-button>
        </template>
      </template>

      <!-- 微信公众号扫码登录 -->
      <template v-else>
        <h2 class="auth-title">微信扫码登录</h2>
        <p class="auth-sub">使用微信扫描二维码，关注公众号即可登录</p>

        <div class="qr-section">
          <!-- 加载中 -->
          <div v-if="qrLoading" class="qr-placeholder">
            <el-skeleton :rows="0" animated style="width:200px;height:200px;border-radius:8px" />
            <p class="qr-status">正在生成二维码…</p>
          </div>

          <!-- 二维码展示 -->
          <template v-else-if="qrImg && qrStatus !== 'expired'">
            <div class="qr-img-wrap">
              <img :src="qrImg" class="qr-img" alt="微信登录二维码" />
              <div v-if="qrStatus === 'scanned'" class="qr-scanned-overlay">
                <el-icon class="check-icon"><CircleCheckFilled /></el-icon>
                <span>已扫码，正在登录…</span>
              </div>
            </div>
            <p class="qr-status">
              <span v-if="qrStatus === 'pending'">
                <span class="pulse-dot" />请用微信扫描二维码（{{ countdown }}s 后过期）
              </span>
              <span v-else-if="qrStatus === 'scanned'" style="color:#07c160">扫码成功，正在登录…</span>
            </p>
          </template>

          <!-- 已过期 -->
          <template v-else>
            <div class="qr-expired-wrap">
              <div class="qr-expired-mask">
                <el-button type="primary" plain round @click="loadQRCode">
                  <el-icon><RefreshRight /></el-icon> 刷新二维码
                </el-button>
              </div>
            </div>
            <p class="qr-status" style="color:#f56c6c">二维码已过期，请刷新</p>
          </template>
        </div>
      </template>

      <div class="link-row">
        还没有账号？<router-link to="/register" class="link-a">立即注册</router-link>
      </div>
    </div>
  </div>

  <!-- 客服微信二维码弹窗（登录成功后） -->
  <el-dialog v-model="showInviterQR" title="专属客服" width="320px" :close-on-click-modal="false" align-center>
    <div class="qr-dialog">
      <p>欢迎回来！扫码添加您的专属客服微信获取使用帮助。</p>
      <img :src="inviterQR" alt="客服微信" class="qr-dialog-img" />
    </div>
    <template #footer>
      <el-button type="primary" @click="closeQRAndGo">进入首页</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive, ref, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import { authApi } from '@/api'
import { ElMessage } from 'element-plus'
import { CircleCheckFilled, RefreshRight } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const store = useUserStore()
const site = useSiteStore()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

// 广告追踪参数（从落地页 URL 读取，随 QR 请求传入）
function getTrackParams() {
  const q = route.query
  const p = {}
  if (q.bd_vid) p.bd_vid = q.bd_vid
  if (q.qh_click_id) p.qh_click_id = q.qh_click_id
  if (q.source_id) p.source_id = q.source_id
  if (q.ocpc_id) p.ocpc_id = q.ocpc_id
  return p
}

// 客服 QR 弹窗
const showInviterQR = ref(false)
const inviterQR = ref('')

// ── 模式切换 ──────────────────────────────────
const mode = ref('password') // 'password' | 'qr'

function switchMode(m) {
  if (mode.value === m) return
  stopPoll()
  mode.value = m
  if (m === 'qr') loadQRCode()
}

// ── 账号密码登录 ──────────────────────────────
async function handleLogin() {
  loading.value = true
  try {
    const res = await authApi.login(form)
    store.setToken(res.token)
    store.setUsername(res.user?.username || form.username)
    if (res.inviter_wechat_qr) {
      inviterQR.value = res.inviter_wechat_qr
      showInviterQR.value = true
    } else {
      router.push('/models')
    }
  } catch {
    // 错误已由 HTTP 拦截器展示
  } finally {
    loading.value = false
  }
}

function closeQRAndGo() {
  showInviterQR.value = false
  router.push('/models')
}

// ── H5 OAuth 微信登录（备用） ─────────────────
const wechatLoading = ref(false)
const wechatState = ref('')

async function openWechat() {
  wechatLoading.value = true
  try {
    const res = await authApi.wechatInit()
    wechatState.value = res.state
    // H5 OAuth 会重定向，无需在此处理
  } catch {
    // 已展示
  } finally {
    wechatLoading.value = false
  }
}

// ── 公众号扫码登录 ────────────────────────────
const qrLoading = ref(false)
const qrImg = ref('')       // base64 data URI
const qrUUID = ref('')
const qrStatus = ref('pending') // 'pending' | 'scanned' | 'expired'
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
    const res = await authApi.wechatMPQRCode(getTrackParams())
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
        ElMessage.success('微信登录成功')
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
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(126.82deg, rgba(236,243,255,.8) 0.58%, rgba(232,247,251,.8) 86.28%), #f2f3f5;
  padding: 24px;
}
.auth-card {
  width: 100%;
  max-width: 400px;
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  padding: 36px 32px;
  box-shadow: 0 4px 20px rgba(0,0,0,.06);
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

/* 模式切换 */
.mode-toggle {
  display: flex; gap: 0; margin-bottom: 20px;
  border: 1px solid #dde1e7; border-radius: 8px; overflow: hidden;
}
.mode-btn {
  flex: 1; padding: 9px 0; border: none; background: #f5f7fa;
  cursor: pointer; font-size: 13px; color: #4e5969; transition: all .18s;
  display: flex; align-items: center; justify-content: center; gap: 5px;
}
.mode-btn.active { background: #165dff; color: #fff; font-weight: 600; }
.mode-btn:not(.active):hover { background: #eef2ff; }
.wechat-dot { color: #07c160; font-size: 9px; line-height: 1; }
.mode-btn.active .wechat-dot { color: #a7f3d0; }

.auth-title {
  margin: 0 0 6px; font-size: 20px; font-weight: 700;
  color: #1d2129; text-align: center;
}
.auth-sub { margin: 0 0 24px; color: #86909c; font-size: 13px; text-align: center; }
.auth-form :deep(.el-form-item__label) { font-weight: 500; color: #1d2129; font-size: 13px; }
.forgot-row { text-align: right; margin-bottom: 16px; }
.forgot-link { font-size: 12px; color: #86909c; text-decoration: none; }
.forgot-link:hover { color: #165dff; }
.submit-btn { width: 100%; height: 40px; font-size: 14px; letter-spacing: .04em; }
.divider {
  display: flex; align-items: center; gap: 10px;
  margin: 16px 0 12px; color: #c2c7d0; font-size: 12px;
}
.divider::before, .divider::after { content: ''; flex: 1; height: 1px; background: #e5e6eb; }
.wechat-btn {
  width: 100%; height: 40px; background: #07C160; color: #fff; border: none;
  font-size: 14px; display: flex; align-items: center; justify-content: center; gap: 8px;
}
.wechat-btn:hover { background: #06AD56; color: #fff; }
.wechat-icon { width: 20px; height: 20px; }

/* 二维码区域 */
.qr-section { display: flex; flex-direction: column; align-items: center; padding: 8px 0 16px; }
.qr-placeholder { display: flex; flex-direction: column; align-items: center; gap: 12px; }
.qr-img-wrap { position: relative; width: 200px; height: 200px; }
.qr-img {
  width: 200px; height: 200px;
  border: 1px solid #e5e6eb; border-radius: 10px;
  display: block;
}
.qr-scanned-overlay {
  position: absolute; inset: 0;
  background: rgba(255,255,255,.88);
  border-radius: 10px;
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px;
  color: #07c160; font-size: 13px; font-weight: 600;
}
.check-icon { font-size: 36px; }
.qr-expired-wrap {
  width: 200px; height: 200px;
  border: 1px solid #e5e6eb; border-radius: 10px;
  background: #f5f7fa;
  display: flex; align-items: center; justify-content: center;
}
.qr-expired-mask { display: flex; flex-direction: column; align-items: center; gap: 12px; }
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

.link-row { margin-top: 20px; text-align: center; color: #86909c; font-size: 13px; }
.link-a { color: #165dff; text-decoration: none; font-weight: 500; }
.link-a:hover { text-decoration: underline; }

.qr-dialog { text-align: center; }
.qr-dialog p { margin: 0 0 16px; color: #374151; }
.qr-dialog-img {
  width: 220px; height: 220px; object-fit: contain;
  border: 1px solid #e5e7eb; border-radius: 8px;
}
</style>

