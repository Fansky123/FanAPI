<template>
  <div class="dashboard">
    <!-- 核心指标卡片 -->
    <el-row :gutter="18" class="cards">
      <el-col :span="6">
        <div class="stat-card">
          <div class="card-label">渠道数量</div>
          <div class="card-value">{{ stats.active_channels ?? '--' }}<span class="card-sub"> / {{ stats.channels ?? '--' }}</span></div>
          <div class="card-tip">活跃 / 全部</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="card-label">用户数量</div>
          <div class="card-value">{{ stats.users ?? '--' }}</div>
          <div class="card-tip">普通用户数</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="card-label">今日收入</div>
          <div class="card-value">¥{{ fmtCredits(stats.today?.revenue) }}</div>
          <div class="card-tip">今日结算 {{ stats.today?.count ?? 0 }} 笔</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card" :class="profitClass(stats.today?.profit)">
          <div class="card-label">今日利润</div>
          <div class="card-value">¥{{ fmtCredits(stats.today?.profit) }}</div>
          <div class="card-tip">收入 - 上游成本</div>
        </div>
      </el-col>
    </el-row>

    <!-- 累计数据 -->
    <el-row :gutter="18" class="cards">
      <el-col :span="8">
        <div class="stat-card alt">
          <div class="card-label">累计营收</div>
          <div class="card-value">¥{{ fmtCredits(stats.total?.revenue) }}</div>
          <div class="card-tip">历史全部结算</div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="stat-card alt">
          <div class="card-label">累计成本</div>
          <div class="card-value">¥{{ fmtCredits(stats.total?.cost) }}</div>
          <div class="card-tip">上游 API 消耗</div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="stat-card alt" :class="profitClass(stats.total?.profit)">
          <div class="card-label">累计利润</div>
          <div class="card-value">¥{{ fmtCredits(stats.total?.profit) }}</div>
          <div class="card-tip">历史净利润（含今日）</div>
        </div>
      </el-col>
    </el-row>

    <!-- 利润率 -->
    <el-row :gutter="18" class="cards" v-if="stats.total?.revenue > 0">
      <el-col :span="24">
        <div class="stat-card margin-row">
          <div class="card-label">综合利润率</div>
          <div class="card-value">{{ marginPct }}%</div>
          <div class="card-tip">累计利润 ÷ 累计营收</div>
          <el-progress
            :percentage="Math.min(Math.max(parseFloat(marginPct), 0), 100)"
            :color="parseFloat(marginPct) >= 0 ? '#52c41a' : '#ff4d4f'"
            :show-text="false"
            style="margin-top:12px"
          />
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { statsApi } from '@/api/admin'

const stats = ref({})

async function loadStats() {
  try {
    stats.value = await statsApi.get()
  } catch {
    // 后端未就绪时静默失败
  }
}

onMounted(loadStats)

// credits 单位为 1/1000000 元（微元）
function fmtCredits(v) {
  if (v == null) return '--'
  return (v / 1000000).toFixed(4)
}

function profitClass(v) {
  if (v == null) return ''
  return v >= 0 ? 'profit-pos' : 'profit-neg'
}

const marginPct = computed(() => {
  const r = stats.value.total?.revenue
  const p = stats.value.total?.profit
  if (!r || r === 0) return '0.00'
  return ((p / r) * 100).toFixed(2)
})
</script>

<style scoped>
.dashboard { padding: 4px 0 }
.cards { margin-bottom: 18px }
.stat-card {
  background: #fff;
  border-radius: 14px;
  padding: 24px 28px;
  box-shadow: 0 2px 12px rgba(0,0,0,.07);
  min-height: 110px;
}
.stat-card.alt {
  background: linear-gradient(135deg,#f6f9ff,#eef3ff);
}
.stat-card.profit-pos { border-left: 4px solid #52c41a }
.stat-card.profit-neg { border-left: 4px solid #ff4d4f }
.card-label { font-size: .84rem; color: #888; margin-bottom: 6px }
.card-value { font-size: 1.8rem; font-weight: 700; color: #1a2540; line-height: 1.15 }
.card-sub { font-size: 1.1rem; color: #aaa; font-weight: 400 }
.card-tip { font-size: .78rem; color: #bbb; margin-top: 4px }
.margin-row { margin-top: 0 }
</style>
