<template>
  <el-container class="shell">
    <el-aside width="246px" class="sidebar">
      <div class="logo-wrap">
        <img v-if="siteLogo && !logoImgErr" :src="siteLogo" class="logo-img" alt="logo" @error="logoImgErr = true" />
        <div v-else class="logo-mark">A</div>
        <div>
          <div class="logo">{{ siteName }} Admin</div>
          <div class="logo-sub">Control Panel</div>
        </div>
      </div>
      <el-menu :default-active="route.path" router class="side-menu">
        <el-menu-item index="/admin/dashboard"><el-icon><DataAnalysis /></el-icon>数据概览</el-menu-item>
        <el-menu-item index="/admin/channels"><el-icon><Connection /></el-icon>渠道管理</el-menu-item>
        <el-menu-item index="/admin/key-pools"><el-icon><Key /></el-icon>号池管理</el-menu-item>
        <el-menu-item index="/admin/users"><el-icon><User /></el-icon>用户管理</el-menu-item>
        <el-menu-item index="/admin/billing"><el-icon><Tickets /></el-icon>账单流水</el-menu-item>
        <el-menu-item index="/admin/tasks"><el-icon><Document /></el-icon>任务中心</el-menu-item>
        <el-menu-item index="/admin/llm-logs"><el-icon><ChatLineSquare /></el-icon>LLM 日志</el-menu-item>
        <el-menu-item index="/admin/cards"><el-icon><CreditCard /></el-icon>卡密管理</el-menu-item>
        <el-menu-item index="/admin/ocpc"><el-icon><Promotion /></el-icon>OCPC 上报</el-menu-item>
        <el-menu-item index="/admin/vendors"><el-icon><Goods /></el-icon>号商管理</el-menu-item>
        <el-menu-item index="/admin/withdraw">
          <el-icon><Money /></el-icon>
          提现管理
          <el-badge v-if="pendingWithdraw > 0" :value="pendingWithdraw" type="danger" class="menu-badge" />
        </el-menu-item>
        <el-menu-item index="/admin/settings"><el-icon><Setting /></el-icon>系统设置</el-menu-item>
      </el-menu>
      <div class="sidebar-bottom" @click="logout"><el-icon><SwitchButton /></el-icon>退出</div>
    </el-aside>
    <el-container class="content-wrap">
      <el-header class="header">
        <div>
          <div class="page-title">{{ pageTitle }}</div>
          <div class="page-subtitle">平台管理、利润分析与运营控制</div>
        </div>
        <el-dropdown @command="handleCmd">
          <div class="avatar-btn"><el-icon><UserFilled /></el-icon>我的账户<el-icon class="el-icon--right"><ArrowDown /></el-icon></div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="pwd">修改密码</el-dropdown-item>
              <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main class="page-main"><router-view /></el-main>
    </el-container>
  </el-container>

  <!-- 修改密码 -->
  <el-dialog v-model="showPwd" title="修改密码" width="380px">
    <el-form :model="pwdForm" :rules="pwdRules" ref="pwdFormRef" label-width="90px">
      <el-form-item label="旧密码" prop="old_password">
        <el-input v-model="pwdForm.old_password" type="password" show-password />
      </el-form-item>
      <el-form-item label="新密码" prop="new_password">
        <el-input v-model="pwdForm.new_password" type="password" show-password placeholder="至少 8 位" />
      </el-form-item>
      <el-form-item label="确认密码" prop="confirm">
        <el-input v-model="pwdForm.confirm" type="password" show-password />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="showPwd = false">取消</el-button>
      <el-button type="primary" @click="doChangePwd">确认修改</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { authApi, settingsApi, withdrawApi } from '@/api/admin'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const titles = {
  '/admin/dashboard': '数据概览',
  '/admin/channels': '渠道管理',
  '/admin/key-pools': '号池管理',
  '/admin/users': '用户管理',
  '/admin/billing': '账单流水',
  '/admin/tasks': '任务中心',
  '/admin/cards': '卡密管理',
  '/admin/ocpc': 'OCPC 上报',
  '/admin/llm-logs': 'LLM 日志',
  '/admin/vendors': '号商管理',
  '/admin/withdraw': '提现管理',
  '/admin/settings': '系统设置',
}
const pageTitle = computed(() => titles[route.path] ?? 'FanAPI 管理后台')

