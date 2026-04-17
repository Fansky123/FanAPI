<template>
  <div class="settings-page">
    <el-tabs v-model="activeTab" type="border-card" class="settings-tabs">
      <!-- 基本设置 -->
      <el-tab-pane label="基本设置" name="basic">
        <el-form :model="form" label-width="140px" class="settings-form">
          <el-form-item label="站点名称">
            <el-input v-model="form.site_name" placeholder="例如：FanAPI" />
            <div class="form-tip">显示在浏览器标题栏和页面 Logo 旁</div>
          </el-form-item>
          <el-form-item label="Logo 图片 URL">
            <el-input v-model="form.logo_url" placeholder="https://example.com/logo.png（留空则显示文字）" />
            <div class="form-tip">支持 PNG / SVG，建议尺寸 32×32 或 64×64，留空则使用首字母</div>
          </el-form-item>
          <el-form-item label="Logo 预览" v-if="form.logo_url">
            <div class="logo-preview">
              <img :src="form.logo_url" alt="Logo" @error="logoErr = true" v-if="!logoErr" />
              <span class="logo-err" v-else>图片加载失败，请检查 URL</span>
            </div>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- 页面装饰 -->
      <el-tab-pane label="页眉 & 页脚" name="appearance">
        <el-alert type="warning" :closable="false" show-icon style="margin-bottom:16px">
          <template #title>安全提示：页眉/页脚内容直接通过 <code>v-html</code> 渲染，请勿插入不可信的第三方脚本，避免 XSS 风险。</template>
        </el-alert>
        <el-form :model="form" label-width="140px" class="settings-form">
          <el-form-item label="页眉 HTML">
            <el-input
              v-model="form.header_html"
              type="textarea"
              :rows="6"
              placeholder="<div style='text-align:center;padding:8px;background:#1677ff;color:#fff'>公告：xxx 系统维护中</div>"
            />
            <div class="form-tip">留空则不显示页眉；支持 HTML 和内联样式</div>
          </el-form-item>
          <el-form-item label="页脚 HTML">
            <el-input
              v-model="form.footer_html"
              type="textarea"
              :rows="6"
              placeholder="<div style='text-align:center;padding:16px;color:#888'>© 2025 FanAPI · 服务条款 · 隐私政策</div>"
            />
            <div class="form-tip">留空则不显示页脚；支持 HTML 和内联样式</div>
          </el-form-item>
          <el-form-item label="预览">
            <div class="preview-box">
              <div class="preview-label">页眉预览</div>
              <div class="preview-content" v-html="form.header_html || '<span style=\'color:#aaa\'>（空）</span>'"></div>
              <div class="preview-label" style="margin-top:12px">页脚预览</div>
              <div class="preview-content" v-html="form.footer_html || '<span style=\'color:#aaa\'>（空）</span>'"></div>
            </div>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- 支付设置 -->
      <el-tab-pane label="支付设置" name="payment">
        <el-form :model="form" label-width="140px" class="settings-form">
          <el-form-item label="启用易支付">
            <el-switch v-model="epayEnabledBool" @change="onEpaySwitch" />
            <div class="form-tip">开启后用户可以通过易支付（支付宝 / 微信）充值余额</div>
          </el-form-item>

          <template v-if="epayEnabledBool">
            <el-form-item label="易支付地址">
              <el-input v-model="form.epay_url" placeholder="https://your-epay.com" />
              <div class="form-tip">易支付平台的域名（不含末尾斜杠）</div>
            </el-form-item>
            <el-form-item label="商户 PID">
              <el-input v-model="form.epay_pid" placeholder="您的易支付商户 PID" />
            </el-form-item>
            <el-form-item label="商户密钥">
              <el-input v-model="form.epay_key" type="password" show-password placeholder="您的易支付商户密钥" />
            </el-form-item>
            <el-form-item label="异步通知地址">
              <el-input v-model="form.epay_notify_url" placeholder="https://api.yoursite.com/pay/epay/callback" />
              <div class="form-tip">易支付回调到本系统的地址，必须可从公网访问</div>
            </el-form-item>
            <el-form-item label="同步跳转地址">
              <el-input v-model="form.epay_return_url" placeholder="https://yoursite.com/billing" />
              <div class="form-tip">用户支付成功后跳回的前端页面地址</div>
            </el-form-item>
          </template>

          <el-divider />

          <el-form-item label="启用中台支付">
            <el-switch v-model="payApplyEnabledBool" @change="onPayApplySwitch" />
            <div class="form-tip">开启后用户可通过支付中台（微信 / 支付宝）充值余额</div>
          </el-form-item>

          <template v-if="payApplyEnabledBool">
            <el-form-item label="中台根地址">
              <el-input v-model="form.pay_apply_urlroot" placeholder="https://pay.example.com" />
              <div class="form-tip">支付中台的域名（不含末尾斜杠）</div>
            </el-form-item>
            <el-form-item label="中台商品 Key">
              <el-input v-model="form.pay_apply_key" type="password" show-password placeholder="支付中台分配的商品 key" />
              <div class="form-tip">中台回调时会携带此 key 用于验签，请妥善保管</div>
            </el-form-item>
            <el-form-item label="回调地址">
              <el-input :value="payApplyNotifyUrl" readonly />
              <div class="form-tip">将此地址填写到支付中台的回调配置中，必须可从公网访问</div>
            </el-form-item>
          </template>
        </el-form>
      </el-tab-pane>

      <!-- 公告 & 联系方式 -->
      <el-tab-pane label="公告 & 联系方式" name="notice">
        <el-form :model="form" label-width="140px" class="settings-form">
          <el-form-item label="公告标题">
            <el-input v-model="form.notice_title" placeholder="例如：📢 最新公告" />
            <div class="form-tip">显示在用户数据看板右侧顶部，留空则不显示公告模块</div>
          </el-form-item>
          <el-form-item label="公告内容">
            <el-input
              v-model="form.notice_content"
              type="textarea"
              :rows="6"
              placeholder="支持换行，每行一条公告内容"
            />
            <div class="form-tip">纯文本，每行作为一条，不支持 HTML</div>
          </el-form-item>
          <el-form-item label="联系方式">
            <el-input
              v-model="form.contact_info"
              type="textarea"
              :rows="4"
              placeholder="例如：微信：fanapi&#10;QQ群：123456789&#10;邮箱：support@example.com"
            />
            <div class="form-tip">纯文本，每行一条联系方式，显示在数据看板公告区域</div>
          </el-form-item>
          <el-divider content-position="left" style="margin:8px 0 12px">
            <span style="font-size:13px;color:#666">头部快捷按钮</span>
          </el-divider>
          <el-form-item label="QQ 交流群二维码">
            <div class="qrcode-input-row">
              <el-input
                v-model="form.qq_group_url"
                placeholder="https://example.com/qq.png 或粘贴 base64 数据"
                @input="qqQrcodeErr = false"
                clearable
              />
              <el-button @click="qqQrcodeFileInput?.click()">本地上传</el-button>
              <input ref="qqQrcodeFileInput" type="file" accept="image/*" style="display:none" @change="onQQQrcodeFile" />
            </div>
            <div class="form-tip">填写后，用户页面头部将显示「QQ交流群」按钮，点击弹出二维码；留空则不显示</div>
          </el-form-item>
          <el-form-item label="QQ 二维码预览" v-if="form.qq_group_url">
            <div class="qrcode-preview">
              <img :src="form.qq_group_url" alt="QQ二维码" @error="qqQrcodeErr = true" v-if="!qqQrcodeErr" />
              <span class="logo-err" v-else>图片加载失败，请检查 URL 或 base64 数据</span>
            </div>
          </el-form-item>
          <el-form-item label="微信客服二维码">
            <div class="qrcode-input-row">
              <el-input
                v-model="form.wechat_cs_url"
                placeholder="https://example.com/wechat.png 或粘贴 base64 数据"
                @input="wechatQrcodeErr = false"
                clearable
              />
              <el-button @click="wechatQrcodeFileInput?.click()">本地上传</el-button>
              <input ref="wechatQrcodeFileInput" type="file" accept="image/*" style="display:none" @change="onWechatQrcodeFile" />
            </div>
            <div class="form-tip">填写后，用户页面头部将显示「微信客服」按钮，点击弹出二维码；留空则不显示</div>
          </el-form-item>
          <el-form-item label="微信二维码预览" v-if="form.wechat_cs_url">
            <div class="qrcode-preview">
              <img :src="form.wechat_cs_url" alt="微信二维码" @error="wechatQrcodeErr = true" v-if="!wechatQrcodeErr" />
              <span class="logo-err" v-else>图片加载失败，请检查 URL 或 base64 数据</span>
            </div>
          </el-form-item>
          <el-form-item label="二维码图片">
            <div class="qrcode-input-row">
              <el-input
                v-model="form.qrcode_url"
                placeholder="https://example.com/qrcode.png 或粘贴 base64 数据"
                @input="qrcodeErr = false"
                clearable
              />
              <el-button @click="triggerQrcodeUpload">本地上传</el-button>
              <input ref="qrcodeFileInput" type="file" accept="image/*" style="display:none" @change="onQrcodeFile" />
            </div>
            <div class="form-tip">支持图片 URL 或本地上传（自动转 base64）；留空则不显示</div>
          </el-form-item>
          <el-form-item label="二维码预览" v-if="form.qrcode_url">
            <div class="qrcode-preview">
              <img :src="form.qrcode_url" alt="二维码" @error="qrcodeErr = true" v-if="!qrcodeErr" />
              <span class="logo-err" v-else>图片加载失败，请检查 URL 或 base64 数据</span>
            </div>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- 充值套餐 -->
      <el-tab-pane label="充值套餐" name="plans">
        <el-form label-width="140px" class="settings-form">
          <el-form-item label="允许自定义金额">
            <el-switch v-model="allowCustomBool" />
            <div class="form-tip">开启后用户可在套餐之外自由输入充值金额；关闭则只能选套餐</div>
          </el-form-item>
          <el-form-item label="套餐列表">
            <div style="width:100%">
              <el-table :data="planRows" border size="small" style="margin-bottom:12px">
                <el-table-column label="积分数" width="110">
                  <template #default="{ row }">
                    <el-input-number v-model="row.credits" :min="1" :step="100" size="small" style="width:100%" />
                  </template>
                </el-table-column>
                <el-table-column label="赠送积分" width="110">
                  <template #default="{ row }">
                    <el-input-number v-model="row.bonus" :min="0" :step="100" size="small" style="width:100%" />
                  </template>
                </el-table-column>
                <el-table-column label="金额（元）" width="120">
                  <template #default="{ row }">
                    <el-input-number v-model="row.amount" :min="0.01" :precision="2" :step="10" size="small" style="width:100%" />
                  </template>
                </el-table-column>
                <el-table-column label="原价（元）" width="120">
                  <template #default="{ row }">
                    <el-input-number v-model="row.origin_amount" :min="0" :precision="2" :step="10" size="small" style="width:100%" />
                    <div style="font-size:11px;color:#aaa">留 0 则不显示</div>
                  </template>
                </el-table-column>
                <el-table-column label="描述">
                  <template #default="{ row }">
                    <el-input v-model="row.desc" placeholder="购买可获得xx积分" size="small" />
                  </template>
                </el-table-column>
                <el-table-column label="" width="60" align="center">
                  <template #default="{ $index }">
                    <el-button type="danger" size="small" plain @click="removePlan($index)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
              <el-button type="primary" plain size="small" @click="addPlan">+ 添加套餐</el-button>
              <div class="form-tip" style="margin-top:8px">套餐顺序即展示顺序；不设置套餐则只显示自定义金额输入框</div>
            </div>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- 微信登录 -->
      <el-tab-pane label="微信登录" name="wechat">
        <el-form :model="form" label-width="160px" class="settings-form">
          <el-form-item label="启用微信扫码登录">
            <el-switch v-model="wechatLoginBool" @change="onWechatLoginSwitch" />
            <div class="form-tip">开启后用户登录/注册页面将显示微信扫码登录按钮</div>
          </el-form-item>
          <template v-if="wechatLoginBool">
            <el-form-item label="公众号 AppID">
              <el-input v-model="form.wechat_appid" placeholder="wx开头的 AppID" />
              <div class="form-tip">微信公众平台 → 设置 → 公众号设置 → 基本配置</div>
            </el-form-item>
            <el-form-item label="公众号 AppSecret">
              <el-input v-model="form.wechat_secret" type="password" show-password placeholder="AppSecret（机密，请勿外泄）" />
            </el-form-item>
            <el-form-item label="回调域名（Base URL）">
              <el-input v-model="form.wechat_redirect_base_url" placeholder="https://yourdomain.com" />
              <div class="form-tip">系统将在此域名下接收微信 OAuth 回调，格式：https://域名（无末尾斜杠）</div>
            </el-form-item>
            <el-form-item label="登录成功跳转地址">
              <el-input v-model="form.wechat_frontend_url" placeholder="https://yourdomain.com（留空则跳转首页）" />
              <div class="form-tip">微信授权成功后跳转到的前端地址（通常不需要配置）</div>
            </el-form-item>
          </template>
        </el-form>
      </el-tab-pane>

      <!-- 公众号扫码登录 -->
      <el-tab-pane label="公众号扫码" name="wechat_mp">
        <el-form :model="form" label-width="160px" class="settings-form">
          <el-alert type="info" :closable="false" show-icon style="margin-bottom:16px">
            <template #title>此功能使用微信公众号事件推送实现扫码登录，与"微信登录"（H5 OAuth）相互独立，可同时开启。</template>
          </el-alert>
          <el-form-item label="启用公众号扫码登录">
            <el-switch v-model="wechatMPLoginBool" @change="onWechatMPLoginSwitch" />
            <div class="form-tip">开启后用户登录页面将显示公众号二维码，扫码关注即可登录</div>
          </el-form-item>
          <template v-if="wechatMPLoginBool">
            <el-form-item label="公众号 AppID">
              <el-input v-model="form.wechat_mp_appid" placeholder="wx开头的 AppID" />
              <div class="form-tip">微信公众平台 → 设置 → 基本配置 → AppID</div>
            </el-form-item>
            <el-form-item label="公众号 AppSecret">
              <el-input v-model="form.wechat_mp_secret" type="password" show-password placeholder="AppSecret（机密，请勿外泄）" />
            </el-form-item>
            <el-form-item label="Webhook Token">
              <el-input v-model="form.wechat_mp_token" placeholder="自定义字符串，需与公众号后台一致" />
              <div class="form-tip">
                用于验证微信推送请求的合法性。<br>
                微信公众平台 → 设置 → 公众号设置 → 功能设置 → 服务器配置<br>
                Webhook URL：<code>{{ mpWebhookUrl }}</code>
              </div>
            </el-form-item>
          </template>
        </el-form>
      </el-tab-pane>

      <!-- OCPC 推广（已迁移至独立的 OCPC 管理页面） -->
      <el-tab-pane label="OCPC 推广" name="ocpc">
        <el-alert type="info" :closable="false" show-icon style="margin-bottom:16px">
          <template #title>
            OCPC 推广账户配置已迁移至专属管理页面，支持多账户。
            请前往左侧菜单 <b>OCPC 管理</b> 进行账户添加与上报操作。
          </template>
        </el-alert>
      </el-tab-pane>
    </el-tabs>

    <div class="save-bar">
      <el-button type="primary" :loading="saving" @click="save" size="large">
        <el-icon><Check /></el-icon>
        保存设置
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Check } from '@element-plus/icons-vue'
import { settingsApi } from '@/api/admin'

