<template>
  <div class="dashboard">
    <!-- Row 1: Stats cards -->
    <div class="stats-grid">
      <div class="stat-card stat-blue">
        <div class="stat-body">
          <div class="stat-value">{{ fmtCredits(store.balance) }}</div>
          <div class="stat-label">剩余积分</div>
        </div>
        <div class="stat-icon-wrap stat-icon-blue">
          <svg width="22" height="22" fill="none" viewBox="0 0 24 24"><path d="M21 7H3a1 1 0 0 0-1 1v12a1 1 0 0 0 1 1h18a1 1 0 0 0 1-1V8a1 1 0 0 0-1-1ZM3 5h15a1 1 0 0 0 0-2H3a1 1 0 0 0 0 2Zm13 9a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z" fill="currentColor"/></svg>
        </div>
      </div>
      <div class="stat-card stat-green">
        <div class="stat-body">
          <div class="stat-value">{{ fmtCredits(stats.total_consumed) }}</div>
          <div class="stat-label">累计消耗积分</div>
        </div>
        <div class="stat-icon-wrap stat-icon-green">
          <svg width="22" height="22" fill="none" viewBox="0 0 24 24"><path d="M3 3v16a2 2 0 0 0 2 2h16" stroke="currentColor" stroke-width="2" stroke-linecap="round"/><path d="m7 16 4-5 4 3 4-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </div>
      </div>
      <div class="stat-card stat-orange">
        <div class="stat-body">
          <div class="stat-value">{{ fmtCredits(stats.today_consumed) }}</div>
          <div class="stat-label">今日消耗积分</div>
        </div>
        <div class="stat-icon-wrap stat-icon-orange">
          <svg width="22" height="22" fill="none" viewBox="0 0 24 24"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2"/><path d="M12 7v5l3 3" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
        </div>
      </div>
    </div>

    <!-- Row 2: Guide card -->
    <div class="guide-card">
      <div class="tips-bar">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" style="flex-shrink:0;margin-top:1px"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2"/><path d="M12 8v5M12 16v.5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
        <span><b>Tips：</b>使用的过程中遇到任何问题，可以添加 QQ 交流群进行咨询：<b style="color:var(--ow-primary)">{{ siteStore.contactInfo || '1022415589' }}</b>，我们会尽快回答你的问题。</span>
      </div>

      <div class="guide-header">快速入门步骤</div>
      <div class="guide-steps">
        <div class="guide-step">
          <div class="step-num">1</div>
          <span>点击左侧【API 密钥】创建密钥</span>
          <router-link to="/keys" class="guide-link">立即前往 →</router-link>
        </div>
        <div class="guide-step">
          <div class="step-num">2</div>
          <span>点击左侧【模型列表】查看模型 ID 和接口调用地址</span>
          <router-link to="/models" class="guide-link">立即前往 →</router-link>
        </div>
        <div class="guide-step">
          <div class="step-num">3</div>
          <span>点击左侧【文本对话】在线体验所有 AI 聊天模型</span>
          <router-link to="/playground" class="guide-link">立即前往 →</router-link>
        </div>
        <div class="guide-step">
          <div class="step-num">4</div>
          <span>点击左侧【图片生成】在线体验所有图片生成模型</span>
          <router-link to="/image-gen" class="guide-link">立即前往 →</router-link>
        </div>
        <div class="guide-step">
          <div class="step-num">5</div>
          <span>点击左侧【积分充值】充值积分</span>
          <router-link to="/recharge" class="guide-link">立即前往 →</router-link>
        </div>
      </div>

      <div class="tips-bar" style="margin-top:4px">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" style="flex-shrink:0;margin-top:1px"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2"/><path d="M12 8v5M12 16v.5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
        <span><b>Tips：</b>本站大模型接口网关：<b style="color:var(--ow-primary)">{{ currentHost }}</b>，将模型基址替换为接口网关，完全兼容 OpenAI 协议。</span>
      </div>
    </div>

    <!-- Row 3: Credits trend chart -->
    <div class="chart-card">
      <div class="chart-card-header">积分消耗趋势</div>
      <div class="chart-wrap">
        <canvas ref="creditsChart" height="80"></canvas>
        <div ref="creditsTooltip" class="chart-tooltip"></div>
      </div>
    </div>

    <!-- Row 4: Requests chart -->
    <div class="chart-card">
      <div class="chart-card-header">请求次数统计</div>
      <div class="chart-wrap">
        <canvas ref="reqChart" height="80"></canvas>
        <div ref="reqTooltip" class="chart-tooltip"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { useUserStore } from '@/stores/user'
import { useSiteStore } from '@/stores/site'
import { userApi } from '@/api'

const store = useUserStore()
const siteStore = useSiteStore()
const stats = ref({ total_consumed: 0, today_consumed: 0, daily_credits: [], daily_requests: [] })
const creditsChart = ref(null)
const reqChart = ref(null)
const creditsTooltip = ref(null)
const reqTooltip = ref(null)
const currentHost = window.location.host

