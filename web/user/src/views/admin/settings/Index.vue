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
            <el-switch v-model="epayEnabledBool" @change="v => form.epay_enabled = v ? 'true' : 'false'" />
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
            <div class="form-tip">纯文本，每行一条联系方式</div>
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
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Check } from '@element-plus/icons-vue'
import { settingsApi } from '@/api/admin'

const activeTab = ref('basic')
const saving = ref(false)
const logoErr = ref(false)
const qrcodeErr = ref(false)
const qrcodeFileInput = ref(null)

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
  // 重置 input 以便同一文件可再次选择
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
  notice_title: '',
  notice_content: '',
  contact_info: '',
  qrcode_url: '',
})

const epayEnabledBool = ref(false)

watch(() => form.logo_url, () => { logoErr.value = false })
watch(() => form.epay_enabled, (v) => { epayEnabledBool.value = v === 'true' })

onMounted(async () => {
  try {
    const res = await settingsApi.get()
    const s = res.settings || {}
    Object.keys(form).forEach(k => { if (s[k] !== undefined) form[k] = s[k] })
    epayEnabledBool.value = form.epay_enabled === 'true'
  } catch {
    ElMessage.error('加载设置失败')
  }
})

async function save() {
  saving.value = true
  try {
    await settingsApi.update({ ...form })
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