const activeTab = ref('basic')
const saving = ref(false)
const logoErr = ref(false)
const qrcodeErr = ref(false)
const qqQrcodeErr = ref(false)
const wechatQrcodeErr = ref(false)
const qrcodeFileInput = ref(null)
const qqQrcodeFileInput = ref(null)
const wechatQrcodeFileInput = ref(null)

function triggerQrcodeUpload() {
  qrcodeFileInput.value?.click()
}

function onQrcodeFile(e) {
  const file = e.target.files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = (ev) => {
    form.qrcode_url = ev.target.result
    qrcodeErr.value = false
  }
  reader.readAsDataURL(file)
  e.target.value = ''
}

function onQQQrcodeFile(e) {
  const file = e.target.files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = (ev) => {
    form.qq_group_url = ev.target.result
    qqQrcodeErr.value = false
  }
  reader.readAsDataURL(file)
  e.target.value = ''
}

function onWechatQrcodeFile(e) {
  const file = e.target.files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = (ev) => {
    form.wechat_cs_url = ev.target.result
    wechatQrcodeErr.value = false
  }
  reader.readAsDataURL(file)
  e.target.value = ''
}

const form = reactive({
  site_name: '',
  logo_url: '',
  header_html: '',
  footer_html: '',
  epay_enabled: 'false',
  epay_url: '',
  epay_pid: '',
  epay_key: '',
  epay_notify_url: '',
  epay_return_url: '',
  pay_apply_enabled: 'false',
  pay_apply_urlroot: '',
  pay_apply_key: '',
  notice_title: '',
  notice_content: '',
  contact_info: '',
  qrcode_url: '',
  qq_group_url: '',
  wechat_cs_url: '',
  recharge_allow_custom: 'true',
  recharge_plans: '[]',
  // 微信登录
  wechat_login_enabled: 'false',
  wechat_appid: '',
  wechat_secret: '',
  wechat_redirect_base_url: '',
  wechat_frontend_url: '',
  // 公众号扫码登录
  wechat_mp_login_enabled: 'false',
  wechat_mp_appid: '',
  wechat_mp_secret: '',
  wechat_mp_token: '',
  // 百度 OCPC
  ocpc_baidu_enabled: 'false',
  ocpc_baidu_token: '',
  ocpc_baidu_page_url: '',
  // 360 OCPC
  ocpc_360_enabled: 'false',
  ocpc_360_key: '',
  ocpc_360_secret: '',
  ocpc_360_jzqs: '',
  ocpc_360_so_type: '1',
})