function fmtCredits(v) {
  if (!v) return '0.00'
  return (v / 1e6).toFixed(2)
}

function buildDays(data, key) {
  const days = []
  for (let i = 6; i >= 0; i--) {
    const d = new Date()
    d.setDate(d.getDate() - i)
    const label = String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
    const found = (data || []).find(r => r.day === label)
    days.push({ label, value: found ? (found[key] ?? 0) : 0 })
  }
  return days
}

function drawLineChart(canvas, datasets, labels) {
  if (!canvas) return null
  const ctx = canvas.getContext('2d')
  const W = canvas.offsetWidth || 800
  const H = canvas.offsetHeight || 120
  canvas.width = W
  canvas.height = H

  const pad = { top: 16, right: 16, bottom: 32, left: 52 }
  const cW = W - pad.left - pad.right
  const cH = H - pad.top - pad.bottom
  const allVals = datasets.flatMap(d => d.values)
  const maxVal = Math.max(...allVals, 1)
  const colors = ['#2563eb', '#f59e0b']

  function render(hoverIdx = -1) {
    ctx.clearRect(0, 0, W, H)

    ctx.strokeStyle = '#e2e8f0'
    ctx.lineWidth = 1
    ctx.setLineDash([])
    for (let i = 0; i <= 5; i++) {
      const y = pad.top + (cH * i) / 5
      ctx.beginPath(); ctx.moveTo(pad.left, y); ctx.lineTo(pad.left + cW, y); ctx.stroke()
      const val = maxVal * (1 - i / 5)
      ctx.fillStyle = '#94a3b8'
      ctx.font = '11px system-ui, sans-serif'
      ctx.textAlign = 'right'
      ctx.fillText(val.toFixed(val < 1 ? 2 : 0), pad.left - 6, y + 4)
    }

    ctx.fillStyle = '#94a3b8'
    ctx.font = '11px system-ui, sans-serif'
    ctx.textAlign = 'center'
    labels.forEach((label, i) => {
      const x = pad.left + (i / (labels.length - 1)) * cW
      ctx.fillText(label, x, H - 8)
    })

    if (hoverIdx >= 0) {
      const hx = pad.left + (hoverIdx / (labels.length - 1)) * cW
      ctx.save()
      ctx.strokeStyle = '#cbd5e1'
      ctx.lineWidth = 1
      ctx.setLineDash([4, 3])
      ctx.beginPath(); ctx.moveTo(hx, pad.top); ctx.lineTo(hx, pad.top + cH); ctx.stroke()
      ctx.restore()
    }

    datasets.forEach((ds, di) => {
      ctx.strokeStyle = colors[di] || '#165dff'
      ctx.lineWidth = 2
      ctx.setLineDash([])
      ctx.beginPath()
      ds.values.forEach((v, i) => {
        const x = pad.left + (i / (labels.length - 1)) * cW
        const y = pad.top + cH * (1 - v / maxVal)
        i === 0 ? ctx.moveTo(x, y) : ctx.lineTo(x, y)
      })
      ctx.stroke()

      ctx.fillStyle = colors[di] || '#165dff'
      ds.values.forEach((v, i) => {
        const x = pad.left + (i / (labels.length - 1)) * cW
        const y = pad.top + cH * (1 - v / maxVal)
        ctx.beginPath(); ctx.arc(x, y, i === hoverIdx ? 5 : 3, 0, Math.PI * 2); ctx.fill()
      })
    })
  }

  render()
  return { pad, cW, W, H, labels, datasets, render }
}

function setupTooltip(canvas, tooltipEl, stateHolder, formatter) {
  canvas.addEventListener('mousemove', (e) => {
    const s = stateHolder.state
    if (!s) return
    const { pad, cW, W, labels, render } = s
    const rect = canvas.getBoundingClientRect()
    const mx = (e.clientX - rect.left) * (canvas.width / rect.width)
    const n = labels.length
    let idx = Math.round((mx - pad.left) / cW * (n - 1))
    idx = Math.max(0, Math.min(n - 1, idx))
    render(idx)

    const xPct = (pad.left + (idx / (n - 1)) * cW) / W * 100
    tooltipEl.style.display = 'block'
    tooltipEl.style.left = xPct + '%'
    tooltipEl.innerHTML = formatter(idx)
  })
  canvas.addEventListener('mouseleave', () => {
    const s = stateHolder.state
    if (s) s.render(-1)
    tooltipEl.style.display = 'none'
  })
}