// 待处理提现徽标
const pendingWithdraw = ref(0)
let withdrawTimer = null
async function pollWithdraw() {
  try { const r = await withdrawApi.pendingCount(); pendingWithdraw.value = r.count || 0 } catch {}
}

// 动态品牌
const siteName = ref('FanAPI')
const siteLogo = ref('')
const logoImgErr = ref(false)

onMounted(async () => {
  try {
    const res = await settingsApi.get()
    const s = res.settings || {}
    if (s.site_name) siteName.value = s.site_name
    if (s.logo_url) siteLogo.value = s.logo_url
  } catch { /* ignore */ }
  pollWithdraw()
  withdrawTimer = setInterval(pollWithdraw, 30000)
})

onUnmounted(() => clearInterval(withdrawTimer))

// 账户菜单
const showPwd = ref(false)
const pwdFormRef = ref(null)
const pwdForm = ref({ old_password: '', new_password: '', confirm: '' })
const pwdRules = {
  old_password: [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
  new_password: [{ required: true, min: 8, message: '至少 8 位', trigger: 'blur' }],
  confirm: [{ validator: (r, v, cb) => v === pwdForm.value.new_password ? cb() : cb(new Error('两次密码不一致')), trigger: 'blur' }]
}

function handleCmd(cmd) {
  if (cmd === 'pwd') {
    pwdForm.value = { old_password: '', new_password: '', confirm: '' }
    showPwd.value = true
  } else if (cmd === 'logout') {
    logout()
  }
}

async function doChangePwd() {
  await pwdFormRef.value.validate()
  await authApi.changePassword({ old_password: pwdForm.value.old_password, new_password: pwdForm.value.new_password })
  ElMessage.success('密码已修改，请重新登录')
  showPwd.value = false
  setTimeout(logout, 1200)
}

function logout() {
  localStorage.removeItem('admin_token')
  clearInterval(withdrawTimer)
  router.push('/admin/login')
}
</script>

<style scoped>
.shell { min-height:100vh }
.sidebar {
  background:linear-gradient(200deg,#0b1227 0%,#102145 45%,#163575 100%);
  display:flex;
  flex-direction:column;
  padding:16px 14px;
}
.logo-wrap { display:flex;align-items:center;gap:10px;padding:8px 8px 16px }
.logo-mark {
  width:36px;height:36px;border-radius:10px;display:grid;place-items:center;
  font-weight:800;color:#fff;background:linear-gradient(140deg,#1e66ff,#00b4ff);
  flex-shrink:0;
}
.logo-img {
  width:36px;height:36px;border-radius:10px;object-fit:contain;background:#fff;
  flex-shrink:0;
}
.logo { font-size:1.05rem;font-weight:700;color:#fff }
.logo-sub { color:rgba(255,255,255,.72);font-size:.76rem }
.side-menu { border:none;background:transparent }
:deep(.side-menu .el-menu-item) {
  border-radius:10px;margin:4px 0;color:rgba(232,239,255,.84)
}
:deep(.side-menu .el-menu-item:hover) { background:rgba(255,255,255,.1) }
:deep(.side-menu .el-menu-item.is-active) {
  background:linear-gradient(90deg,rgba(30,102,255,.36),rgba(14,197,255,.24));color:#fff
}
.sidebar-bottom {
  margin-top:auto;padding:14px 12px;color:rgba(235,242,255,.82);cursor:pointer;display:flex;align-items:center;gap:8px;border-radius:10px
}
.sidebar-bottom:hover { background:rgba(255,255,255,.08);color:#fff }
.avatar-btn {
  display:flex;align-items:center;gap:6px;cursor:pointer;
  color:#374151;font-size:.9rem;padding:6px 10px;border-radius:8px;
}
.avatar-btn:hover { background:#f3f6fa }
.header {
  display:flex;align-items:center;justify-content:space-between;border-bottom:1px solid #e7edf5;background:rgba(255,255,255,.84);backdrop-filter:blur(8px);padding:0 24px;height:76px
}
.page-title { font-weight:700;font-size:1.1rem }
.page-subtitle { color:#6b7a90;font-size:.82rem;margin-top:3px }
.page-main { padding:22px }
.menu-badge { position:absolute; right:10px; top:50%; transform:translateY(-50%); }
@media (max-width:900px) {
  .shell { display:block }
  .sidebar { width:100% !important;padding:10px }
  .page-main { padding:12px }
}
</style>