// 套餐行（临时可编辑数组）
const planRows = ref([])
const allowCustomBool = ref(true)

function addPlan() {
  planRows.value.push({ credits: 100, bonus: 0, amount: 10, origin_amount: 0, desc: '' })
}
function removePlan(index) {
  planRows.value.splice(index, 1)
}

const epayEnabledBool = ref(false)
const payApplyEnabledBool = ref(false)
const wechatLoginBool = ref(false)
const wechatMPLoginBool = ref(false)
const baiduOcpcBool = ref(false)
const e360OcpcBool = ref(false)

function onEpaySwitch(v) {
  form.epay_enabled = v ? 'true' : 'false'
  if (v) {
    payApplyEnabledBool.value = false
    form.pay_apply_enabled = 'false'
  }
}
function onPayApplySwitch(v) {
  form.pay_apply_enabled = v ? 'true' : 'false'
  if (v) {
    epayEnabledBool.value = false
    form.epay_enabled = 'false'
  }
}
function onWechatLoginSwitch(v) {
  form.wechat_login_enabled = v ? 'true' : 'false'
}
function onWechatMPLoginSwitch(v) {
  form.wechat_mp_login_enabled = v ? 'true' : 'false'
}
function onBaiduOcpcSwitch(v) {
  form.ocpc_baidu_enabled = v ? 'true' : 'false'
}
function on360OcpcSwitch(v) {
  form.ocpc_360_enabled = v ? 'true' : 'false'
}

