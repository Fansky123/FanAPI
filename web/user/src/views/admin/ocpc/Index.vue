<template>
  <div class="ocpc-page">
    <!-- ─── Hero ─────────────────────────────────────── -->
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">OCPC</div>
          <h3>推广账户管理</h3>
          <p>
            管理百度 / 360 OCPC 转化上报账户。在落地页 URL 末尾追加
            <el-tag size="small" type="info" style="font-family:monospace">?ocpc_id=&lt;ID&gt;</el-tag>
            以识别来源账户。
          </p>
        </div>
        <el-button type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon> 新增账户
        </el-button>
      </div>
    </el-card>

    <!-- ─── 账户列表 ──────────────────────────────────── -->
    <el-card>
      <el-table :data="platforms" stripe border v-loading="tableLoading" empty-text="暂无账户配置">
        <el-table-column prop="id" label="ID" width="64" align="center" />
        <el-table-column label="平台" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.platform === 'baidu' ? 'primary' : 'warning'" size="small">
              {{ row.platform === 'baidu' ? '百度 OCPC' : '360 OCPC' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称 / 备注" min-width="140" />
        <el-table-column label="关键配置" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="mono-text">
              <template v-if="row.platform === 'baidu'">{{ row.baidu_page_url || '—' }}</template>
              <template v-else>Key: {{ row.e360_key || '—' }}</template>
            </span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-switch :model-value="row.enabled" @change="toggle(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" align="center">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-popconfirm
              :title="`确认删除「${row.name}」？`"
              confirm-button-text="删除"
              cancel-button-text="取消"
              confirm-button-type="danger"
              @confirm="remove(row)"
            >
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- ─── 手动上报 ──────────────────────────────────── -->
    <el-card>
      <template #header>
        <div class="card-header-row">
          <span class="card-header-title">手动触发上报</span>
        </div>
      </template>
      <p class="upload-desc">将所有待上报的注册 / 订单转化推送到各平台。</p>
      <el-button type="primary" :loading="uploading" @click="upload">
        <el-icon v-if="!uploading"><Upload /></el-icon>
        {{ uploading ? '上报中…' : '立即上报' }}
      </el-button>

      <el-descriptions
        v-if="result"
        :column="2"
        border
        size="small"
        style="margin-top:20px;max-width:480px"
      >
        <el-descriptions-item label="注册成功">
          <el-text type="success" tag="b">{{ result.reg_ok }}</el-text>
        </el-descriptions-item>
        <el-descriptions-item label="注册失败">
          <el-text type="danger" tag="b">{{ result.reg_fail }}</el-text>
        </el-descriptions-item>
        <el-descriptions-item label="订单成功">
          <el-text type="success" tag="b">{{ result.order_ok }}</el-text>
        </el-descriptions-item>
        <el-descriptions-item label="订单失败">
          <el-text type="danger" tag="b">{{ result.order_fail }}</el-text>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- ─── 定时上报配置 ───────────────────────────────── -->
    <el-card>
      <template #header>
        <div class="card-header-row">
          <span class="card-header-title">定时上报配置</span>
        </div>
      </template>
      <el-form :model="schedule" label-width="110px" style="max-width:480px" v-loading="scheduleLoading">
        <el-form-item label="自动上报">
          <el-switch
            v-model="schedule.enabled"
            active-text="启用"
            inactive-text="停用"
          />
        </el-form-item>
        <el-form-item label="上报间隔">
          <el-select v-model="schedule.interval" style="width:180px" :disabled="!schedule.enabled">
            <el-option label="每 10 分钟" :value="10" />
            <el-option label="每 15 分钟" :value="15" />
            <el-option label="每 30 分钟" :value="30" />
            <el-option label="每 60 分钟" :value="60" />
            <el-option label="每 120 分钟" :value="120" />
            <el-option label="每 360 分钟" :value="360" />
          </el-select>
        </el-form-item>
        <el-form-item label="上次运行">
          <el-text type="info">{{ schedule.lastRunAt || '—' }}</el-text>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="scheduleSaving" @click="saveSchedule">保存</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- ─── 新增 / 编辑弹窗 ───────────────────────────── -->
    <el-dialog
      v-model="dialog.show"
      :title="dialog.id ? '编辑账户' : '新增账户'"
      width="520px"
      top="8vh"
      destroy-on-close
    >
      <el-form :model="dialog.form" label-width="110px" style="padding-right:8px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="平台">
              <el-select v-model="dialog.form.platform" :disabled="!!dialog.id" style="width:100%">
                <el-option label="百度 OCPC" value="baidu" />
                <el-option label="360 OCPC" value="360" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="名称/备注">
              <el-input v-model="dialog.form.name" placeholder="如：百度主账户" />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 百度字段 -->
        <template v-if="dialog.form.platform === 'baidu'">
          <el-form-item label="OCPC Token">
            <el-input v-model="dialog.form.baidu_token" placeholder="百度 OCPC Token" show-password />
          </el-form-item>
          <el-form-item label="落地页 URL">
            <el-input v-model="dialog.form.baidu_page_url" placeholder="https://yoursite.com/register" />
            <div class="form-tip">系统会自动拼接 ?bd_vid=xxx</div>
          </el-form-item>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="注册转化类型">
                <el-select v-model="dialog.form.baidu_reg_type" style="width:100%">
                  <el-option v-for="t in baiduTypes" :key="t.value" :label="t.label" :value="t.value" />
                </el-select>
                <div class="form-tip">默认 68（注册/关注）</div>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="购买转化类型">
                <el-select v-model="dialog.form.baidu_order_type" style="width:100%">
                  <el-option v-for="t in baiduTypes" :key="t.value" :label="t.label" :value="t.value" />
                </el-select>
                <div class="form-tip">默认 10（购买）</div>
              </el-form-item>
            </el-col>
          </el-row>
        </template>

        <!-- 360 字段 -->
        <template v-if="dialog.form.platform === '360'">
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="App-Key">
                <el-input v-model="dialog.form.e360_key" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="App-Secret">
                <el-input v-model="dialog.form.e360_secret" placeholder="留空保持不变" show-password />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="推广类型">
                <el-select v-model="dialog.form.e360_so_type" style="width:100%">
                  <el-option label="搜索推广 (1)" value="1" />
                  <el-option label="展示广告 (2)" value="2" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="计划ID (jzqs)">
                <el-input v-model="dialog.form.e360_jzqs" placeholder="展示广告用" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="注册转化事件">
                <el-select v-model="dialog.form.e360_reg_event" style="width:100%" filterable>
                  <el-option label="REGISTERED（默认）" value="" />
                  <el-option v-for="e in e360Events" :key="e.value" :label="e.label" :value="e.value" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="购买转化事件">
                <el-select v-model="dialog.form.e360_order_event" style="width:100%" filterable>
                  <el-option label="ORDER（默认）" value="" />
                  <el-option v-for="e in e360Events" :key="e.value" :label="e.label" :value="e.value" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>
        </template>

        <el-form-item label="状态">
          <el-switch v-model="dialog.form.enabled" active-text="启用" inactive-text="停用" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialog.show = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">
          {{ saving ? '保存中…' : '保存' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Upload } from '@element-plus/icons-vue'
import { ocpcApi, ocpcPlatformApi } from '@/api/admin'

// 百度 newType 选项
const baiduTypes = [
  { value: 68,  label: '68 — 注册 / 关注' },
  { value: 10,  label: '10 — 购买' },
  { value: 3,   label: '3  — 表单提交' },
  { value: 6,   label: '6  — 加入购物车' },
  { value: 14,  label: '14 — 下单' },
  { value: 15,  label: '15 — 付款' },
  { value: 19,  label: '19 — APP 激活' },
  { value: 1,   label: '1  — 网页浏览' },
]

// 360 转化事件（完整列表）
const e360Events = [
  { value: 'SUBMIT',                 label: 'SUBMIT — 表单提交' },
  { value: 'CALL',                   label: 'CALL — 有效电话拨打' },
  { value: 'ADVISORY',              label: 'ADVISORY — 一句话咨询' },
  { value: 'SITEDOWNLOAD',          label: 'SITEDOWNLOAD — 下载按钮点击' },
  { value: 'SUBMIT_BUTTON',         label: 'SUBMIT_BUTTON — 表单按钮点击' },
  { value: 'ADVISORY_BUTTON',       label: 'ADVISORY_BUTTON — 咨询按钮点击' },
  { value: 'CALL_BUTTON',           label: 'CALL_BUTTON — 电话按钮点击' },
  { value: 'SHOP_BUTTON',           label: 'SHOP_BUTTON — 购买按钮点击' },
  { value: 'CART_BUTTON',           label: 'CART_BUTTON — 加购物车按钮点击' },
  { value: 'ORDER',                  label: 'ORDER — 订单' },
  { value: 'REGISTERED',            label: 'REGISTERED — 注册' },
  { value: 'ROLE_CREAT',            label: 'ROLE_CREAT — 创建角色' },
  { value: 'SITE_VISIT_DEPTH',      label: 'SITE_VISIT_DEPTH — 深度页面访问' },
  { value: 'COUSTOMIZE',            label: 'COUSTOMIZE — 客户自定义' },
  { value: 'MIDDLE_PAGE',           label: 'MIDDLE_PAGE — 中间页' },
  { value: 'REGISTER_BUTTON',       label: 'REGISTER_BUTTON — 注册按钮点击' },
  { value: 'BROWSE_DEPTH',          label: 'BROWSE_DEPTH — 有效浏览' },
  { value: 'BROWSETIME',            label: 'BROWSETIME — 浏览时长' },
  { value: 'SCAN_BUTTON',           label: 'SCAN_BUTTON — 扫码点击' },
  { value: 'ADVISORY_DEPTH',        label: 'ADVISORY_DEPTH — 三句话咨询' },
  { value: 'LOW_PAY',               label: 'LOW_PAY — 低价订单付费' },
  { value: 'ADD_FANS_WX',           label: 'ADD_FANS_WX — 微信加粉' },
  { value: 'PAY',                    label: 'PAY — 付费' },
  { value: 'SCAN_CODE',             label: 'SCAN_CODE — 扫码' },
  { value: 'APPLET_STARTUP',        label: 'APPLET_STARTUP — 小程序调起' },
  { value: 'LOGIN',                  label: 'LOGIN — 登录' },
  { value: 'ADD_TO_CART',           label: 'ADD_TO_CART — 加购物车' },
  { value: 'VPPV',                   label: 'VPPV — 深度页面访问' },
  { value: 'INTENTIONAL',           label: 'INTENTIONAL — 有意向' },
  { value: 'REAL_NAME',             label: 'REAL_NAME — 实名认证' },
  { value: 'RETENTION',             label: 'RETENTION — 次日留存' },
  { value: 'PLACE_ORDER',           label: 'PLACE_ORDER — 订单提交' },
  { value: 'EFFECTIVE_ADVISORY',    label: 'EFFECTIVE_ADVISORY — 有效咨询' },
  { value: 'ORDER_VALIDITY',        label: 'ORDER_VALIDITY — 订单有效性' },
  { value: 'ACTIVATION',            label: 'ACTIVATION — 激活' },
  { value: 'DETAILS_PAGE_ARRIVED',  label: 'DETAILS_PAGE_ARRIVED — 详情页到达' },
  { value: 'PAY_SUCCESS',           label: 'PAY_SUCCESS — 支付成功' },
  { value: 'CREDIT',                 label: 'CREDIT — 授信' },
  { value: 'WX_BUTTON_C',           label: 'WX_BUTTON_C — 微信复制按钮点击' },
  { value: 'CONCEM',                 label: 'CONCEM — 关注' },
  { value: 'LEAVE_CONTACT',         label: 'LEAVE_CONTACT — 留联' },
  { value: 'RELEASE',               label: 'RELEASE — 发布' },
  { value: 'TRY_TO_PLAY',           label: 'TRY_TO_PLAY — 试玩' },
  { value: 'SUBMIT_RESUME',         label: 'SUBMIT_RESUME — 投递简历' },
  { value: 'ENTERPRISE_CERTIFICATION', label: 'ENTERPRISE_CERTIFICATION — 企业认证' },
  { value: 'VISIT_CLINIC',          label: 'VISIT_CLINIC — 到诊' },
  { value: 'APPLET_PAY',            label: 'APPLET_PAY — 小程序内付费' },
  { value: 'APPLET_ROLE_CREAT',     label: 'APPLET_ROLE_CREAT — 小程序内创角' },
  { value: 'TRIAL_LESSON_PERFORM',  label: 'TRIAL_LESSON_PERFORM — 试听到课' },
  { value: 'TRIAL_LESSON_COMPLETE', label: 'TRIAL_LESSON_COMPLETE — 试听完课' },
  { value: 'TRIAL',                  label: 'TRIAL — 试用' },
]

const platforms = ref([])
const tableLoading = ref(false)
const uploading = ref(false)
const saving = ref(false)
const result = ref(null)

// 定时上报
const scheduleLoading = ref(false)
const scheduleSaving = ref(false)
const schedule = ref({ enabled: false, interval: 30, lastRunAt: '' })

const defaultForm = () => ({
  platform: 'baidu',
  name: '',
  enabled: true,
  baidu_token: '',
  baidu_page_url: '',
  baidu_reg_type: 68,
  baidu_order_type: 10,
  e360_key: '',
  e360_secret: '',
  e360_jzqs: '',
  e360_so_type: '1',
  e360_reg_event: '',
  e360_order_event: '',
})

const dialog = ref({ show: false, id: null, form: defaultForm() })

async function load() {
  tableLoading.value = true
  try {
    const res = await ocpcPlatformApi.list()
    platforms.value = res.list || []
  } catch (e) {
    ElMessage.error('加载失败')
  } finally {
    tableLoading.value = false
  }
}

function openCreate() {
  dialog.value = { show: true, id: null, form: defaultForm() }
}

function openEdit(p) {
  dialog.value = {
    show: true,
    id: p.id,
    form: {
      platform: p.platform,
      name: p.name,
      enabled: p.enabled,
      baidu_token: p.baidu_token || '',
      baidu_page_url: p.baidu_page_url || '',
      baidu_reg_type: p.baidu_reg_type || 68,
      baidu_order_type: p.baidu_order_type || 10,
      e360_key: p.e360_key || '',
      e360_secret: p.e360_secret || '',
      e360_jzqs: p.e360_jzqs || '',
      e360_so_type: p.e360_so_type || '1',
      e360_reg_event: p.e360_reg_event || '',
      e360_order_event: p.e360_order_event || '',
    },
  }
}

async function save() {
  saving.value = true
  try {
    if (dialog.value.id) {
      await ocpcPlatformApi.update(dialog.value.id, dialog.value.form)
    } else {
      await ocpcPlatformApi.create(dialog.value.form)
    }
    dialog.value.show = false
    ElMessage.success('保存成功')
    await load()
  } catch (e) {
    ElMessage.error('保存失败: ' + (e?.response?.data?.error || e.message))
  } finally {
    saving.value = false
  }
}

async function toggle(row) {
  try {
    const res = await ocpcPlatformApi.toggle(row.id)
    row.enabled = res.enabled
    ElMessage.success(res.enabled ? '已启用' : '已停用')
  } catch {
    ElMessage.error('操作失败')
  }
}

async function remove(p) {
  try {
    await ocpcPlatformApi.delete(p.id)
    ElMessage.success('已删除')
    await load()
  } catch (e) {
    ElMessage.error('删除失败')
  }
}

async function upload() {
  uploading.value = true
  result.value = null
  try {
    const res = await ocpcApi.upload()
    result.value = res
    ElMessage.success('上报完成')
  } catch (e) {
    ElMessage.error('上报失败: ' + (e?.response?.data?.error || e.message))
  } finally {
    uploading.value = false
  }
}

async function loadSchedule() {
  scheduleLoading.value = true
  try {
    const res = await ocpcApi.getSchedule()
    const s = res.schedule || {}
    schedule.value.enabled = s.ocpc_schedule_enabled === 'true'
    schedule.value.interval = parseInt(s.ocpc_schedule_interval) || 30
    if (s.ocpc_last_run_at) {
      const ts = parseInt(s.ocpc_last_run_at)
      schedule.value.lastRunAt = ts ? new Date(ts * 1000).toLocaleString('zh-CN') : '—'
    } else {
      schedule.value.lastRunAt = '—'
    }
  } catch {
    ElMessage.error('加载定时配置失败')
  } finally {
    scheduleLoading.value = false
  }
}

async function saveSchedule() {
  scheduleSaving.value = true
  try {
    await ocpcApi.updateSchedule({ enabled: schedule.value.enabled, interval: schedule.value.interval })
    ElMessage.success('定时配置已保存')
  } catch (e) {
    ElMessage.error('保存失败: ' + (e?.response?.data?.error || e.message))
  } finally {
    scheduleSaving.value = false
  }
}

onMounted(() => { load(); loadSchedule() })
</script>

<style scoped>
.ocpc-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Hero card */
.hero-card :deep(.el-card__body) { padding: 24px 28px; }
.hero-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}
.eyebrow {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: .08em;
  text-transform: uppercase;
  color: var(--el-color-primary);
  margin-bottom: 4px;
}
.hero-row h3 {
  margin: 0 0 6px;
  font-size: 18px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}
.hero-row p {
  margin: 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

/* Card header */
.card-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.card-header-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

/* Upload section */
.upload-desc {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin: 0 0 16px;
  line-height: 1.7;
}

/* Monospace cell */
.mono-text {
  font-family: 'SFMono-Regular', Consolas, monospace;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

/* Form tip */
.form-tip {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  margin-top: 4px;
}
</style>
