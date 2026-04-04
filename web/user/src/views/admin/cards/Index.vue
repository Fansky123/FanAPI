<template>
  <div>
    <!-- 生成卡密 -->
    <el-card style="margin-bottom:18px">
      <template #header>批量生成卡密</template>
      <el-form :model="genForm" inline>
        <el-form-item label="数量">
          <el-input-number v-model="genForm.count" :min="1" :max="500" style="width:120px" />
        </el-form-item>
        <el-form-item label="面值（元）">
          <el-input-number v-model="genForm.yuanAmount" :min="0.0001" :precision="4" :step="1" style="width:150px" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="genForm.note" placeholder="可选备注" style="width:160px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="generating" @click="doGenerate">生成</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 卡密列表 -->
    <el-card>
      <template #header>
        <div style="display:flex;justify-content:space-between;align-items:center">
          <span>卡密列表</span>
          <div style="display:flex;gap:8px">
            <el-select v-model="filterStatus" @change="fetchCards(1)" style="width:110px" placeholder="全部状态">
              <el-option value="" label="全部" />
              <el-option value="unused" label="未使用" />
              <el-option value="used" label="已使用" />
            </el-select>
            <el-button @click="fetchCards(page)"><el-icon><Refresh /></el-icon></el-button>
          </div>
        </div>
      </template>

      <el-table :data="cards" stripe>
        <el-table-column prop="code" label="兑换码" min-width="220">
          <template #default="{ row }">
            <el-text class="card-code" @click="copyCode(row.code)" style="cursor:pointer;font-family:monospace">
              {{ row.code }}
            </el-text>
          </template>
        </el-table-column>
        <el-table-column label="面值" width="130">
          <template #default="{ row }">¥{{ (row.credits / 1e6).toFixed(4) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'unused' ? 'success' : 'info'" size="small">
              {{ row.status === 'unused' ? '未使用' : '已使用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="note" label="备注" show-overflow-tooltip />
        <el-table-column label="使用时间" width="160">
          <template #default="{ row }">{{ row.used_at ? fmtTime(row.used_at) : '—' }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="生成时间" width="160" :formatter="(r) => fmtTime(r.created_at)" />
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'unused'"
              type="danger" size="small" text
              @click="doDelete(row)"
            >删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        :page-size="20"
        :total="total"
        style="margin-top:16px"
        @current-change="fetchCards"
      />
    </el-card>

    <!-- 生成结果弹窗 -->
    <el-dialog v-model="showResult" title="生成成功" width="600px">
      <el-alert type="success" :closable="false" style="margin-bottom:12px">
        已生成 {{ lastGenerated.length }} 张卡密，请及时复制保存。
      </el-alert>
      <el-input
        type="textarea"
        :rows="12"
        :value="resultText"
        readonly
      />
      <template #footer>
        <el-button @click="copyResult">复制全部</el-button>
        <el-button type="primary" @click="showResult = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { cardApi } from '@/api/admin'

const cards = ref([])
const total = ref(0)
const page = ref(1)
const filterStatus = ref('')
const generating = ref(false)
const showResult = ref(false)
const lastGenerated = ref([])

const genForm = ref({ count: 10, yuanAmount: 10, note: '' })

const resultText = computed(() =>
  lastGenerated.value.map(c => `${c.code}  ¥${(c.credits / 1e6).toFixed(4)}`).join('\n')
)

onMounted(() => fetchCards(1))

async function fetchCards(p = 1) {
  page.value = p
  const res = await cardApi.list({ page: p, size: 20, status: filterStatus.value })
  cards.value = res.cards ?? []
  total.value = res.total ?? 0
}

async function doGenerate() {
  generating.value = true
  try {
    const res = await cardApi.generate({
      count: genForm.value.count,
      credits: Math.round(genForm.value.yuanAmount * 1e6),
      note: genForm.value.note,
    })
    lastGenerated.value = res.cards ?? []
    showResult.value = true
    fetchCards(1)
  } finally {
    generating.value = false
  }
}

async function doDelete(row) {
  await ElMessageBox.confirm(`确定删除卡密 ${row.code}？`, '确认删除', { type: 'warning' })
  await cardApi.remove(row.id)
  ElMessage.success('已删除')
  fetchCards(page.value)
}

function copyCode(code) {
  navigator.clipboard.writeText(code)
  ElMessage.success('已复制')
}

function copyResult() {
  navigator.clipboard.writeText(resultText.value)
  ElMessage.success('已复制全部卡密')
}

function fmtTime(v) {
  if (!v) return '—'
  return new Date(v).toLocaleString('zh-CN', { hour12: false })
}
</script>

<style scoped>
.card-code:hover { color: var(--el-color-primary) }
</style>
