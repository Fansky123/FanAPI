<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <div class="stat-row">
      <div class="stat-card">
        <div class="stat-val">{{ fmtCredits(store.balance) }}</div>
        <div class="stat-label">剩余积分</div>
      </div>
      <div class="stat-divider" />
      <div class="stat-card">
        <div class="stat-val">{{ fmtCredits(stats.total_consumed) }}</div>
        <div class="stat-label">累计消耗积分</div>
      </div>
      <div class="stat-divider" />
      <div class="stat-card">
        <div class="stat-val">{{ fmtCredits(stats.today_consumed) }}</div>
        <div class="stat-label">今日消耗积分</div>
      </div>
    </div>

    <!-- 快速入门 -->
    <div class="guide-card">
      <div class="guide-title">快速入门步骤：</div>
      <div class="guide-list">
        <div class="guide-item">
          第一步：点击左侧【API 密钥】创建密钥
          <router-link to="/keys" class="guide-link">立即前往</router-link>
        </div>
        <div class="guide-item">
          第二步：点击左侧【模型列表】查看模型 ID 和接口调用地址
          <router-link to="/models" class="guide-link">立即前往</router-link>
        </div>
        <div class="guide-item">
          第三步：点击左侧【文本对话】在线体验所有 AI 聊天模型
          <router-link to="/playground" class="guide-link">立即前往</router-link>
        </div>
        <div class="guide-item">
          第四步：点击左侧【充值积分】充值积分或兑换卡密
          <router-link to="/recharge" class="guide-link">立即前往</router-link>
        </div>
      </div>
    </div>

    <!-- 积分消耗趋势 -->
    <div class="chart-card">
      <div class="chart-title">积分消耗趋势（最近7天）</div>
      <div class="chart-wrap">
        <canvas ref="creditsChart" height="80"></canvas>
        <div ref="creditsTooltip" class="chart-tooltip"></div>
      </div>
    </div>

    <!-- 请求次数统计 -->
    <div class="chart-card">
      <div class="chart-title">请求次数统计（最近7天）</div>
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
import { userApi } from '@/api'

const store = useUserStore()
const stats = ref({ total_consumed: 0, today_consumed: 0, daily_credits: [], daily_requests: [] })
const creditsChart = ref(null)
const reqChart = ref(null)
const creditsTooltip = ref(null)
const reqTooltip = ref(null)

function fmtCredits(v) {
  if (!v) return '0.00'
  return (v / 1e6).toFixed(2)
}

function buildDays(data, key) {
  const days = []
  for (let i = 6; i >= 0; i--) {
    const d = new Date()
    d.setDate(d.getDate() - i)
    const label = `${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
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
  const colors = ['#165dff', '#ff7d00']

  function render(hoverIdx = -1) {
    ctx.clearRect(0, 0, W, H)

    // Grid lines
    ctx.strokeStyle = '#e5e6eb'
    ctx.lineWidth = 1
    ctx.setLineDash([])
    for (let i = 0; i <= 5; i++) {
      const y = pad.top + (cH * i) / 5
      ctx.beginPath(); ctx.moveTo(pad.left, y); ctx.lineTo(pad.left + cW, y); ctx.stroke()
      const val = maxVal * (1 - i / 5)
      ctx.fillStyle = '#86909c'
      ctx.font = '11px sans-serif'
      ctx.textAlign = 'right'
      ctx.fillText(val.toFixed(val < 1 ? 2 : 0), pad.left - 6, y + 4)
    }

    // X labels
    ctx.fillStyle = '#86909c'
    ctx.font = '11px sans-serif'
    ctx.textAlign = 'center'
    labels.forEach((label, i) => {
      const x = pad.left + (i / (labels.length - 1)) * cW
      ctx.fillText(label, x, H - 8)
    })

    // Hover vertical dashed line
    if (hoverIdx >= 0) {
      const hx = pad.left + (hoverIdx / (labels.length - 1)) * cW
      ctx.save()
      ctx.strokeStyle = '#c9cdd4'
      ctx.lineWidth = 1
      ctx.setLineDash([4, 3])
      ctx.beginPath(); ctx.moveTo(hx, pad.top); ctx.lineTo(hx, pad.top + cH); ctx.stroke()
      ctx.restore()
    }

    // Lines + dots
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
      return `<div class="tip-day">${day}</div><div class="tip-row">消耗积分 <b>${val}</b></div>`
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
      return `<div class="tip-day">${day}</div>` +
        `<div class="tip-row"><span class="tip-dot" style="background:#165dff"></span>成功 <b>${reqDays[idx].value}</b></div>` +
        `<div class="tip-row"><span class="tip-dot" style="background:#ff7d00"></span>失败 <b>${failDays[idx].value}</b></div>`
    })
  } catch {
    // 静默失败
  }
}

onMounted(() => {
  if (store.token) store.fetchBalance()
  loadStats()
})
</script>

<style scoped>
.dashboard { max-width: 1100px; display: flex; flex-direction: column; gap: 16px; }

/* 统计卡片 */
.stat-row {
  background: #fff;
  border-radius: 6px;
  border: 1px solid #e5e6eb;
  display: flex;
  align-items: center;
  padding: 0;
  overflow: hidden;
}
.stat-card {
  flex: 1;
  padding: 28px 24px;
  text-align: center;
}
.stat-val { font-size: 1.8rem; font-weight: 700; color: #1d2129; line-height: 1.2; }
.stat-label { margin-top: 8px; font-size: .85rem; color: #86909c; }
.stat-divider { width: 1px; height: 60px; background: #e5e6eb; flex-shrink: 0; }

/* 快速入门 */
.guide-card {
  background: #fff;
  border-radius: 6px;
  border: 1px solid #e5e6eb;
  padding: 20px 24px;
}
.guide-title { font-size: 1.05rem; font-weight: 700; color: var(--ow-primary, #165dff); margin-bottom: 14px; }
.guide-list { display: flex; flex-direction: column; gap: 10px; }
.guide-item { font-size: .9rem; color: #4e5969; }
.guide-link { color: var(--ow-primary, #165dff); text-decoration: none; margin-left: 8px; font-weight: 500; }
.guide-link:hover { text-decoration: underline; }

/* 图表 */
.chart-card {
  background: #fff;
  border-radius: 6px;
  border: 1px solid #e5e6eb;
  padding: 20px 24px;
}
.chart-title { font-size: .95rem; font-weight: 600; color: #1d2129; margin-bottom: 16px; }
canvas { width: 100%; display: block; cursor: crosshair; }
.chart-wrap { position: relative; }
.chart-tooltip {
  display: none;
  position: absolute;
  top: 10px;
  transform: translateX(-50%);
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  padding: 8px 12px;
  font-size: 12px;
  color: #1d2129;
  box-shadow: 0 4px 12px rgba(0,0,0,.1);
  pointer-events: none;
  white-space: nowrap;
  z-index: 10;
  line-height: 1.8;
}
.chart-tooltip :deep(.tip-day) { color: #86909c; font-size: 11px; margin-bottom: 2px; }
.chart-tooltip :deep(.tip-row) { display: flex; align-items: center; gap: 4px; }
.chart-tooltip :deep(.tip-dot) { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.chart-tooltip :deep(b) { margin-left: auto; padding-left: 12px; font-weight: 600; }
</style>