// 中台回调地址自动生成（当前域名 + 固定路径）
const payApplyNotifyUrl = computed(() => {
  const origin = window.location.origin.replace(':3001', '')
  return `${origin}/pay/apply/notify`
})

// 公众号扫码 Webhook 地址提示（只读）
const mpWebhookUrl = computed(() => {
  const origin = window.location.origin.replace(':3001', '')
  return `${origin}/auth/wechat-mp/event`
})

watch(() => form.logo_url, () => { logoErr.value = false })
watch(() => form.epay_enabled, (v) => { epayEnabledBool.value = v === 'true' })
watch(() => form.pay_apply_enabled, (v) => { payApplyEnabledBool.value = v === 'true' })
watch(() => form.wechat_login_enabled, (v) => { wechatLoginBool.value = v === 'true' })
watch(() => form.wechat_mp_login_enabled, (v) => { wechatMPLoginBool.value = v === 'true' })
watch(() => form.ocpc_baidu_enabled, (v) => { baiduOcpcBool.value = v === 'true' })
watch(() => form.ocpc_360_enabled, (v) => { e360OcpcBool.value = v === 'true' })

onMounted(async () => {
  try {
    const res = await settingsApi.get()
    const s = res.settings || {}
    Object.keys(form).forEach(k => { if (s[k] !== undefined) form[k] = s[k] })
    epayEnabledBool.value = form.epay_enabled === 'true'
    payApplyEnabledBool.value = form.pay_apply_enabled === 'true'
    wechatLoginBool.value = form.wechat_login_enabled === 'true'
    wechatMPLoginBool.value = form.wechat_mp_login_enabled === 'true'
    baiduOcpcBool.value = form.ocpc_baidu_enabled === 'true'
    e360OcpcBool.value = form.ocpc_360_enabled === 'true'
    allowCustomBool.value = form.recharge_allow_custom !== 'false'
    try { planRows.value = JSON.parse(form.recharge_plans || '[]') } catch { planRows.value = [] }
  } catch {
    ElMessage.error('加载设置失败')
  }
})