async function loadStats() {
  if (!store.token) return
  try {
    const res = await userApi.getStats()
    stats.value = res
    await nextTick()

    const creditDays = buildDays(res.daily_credits, 'credits')
    const creditsState = { state: null }
    creditsState.state = drawLineChart(creditsChart.value, [
      { values: creditDays.map(d => d.value / 1e6) }
    ], creditDays.map(d => d.label))
    setupTooltip(creditsChart.value, creditsTooltip.value, creditsState, (idx) => {
      const day = creditDays[idx].label
      const val = fmtCredits(creditDays[idx].value)
      return '<div class="tip-day">' + day + '</div><div class="tip-row">消耗积分 <b>' + val + '</b></div>'
    })

    const reqDays = buildDays(res.daily_requests, 'success')
    const failDays = buildDays(res.daily_requests, 'failed')
    const reqState = { state: null }
    reqState.state = drawLineChart(reqChart.value, [
      { values: reqDays.map(d => d.value), label: '成功' },
      { values: failDays.map(d => d.value), label: '失败' }
    ], reqDays.map(d => d.label))
    setupTooltip(reqChart.value, reqTooltip.value, reqState, (idx) => {
      const day = reqDays[idx].label
      return '<div class="tip-day">' + day + '</div>' +
        '<div class="tip-row"><span class="tip-dot" style="background:#2563eb"></span>成功 <b>' + reqDays[idx].value + '</b></div>' +
        '<div class="tip-row"><span class="tip-dot" style="background:#f59e0b"></span>失败 <b>' + failDays[idx].value + '</b></div>'
    })
  } catch {
    // silent
  }
}

onMounted(() => {
  if (store.token) store.fetchBalance()
  loadStats()
})
</script>

<style scoped>
.dashboard {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* ── Stats grid ── */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}
.stat-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px 22px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 3px rgba(0,0,0,.06);
  border: 1px solid var(--ow-border, #e2e8f0);
  gap: 16px;
}
.stat-body { flex: 1; min-width: 0; }
.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--ow-text, #0f172a);
  letter-spacing: -.02em;
  line-height: 1.2;
}
.stat-label {
  margin-top: 6px;
  font-size: 13px;
  color: var(--ow-subtext, #94a3b8);
}
.stat-icon-wrap {
  width: 48px; height: 48px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
}
.stat-icon-blue  { background: #dbeafe; color: #2563eb; }
.stat-icon-green { background: #d1fae5; color: #10b981; }
.stat-icon-orange{ background: #fef3c7; color: #f59e0b; }

/* ── Guide card ── */
.guide-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,.06);
  border: 1px solid var(--ow-border, #e2e8f0);
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.tips-bar {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  color: #92400e;
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 13.5px;
  line-height: 1.6;
}
.guide-header {
  font-size: 15px;
  font-weight: 700;
  color: var(--ow-text, #0f172a);
  padding-top: 4px;
}
.guide-steps {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.guide-step {
  font-size: 13.5px;
  color: var(--ow-text, #0f172a);
  display: flex;
  align-items: center;
  gap: 12px;
}
.step-num {
  width: 22px; height: 22px;
  border-radius: 50%;
  background: var(--ow-primary-bg, #eff6ff);
  color: var(--ow-primary, #2563eb);
  display: grid;
  place-items: center;
  font-size: 11px;
  font-weight: 700;
  flex-shrink: 0;
}
.guide-link {
  color: var(--ow-primary, #2563eb);
  text-decoration: none;
  flex-shrink: 0;
  margin-left: auto;
  font-size: 13px;
  font-weight: 500;
  transition: opacity .15s;
}
.guide-link:hover { opacity: .75; }

/* ── Chart cards ── */
.chart-card {
  background: #fff;
  border-radius: 12px;
  padding: 16px 16px 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,.06);
  border: 1px solid var(--ow-border, #e2e8f0);
}
.chart-card-header {
  font-size: 15px;
  font-weight: 600;
  color: var(--ow-text, #0f172a);
  margin-bottom: 12px;
  padding: 0 4px;
}
canvas {
  width: 100%;
  display: block;
  cursor: crosshair;
}
.chart-wrap { position: relative; }
.chart-tooltip {
  display: none;
  position: absolute;
  top: 10px;
  transform: translateX(-50%);
  background: #fff;
  border: 1px solid var(--ow-border, #e2e8f0);
  border-radius: 8px;
  padding: 8px 12px;
  font-size: 12px;
  color: var(--ow-text, #0f172a);
  box-shadow: 0 4px 16px rgba(0,0,0,.1);
  pointer-events: none;
  white-space: nowrap;
  z-index: 10;
  line-height: 1.8;
}
.chart-tooltip :deep(.tip-day) { color: var(--ow-subtext); font-size: 11px; margin-bottom: 2px; }
.chart-tooltip :deep(.tip-row) { display: flex; align-items: center; gap: 4px; }
.chart-tooltip :deep(.tip-dot) { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.chart-tooltip :deep(b) { margin-left: auto; padding-left: 12px; font-weight: 600; }

@media (max-width: 640px) {
  .stats-grid { grid-template-columns: 1fr; }
}
</style>
