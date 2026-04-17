<template>
  <div class="dashboard">
    <!-- Row 1: Stats card -->
    <div class="summary-card">
      <div class="summary-item">
        <div class="summary-item-new">{{ fmtCredits(store.balance) }}</div>
        <div class="summary-item-title">剩余积分</div>
      </div>
      <div class="summary-divider" />
      <div class="summary-item">
        <div class="summary-item-new">{{ fmtCredits(stats.total_consumed) }}</div>
        <div class="summary-item-title">累计消耗积分</div>
      </div>
      <div class="summary-divider" />
      <div class="summary-item">
        <div class="summary-item-new">{{ fmtCredits(stats.today_consumed) }}</div>
        <div class="summary-item-title">今日消耗积分</div>
      </div>
    </div>

    <!-- Row 2: Guide card -->
    <div class="novice-guide-card">
      <div class="tips-bar">
        <span class="tips-icon">&#x26A0;</span>
        <span><b>Tips：</b>使用的过程中遇到任何问题，可以添加 QQ 交流群进行咨询：<b style="color:rgb(0,124,255)">{{ siteStore.contactInfo || '1022415589' }}</b>，我们会尽快回答你的问题。</span>
      </div>

      <div class="guide-header">快速入门步骤：</div>
      <div class="guide-steps">
        <div class="guide-step">
          <span>第一步：点击左侧【API 密钥】创建密钥</span>
          <router-link to="/keys" class="guide-link">立即前往</router-link>
        </div>
        <div class="guide-step">
          <span>第二步：点击左侧【模型列表】点击模型卡片查看模型ID和接口调用地址</span>
          <router-link to="/models" class="guide-link">立即前往</router-link>
        </div>
        <div class="guide-step">
          <span>第三步：点击左侧【文本对话】可以在线体验所有的AI聊天模型</span>
          <router-link to="/playground" class="guide-link">立即前往</router-link>
        </div>
        <div class="guide-step">
          <span>第四步：点击左侧【图片生成】可以在线体验所有的图片生成模型</span>
          <router-link to="/image-gen" class="guide-link">立即前往</router-link>
        </div>
        <div class="guide-step">
          <span>第五步：点击左侧【积分充值】充值积分</span>
          <router-link to="/recharge" class="guide-link">立即前往</router-link>
        </div>
      </div>

      <div class="tips-bar">
        <span class="tips-icon">&#x26A0;</span>
        <span><b>Tips：</b>本站大模型接口网关：<b style="color:rgb(0,124,255)">{{ currentHost }}</b>，将模型基址替换为接口网关，完全兼容 OpenAI 协议。</span>
      </div>
    </div>

    <!-- Row 3: Credits trend chart -->
    <div class="summary-count-card">
      <div class="chart-title">积分消耗趋势</div>
      <div class="chart-wrap">
        <canvas ref="creditsChart" height="80"></canvas>
        <div ref="creditsTooltip" class="chart-tooltip"></div>
      </div>
    </div>

    <!-- Row 4: Requests chart -->
    <div class="cash-trend-card">
      <div class="chart-title">请求次数统计</div>
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
  const colors = ['#165dff', '#ff7d00']

  function render(hoverIdx = -1) {
    ctx.clearRect(0, 0, W, H)

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

    ctx.fillStyle = '#86909c'
    ctx.font = '11px sans-serif'
    ctx.textAlign = 'center'
    labels.forEach((label, i) => {
      const x = pad.left + (i / (labels.length - 1)) * cW
      ctx.fillText(label, x, H - 8)
    })

    if (hoverIdx >= 0) {
      const hx = pad.left + (hoverIdx / (labels.length - 1)) * cW
      ctx.save()
      ctx.strokeStyle = '#c9cdd4'
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
        '<div class="tip-row"><span class="tip-dot" style="background:#165dff"></span>成功 <b>' + reqDays[idx].value + '</b></div>' +
        '<div class="tip-row"><span class="tip-dot" style="background:#ff7d00"></span>失败 <b>' + failDays[idx].value + '</b></div>'
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
}

/* Row 1: Stats card */
.summary-card {
  background: #fff;
  border-radius: 12px;
  display: flex;
  align-items: center;
  margin-bottom: 15px;
  overflow: hidden;
}
.summary-item {
  flex: 1;
  padding: 40px 25px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.summary-item-new {
  font-size: 30px;
  font-weight: 700;
  color: #1d2129;
  line-height: 1.2;
}
.summary-item-title {
  margin-top: 8px;
  font-size: 15px;
  color: rgb(101, 101, 101);
}
.summary-divider {
  width: 1px;
  height: 60px;
  background: #e5e6eb;
  flex-shrink: 0;
}

/* Row 2: Guide card */
.novice-guide-card {
  background: #fff;
  border-radius: 12px;
  margin-bottom: 15px;
  padding: 15px;
}
.tips-bar {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  background: rgb(255, 251, 235);
  border: 1px solid rgb(253, 230, 138);
  color: rgb(146, 64, 14);
  padding: 10px;
  border-radius: 4px;
  font-size: 14px;
  line-height: 1.6;
}
.tips-icon {
  color: rgb(217, 119, 6);
  flex-shrink: 0;
  font-size: 15px;
  margin-top: 1px;
}
.guide-header {
  font-size: 23px;
  font-weight: 700;
  color: rgb(22, 93, 255);
  margin: 15px 0 10px;
}
.guide-steps {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 15px;
}
.guide-step {
  font-size: 15px;
  color: #1d2129;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.guide-link {
  color: rgb(0, 124, 255);
  text-decoration: none;
  flex-shrink: 0;
  margin-left: 12px;
  font-weight: 500;
}
.guide-link:hover { text-decoration: underline; }

/* Row 3 & 4: Chart cards */
.summary-count-card {
  background: #fff;
  border-radius: 12px;
  padding: 8px;
  margin-bottom: 15px;
}
.cash-trend-card {
  background: #fff;
  border-radius: 12px;
  padding: 8px;
}
.chart-title {
  font-size: 18px;
  font-weight: bold;
  color: #1d2129;
  margin-bottom: 16px;
  padding: 8px 8px 0;
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