async function save() {
  saving.value = true
  try {
    form.recharge_allow_custom = allowCustomBool.value ? 'true' : 'false'
    form.recharge_plans = JSON.stringify(planRows.value)
    // 回调地址是只读展示，不保存到后端
    const { payApplyNotifyUrl, ...saveForm } = { ...form }
    await settingsApi.update(saveForm)
    ElMessage.success('设置已保存')
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.settings-page {
  max-width: 860px;
}
.settings-tabs {
  border-radius: 12px;
  overflow: hidden;
}
.settings-form {
  padding: 16px 0;
  max-width: 700px;
}
.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.5;
}
.logo-preview {
  width: 80px;
  height: 80px;
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
}
.logo-preview img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}
.logo-err {
  font-size: 12px;
  color: #f56c6c;
  text-align: center;
  padding: 4px;
}
.qrcode-input-row {
  display: flex;
  gap: 8px;
  align-items: center;
  width: 100%;
}
.qrcode-input-row .el-input {
  flex: 1;
}
.qrcode-preview {  width: 140px;
  height: 140px;
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
}
.qrcode-preview img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}
.preview-box {
  width: 100%;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  overflow: hidden;
}
.preview-label {
  font-size: 12px;
  color: #909399;
  padding: 6px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}
.preview-content {
  min-height: 36px;
}
.save-bar {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}
</style>
