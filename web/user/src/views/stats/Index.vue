<template>
  <div class="stats-page">
    <div class="page-title-header">
      <h2>使用统计</h2>
    </div>

    <div class="content-card stat-row">
      <div class="stat-card">
        <div class="stat-val">{{ fmtCredits(stats.total_consumed) }}</div>
        <div class="stat-label">累计消耗积分</div>
      </div>
      <div class="stat-divider" />
      <div class="stat-card">
        <div class="stat-val">{{ fmtCredits(stats.today_consumed) }}</div>
        <div class="stat-label">今日消耗积分</div>
      </div>
      <div class="stat-divider" />
      <div class="stat-card">
        <div class="stat-val">{{ totalRequests }}</div>
        <div class="stat-label">累计请求次数</div>
      </div>
    </div>

    <div class="content-card">
      <div class="chart-title">积分消耗趋势（最近7天）</div>
      <div class="chart-wrap">
        <canvas ref="creditsChart" height="100"></canvas>
        <div ref="creditsTooltip" class="chart-tooltip"></div>
      </div>
    </div>

    <div class="content-card">
      <div class="chart-title">请求次数统计（最近7天）</div>
      <div class="chart-legend">
        <span class="legend-dot" style="background:#165dff"></span>成功次数
        <span class="legend-dot" style="background:#ff7d00;margin-left:16px"></span>失败次数
      </div>
      <div class="chart-wrap">
        <canvas ref="reqChart" height="100"></canvas>
        <div ref="reqTooltip" class="chart-tooltip"></div>
      </div>
    </div>

    <div class="content-card">
      <div class="chart-title">每日明细</div>
      <el-table :data="dailyTable" stripe border size="default">
        <el-table-column prop="day" label="日期" width="120" />
        <el-table-column label="消耗积分" width="140">
          <template #default="{ row }">
            <span style="color:#165dff;font-weight:600">{{ fmtCredits(row.credits) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="成功请求" width="120">
          <template #default="{ row }">
            <el-tag type="success" size="small" effect="plain">{{ row.success }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="失败请求" width="120">
          <template #default="{ row }">
            <el-tag type="danger" size="small" effect="plain" v-if="row.failed > 0">{{ row.failed }}</el-tag>
            <span v-else style="color:#86909c">0</span>
          </template>
        </el-table-column>
        <el-table-column label="成功率">
          <template #default="{ row }">
            <el-progress
              :percentage="row.rate"
              :stroke-width="6"
              :show-text="true"
              :color="row.rate >= 90 ? '#00b42a' : row.rate >= 70 ? '#ff7d00' : '#f53f3f'"
            />
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { userApi } from '@/api'

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

const dailyTable = computed(() => {
  return buildDays(stats.value.daily_credits, 'credits').map((c) => {
    const reqEntry = (stats.value.daily_requests || []).find(r => r.day === c.label) || {}
    const success = reqEntry.success ?? 0
    const failed = reqEntry.failed ?? 0
    const total = success + failed
    return {
      day: c.label,
      credits: c.value,
      success,
      failed,
      rate: total > 0 ? Math.round((success / total) * 100) : 100,
    }
  })
})

const totalRequests = computed(() => {
  return dailyTable.value.reduce((s, r) => s + r.success + r.failed, 0)
})

const successRate = computed(() => {
  const total = dailyTable.value.reduce((s, r) => s + r.success + r.failed, 0)
  if (!total) return 100
  const succ = dailyTable.value.reduce((s, r) => s + r.success, 0)
  return Math.round((succ / total) * 100)
})

function drawLineChart(canvas, datasets, labels) {
  if (!canvas) return null
  const ctx = canvas.getContext('2d')
  const W = canvas.offsetWidth || 800
  const H = canvas.offsetHeight || 140
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
    ctx.strokeStyle = '#e5e6eb'; ctx.lineWidth = 1; ctx.setLineDash([])
    for (let i = 0; i <= 5; i++) {
      const y = pad.top + (cH * i) / 5
      ctx.beginPath(); ctx.moveTo(pad.left, y); ctx.lineTo(pad.left + cW, y); ctx.stroke()
      const val = maxVal * (1 - i / 5)
      ctx.fillStyle = '#86909c'; ctx.font = '11px sans-serif'; ctx.textAlign = 'right'
      ctx.fillText(val.toFixed(val < 1 ? 2 : 0), pad.left - 6, y + 4)
    }
    ctx.fillStyle = '#86909c'; ctx.font = '11px sans-serif'; ctx.textAlign = 'center'
    labels.forEach((label, i) => {
      const x = pad.left + (i / (labels.length - 1)) * cW
      ctx.fillText(label, x, H - 8)
    })
    if (hoverIdx >= 0) {
      const hx = pad.left + (hoverIdx / (labels.length - 1)) * cW
      ctx.save(); ctx.strokeStyle = '#c9cdd4'; ctx.lineWidth = 1; ctx.setLineDash([4, 3])
      ctx.beginPath(); ctx.moveTo(hx, pad.top); ctx.lineTo(hx, pad.top + cH); ctx.stroke(); ctx.restore()
    }
    datasets.forEach((ds, di) => {
      ctx.strokeStyle = colors[di] || '#165dff'; ctx.lineWidth = 2; ctx.setLineDash([])
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
    const s = stateHolder.state; if (!s) return
    const { pad, cW, W, labels, render } = s
    const rect = canvas.getBoundingClientRect()
    const mx = (e.clientX - rect.left) * (canvas.width / rect.width)
    const n = labels.length
    let idx = Math.round((mx - pad.left) / cW * (n - 1))
    idx = Math.max(0, Math.min(n - 1, idx))
    render(idx)
    tooltipEl.style.display = 'block'
    tooltipEl.style.left = ((pad.left + (idx / (n - 1)) * cW) / W * 100) + '%'
    tooltipEl.innerHTML = formatter(idx)
  })
  canvas.addEventListener('mouseleave', () => {
    const s = stateHolder.state; if (s) s.render(-1)
    tooltipEl.style.display = 'none'
  })
}

async function load() {
  try {
    const res = await userApi.getStats()
    stats.value = res
    await nextTick()

    const creditDays = buildDays(res.daily_credits, 'credits')
    const cs = { state: null }
    cs.state = drawLineChart(creditsChart.value, [{ values: creditDays.map(d => d.value / 1e6) }], creditDays.map(d => d.label))
    setupTooltip(creditsChart.value, creditsTooltip.value, cs, (idx) =>
      `<div class="tip-day">${creditDays[idx].label}</div><div class="tip-row">消耗积分 <b>${fmtCredits(creditDays[idx].value)}</b></div>`)

    const reqDays = buildDays(res.daily_requests, 'success')
    const failDays = buildDays(res.daily_requests, 'failed')
    const rs = { state: null }
    rs.state = drawLineChart(reqChart.value, [
      { values: reqDays.map(d => d.value) },
      { values: failDays.map(d => d.value) },
    ], reqDays.map(d => d.label))
    setupTooltip(reqChart.value, reqTooltip.value, rs, (idx) =>
      `<div class="tip-day">${reqDays[idx].label}</div>
       <div class="tip-row"><span class="tip-dot" style="background:#165dff"></span>成功 <b>${reqDays[idx].value}</b></div>
       <div class="tip-row"><span class="tip-dot" style="background:#ff7d00"></span>失败 <b>${failDays[idx].value}</b></div>`)
  } catch {}
}

onMounted(load)
</script>

<style scoped>
.stats-page { display: flex; flex-direction: column; }

.page-title-header {
  padding: 15px 24px;
  border-radius: 12px;
  background: #ffffff;
  border: 1px solid #ffffff;
  box-shadow: rgba(0,0,0,0.02) 0px 10px 20px 0px;
  margin-bottom: 15px;
}
.page-title-header h2 { margin: 0; font-size: 20px; font-weight: 600; color: rgb(26, 27, 28); }

.content-card { background: #ffffff; border-radius: 12px; padding: 20px; margin-bottom: 15px; }

.stat-row { padding: 0; overflow: hidden; display: flex; align-items: center; }
.stat-card { flex: 1; padding: 40px 25px; text-align: center; display: flex; flex-direction: column; align-items: center; }
.stat-val { font-size: 30px; font-weight: 700; color: rgb(0,0,0); line-height: 1.2; }
.stat-label { margin-top: 8px; font-size: 15px; color: rgb(101, 101, 101); }
.stat-divider { width: 1px; height: 60px; background: #e5e6eb; flex-shrink: 0; }

.chart-title { font-size: 16px; font-weight: 600; color: #1d2129; margin-bottom: 12px; }
.chart-legend { font-size: 12px; color: #4e5969; margin-bottom: 10px; display: flex; align-items: center; gap: 4px; }
.legend-dot { display: inline-block; width: 10px; height: 10px; border-radius: 50%; }
.chart-wrap { position: relative; }
canvas { width: 100%; display: block; cursor: crosshair; }
.chart-tooltip {
  display: none; position: absolute; top: 10px; transform: translateX(-50%);
  background: #fff; border: 1px solid #e5e6eb; border-radius: 6px;
  padding: 8px 12px; font-size: 12px; color: #1d2129;
  box-shadow: 0 4px 12px rgba(0,0,0,.1); pointer-events: none;
  white-space: nowrap; z-index: 10; line-height: 1.8;
}
.chart-tooltip :deep(.tip-day) { color: #86909c; font-size: 11px; margin-bottom: 2px; }
.chart-tooltip :deep(.tip-row) { display: flex; align-items: center; gap: 4px; }
.chart-tooltip :deep(.tip-dot) { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.chart-tooltip :deep(b) { margin-left: auto; padding-left: 12px; font-weight: 600; }
</style>
