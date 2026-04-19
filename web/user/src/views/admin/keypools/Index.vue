<template>
  <div class="keypools-page">
    <!-- Hero Card -->
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">Key Pools</div>
          <h3>号池管理</h3>
          <p>管理三方 API Key 集合。同一个号池中的 Key 按 Sticky 策略分配给用户，配额耗尽时自动轮转到下一个 Key，保证上下文连续性与缓存命中率。</p>
        </div>
        <el-button type="primary" @click="openCreatePool">
          <el-icon><Plus /></el-icon> 新建号池
        </el-button>
      </div>
    </el-card>

    <!-- 号池列表 -->
    <el-card>
      <el-table :data="pools" stripe border v-loading="loadingPools">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="号池名称" min-width="160" />
        <el-table-column prop="channel_id" label="渠道 ID" width="100" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.channel_id }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_active" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'danger'" size="small">
              {{ row.is_active ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="号商上传" width="95" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.vendor_submittable ? 'success' : 'info'"
              size="small"
              style="cursor:pointer"
              @click="toggleVendorSubmittable(row)"
            >{{ row.vendor_submittable ? '已开放' : '关闭' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" align="center">
          <template #default="{ row }">
            <el-button size="small" type="primary" plain @click="openKeyDrawer(row)">
              <el-icon><Key /></el-icon> 管理 Keys
            </el-button>
            <el-button
              size="small"
              :type="row.is_active ? 'warning' : 'success'"
              plain
              @click="togglePool(row)"
            >{{ row.is_active ? '停用' : '启用' }}</el-button>
            <el-popconfirm title="确认删除此号池及其所有 Key？" @confirm="deletePool(row.id)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新建号池弹窗 -->
    <el-dialog v-model="createPoolVisible" title="新建号池" width="440px">
      <el-form :model="poolForm" label-width="100px">
        <el-form-item label="渠道 ID" required>
          <el-input-number v-model="poolForm.channel_id" :min="1" style="width:100%" />
          <div style="font-size:11px;color:#aaa;margin-top:4px">渠道管理页面中的 ID</div>
        </el-form-item>
        <el-form-item label="号池名称" required>
          <el-input v-model="poolForm.name" placeholder="例：ChatFire GPT-4 Pool" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createPoolVisible = false">取消</el-button>
        <el-button type="primary" @click="submitCreatePool">创建</el-button>
      </template>
    </el-dialog>

    <!-- Keys 管理抽屉 -->
    <el-drawer
      v-model="keyDrawerVisible"
      :title="`${activePool?.name || ''} — Key 管理`"
      size="600px"
      direction="rtl"
    >
      <div class="drawer-toolbar">
        <span style="color:#6b7a90;font-size:13px">
          共 {{ keys.length }} 个 Key，Sticky 策略：用户首次请求绑定，配额耗尽后自动轮转
        </span>
        <el-button type="primary" size="small" @click="openAddKey">
          <el-icon><Plus /></el-icon> 添加 Key
        </el-button>
      </div>

      <el-table :data="keys" stripe border v-loading="loadingKeys" style="margin-top:12px">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column label="Key Value" min-width="220">
          <template #default="{ row }">
            <div class="key-value">
              <el-icon style="color:#aaa;margin-right:4px"><Lock /></el-icon>
              <span style="font-family:monospace;letter-spacing:1px">
                ••••••••{{ maskKey(row.value) }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="80" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.priority }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_active" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'danger'" size="small">
              {{ row.is_active ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="130" align="center">
          <template #default="{ row }">
            <el-button size="small" type="primary" plain @click="openEditKey(row)">
              编辑
            </el-button>
            <el-popconfirm title="移除此 Key？" @confirm="removeKey(row.id)">
              <template #reference>
                <el-button size="small" type="danger" circle>
                  <el-icon><Delete /></el-icon>
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="!loadingKeys && keys.length === 0" class="empty-tip">
        <el-empty description="还没有 Key，点击右上角添加" :image-size="80" />
      </div>
    </el-drawer>

    <!-- 编辑 Key 弹窗 -->
    <el-dialog v-model="editKeyVisible" title="编辑 Key" width="400px">
      <el-form :model="editKeyForm" label-width="80px">
        <el-form-item label="优先级">
          <el-input-number v-model="editKeyForm.priority" :min="0" :max="999" />
          <span style="margin-left:8px;color:#aaa;font-size:12px">数字越小越优先</span>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch
            v-model="editKeyForm.is_active"
            active-text="启用"
            inactive-text="停用"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editKeyVisible = false">取消</el-button>
        <el-button type="primary" @click="submitEditKey">保存</el-button>
      </template>
    </el-dialog>

    <!-- 添加 Key 弹窗 -->
    <el-dialog v-model="addKeyVisible" title="添加 Key" width="480px">
      <el-form :model="keyForm" label-width="100px">
        <el-form-item label="Key Value" required>
          <el-input
            v-model="keyForm.value"
            type="textarea"
            :rows="3"
            placeholder="sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
            style="font-family:monospace;font-size:13px"
          />
          <div style="font-size:11px;color:#aaa;margin-top:4px">
            完整的三方 API Key，保存后将被加密存储并脱敏展示
          </div>
        </el-form-item>
        <el-form-item label="优先级">
          <el-input-number v-model="keyForm.priority" :min="0" :max="999" />
          <span style="margin-left:8px;color:#aaa;font-size:12px">数字越小越优先分配</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addKeyVisible = false">取消</el-button>
        <el-button type="primary" @click="submitAddKey">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import { keyPoolApi } from '@/api/admin'
import { ElMessage } from 'element-plus'

// ── 号池列表 ──
const pools = ref([])
const loadingPools = ref(false)

async function fetchPools() {
  loadingPools.value = true
  try {
    const res = await keyPoolApi.listPools()
    pools.value = Array.isArray(res) ? res : (res.pools ?? [])
  } finally {
    loadingPools.value = false
  }
}

onMounted(fetchPools)

// ── 新建号池 ──
const createPoolVisible = ref(false)
const poolForm = reactive({ channel_id: 1, name: '' })

function openCreatePool() {
  poolForm.channel_id = 1
  poolForm.name = ''
  createPoolVisible.value = true
}

async function submitCreatePool() {
  if (!poolForm.name.trim()) return ElMessage.error('请填写号池名称')
  await keyPoolApi.createPool({ channel_id: poolForm.channel_id, name: poolForm.name })
  ElMessage.success('号池创建成功')
  createPoolVisible.value = false
  fetchPools()
}

async function togglePool(row) {
  await keyPoolApi.togglePool(row.id)
  ElMessage.success(row.is_active ? '已停用' : '已启用')
  fetchPools()
}

async function toggleVendorSubmittable(row) {
  await keyPoolApi.toggleVendorSubmittable(row.id)
  ElMessage.success(row.vendor_submittable ? '已关闭号商上传' : '已开放号商上传')
  fetchPools()
}

async function deletePool(id) {
  await keyPoolApi.deletePool(id)
  ElMessage.success('已删除')
  fetchPools()
}

// ── Keys 管理抽屉 ──
const keyDrawerVisible = ref(false)
const activePool = ref(null)
const keys = ref([])
const loadingKeys = ref(false)

async function openKeyDrawer(pool) {
  activePool.value = pool
  keyDrawerVisible.value = true
  await fetchKeys(pool.id)
}

async function fetchKeys(poolId) {
  loadingKeys.value = true
  try {
    const res = await keyPoolApi.listKeys(poolId)
    keys.value = Array.isArray(res) ? res : (res.keys ?? [])
  } finally {
    loadingKeys.value = false
  }
}

// ── 添加 Key ──
const addKeyVisible = ref(false)
const keyForm = reactive({ value: '', priority: 0 })

function openAddKey() {
  keyForm.value = ''
  keyForm.priority = 0
  addKeyVisible.value = true
}

async function submitAddKey() {
  if (!keyForm.value.trim()) return ElMessage.error('请填写 Key Value')
  await keyPoolApi.addKey(activePool.value.id, { value: keyForm.value.trim(), priority: keyForm.priority })
  ElMessage.success('Key 添加成功')
  addKeyVisible.value = false
  fetchKeys(activePool.value.id)
}

async function removeKey(keyId) {
  await keyPoolApi.removeKey(keyId)
  ElMessage.success('已移除')
  fetchKeys(activePool.value.id)
}

// ── 编辑 Key ──
const editKeyVisible = ref(false)
const editKeyForm = reactive({ id: 0, priority: 0, is_active: true })

function openEditKey(row) {
  editKeyForm.id = row.id
  editKeyForm.priority = row.priority
  editKeyForm.is_active = row.is_active
  editKeyVisible.value = true
}

async function submitEditKey() {
  await keyPoolApi.updateKey(editKeyForm.id, {
    priority: editKeyForm.priority,
    is_active: editKeyForm.is_active,
  })
  ElMessage.success('已更新')
  editKeyVisible.value = false
  fetchKeys(activePool.value.id)
}

// ── 工具函数 ──
function maskKey(val) {
  if (!val) return '••••••'
  return val.length > 6 ? val.slice(-6) : val
}
</script>

<style scoped>
.keypools-page { display: flex; flex-direction: column; gap: 16px; }

.hero-card :deep(.el-card__body) { padding: 20px 24px; }
.hero-row { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
.hero-row h3 { margin: 4px 0; font-size: 1.15rem; }
.hero-row p { color: #6b7a90; font-size: .85rem; margin: 0; max-width: 520px; }
.eyebrow { font-size: .72rem; font-weight: 600; text-transform: uppercase; letter-spacing: 1px; color: #1e66ff; margin-bottom: 4px; }

.drawer-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0 8px;
  border-bottom: 1px solid #e7edf5;
  margin-bottom: 4px;
}

.key-value { display: flex; align-items: center; }

.empty-tip { padding: 32px 0; display: flex; justify-content: center; }
</style>
