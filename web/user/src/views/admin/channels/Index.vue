<template>
  <div class="channels-page">
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">Channels</div>
          <h3>渠道与定价管理</h3>
          <p>统一管理上游渠道、模型映射、售价/进价和异步轮询策略。</p>
        </div>
        <el-button type="primary" @click="openDialog()">
          <el-icon><Plus /></el-icon> 新增渠道
        </el-button>
      </div>
    </el-card>

    <el-card>
    <el-table :data="channels" stripe border>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="模型名称（路由键）" />
      <el-table-column prop="model" label="模型" />
      <el-table-column prop="type" label="类型" width="90">
        <template #default="{ row }">
          <el-tag size="small">{{ row.type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="协议" width="80" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.protocol && row.protocol !== 'openai'" size="small" type="success">{{ row.protocol }}</el-tag>
          <span v-else style="color:#ccc;font-size:12px">openai</span>
        </template>
      </el-table-column>
      <el-table-column prop="billing_type" label="计费类型" width="100" />
      <el-table-column label="售价" width="160">
        <template #default="{ row }">
          <span style="font-size:12px">{{ formatPrice(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="进价" width="160">
        <template #default="{ row }">
          <span style="font-size:12px;color:#999">{{ formatCost(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="号池" width="80" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.key_pool_id" size="small" type="warning">{{ row.key_pool_id }}</el-tag>
          <span v-else style="color:#ccc;font-size:12px">—</span>
        </template>
      </el-table-column>
      <el-table-column label="优先级/权重" width="100" align="center">
        <template #default="{ row }">
          <span style="font-size:12px">P{{ row.priority ?? 0 }} / W{{ row.weight ?? 1 }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="is_active" label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-switch v-model="row.is_active" @change="toggleActive(row)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" align="center">
        <template #default="{ row }">
          <el-button size="small" @click="openDialog(row)">编辑</el-button>
          <el-popconfirm title="确认删除？" @confirm="deleteRow(row.id)">
            <template #reference>
              <el-button size="small" type="danger">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="editRow ? '编辑渠道' : '新增渠道'" width="760px" top="5vh">
      <el-form :model="form" label-width="120px" style="max-height:70vh;overflow-y:auto">
        <el-form-item label="模型名称" required>
          <el-input v-model="form.name" placeholder="如：nano-1001（用户在 model 字段填写此值路由到本渠道）" />
        </el-form-item>
        <el-form-item label="标准模型名" required>
          <el-input v-model="form.model" placeholder="如：nano-banana-pro（用于前端分组）" />
        </el-form-item>
        <el-form-item label="模型图标 URL">
          <div style="display:flex;gap:8px;align-items:center;width:100%">
            <el-input v-model="form.icon_url" placeholder="https://example.com/icon.png（留空则显示首字母）" clearable style="flex:1" />
            <el-button size="small" @click="iconFileInput.click()">本地上传</el-button>
            <input ref="iconFileInput" type="file" accept="image/*" style="display:none" @change="onIconFile" />
            <img v-if="form.icon_url" :src="form.icon_url" style="width:32px;height:32px;border-radius:6px;border:1px solid #e4e7ed;object-fit:contain;flex-shrink:0" @error="(e)=>e.target.style.display='none'" />
          </div>
          <div style="font-size:11px;color:#aaa;margin-top:4px">支持公网 URL、base64 或本地上传，显示在模型列表卡片左上角</div>
        </el-form-item>
        <el-form-item label="模型描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="简要描述该模型的特点和适用场景" />
        </el-form-item>
        <el-form-item label="接口类型" required>
          <el-select v-model="form.type" style="width:100%" @change="val => { if (!editRow) { form.timeout_ms = defaultTimeoutByType[val] ?? 60000; form.query_timeout_ms = defaultQueryTimeoutByType[val] ?? 30000 } }">
            <el-option label="LLM 对话" value="llm" />
            <el-option label="图片生成" value="image" />
            <el-option label="视频生成" value="video" />
            <el-option label="音频生成" value="audio" />
          </el-select>
        </el-form-item>
        <el-form-item label="API 协议">
          <el-select v-model="form.protocol" style="width:100%">
            <el-option label="OpenAI 兼容格式（默认）" value="openai" />
            <el-option label="Claude 原生格式（Anthropic）" value="claude" />
            <el-option label="Gemini 原生格式（Google）" value="gemini" />
          </el-select>
          <div style="font-size:11px;color:#aaa;margin-top:4px">
            无入参脚本时自动将 OpenAI 格式请求转换为所选协议格式；有入参脚本时脚本优先
          </div>
        </el-form-item>
        <el-form-item label="上游 URL" required>
          <el-input v-model="form.base_url" placeholder="https://api.example.com/v1/..." />
        </el-form-item>
        <el-form-item label="请求方法">
          <el-select v-model="form.method" style="width:100px">
            <el-option label="POST" value="POST" />
            <el-option label="GET" value="GET" />
          </el-select>
        </el-form-item>
        <el-form-item label="请求头（JSON）">
          <el-input v-model="form.headersStr" type="textarea" :rows="3"
            placeholder='{"Authorization": "Bearer YOUR_KEY"}' style="font-family:monospace" />
        </el-form-item>
        <el-form-item label="超时（ms）">
          <el-input-number v-model="form.timeout_ms" :min="1000" :step="1000" />
        </el-form-item>
        <el-form-item label="计费类型" required>
          <el-select v-model="form.billing_type" style="width:100%">
            <el-option label="token 计费" value="token" />
            <el-option label="图片计费" value="image" />
            <el-option label="视频计费" value="video" />
            <el-option label="音频计费" value="audio" />
            <el-option label="按次计费" value="count" />
            <el-option label="自定义脚本" value="custom" />
          </el-select>
        </el-form-item>

        <!-- ===== Token 计费价格 ===== -->
        <template v-if="form.billing_type === 'token'">
          <el-divider content-position="left" style="margin:8px 0 12px">
            <span style="font-size:13px;color:#666">Token 价格（单位：元 / 1M tokens）</span>
          </el-divider>
          <el-form-item label="售价 · 输入">
            <el-input-number v-model="form.bp.input_price_per_1m_tokens" :min="0" :step="1" :precision="4" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">元 / 1M tokens，用户被扣费</span>
          </el-form-item>
          <el-form-item label="售价 · 输出">
            <el-input-number v-model="form.bp.output_price_per_1m_tokens" :min="0" :step="1" :precision="4" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">元 / 1M tokens，用户被扣费</span>
          </el-form-item>
          <el-form-item label="进价 · 输入">
            <el-input-number v-model="form.bp.input_cost_per_1m_tokens" :min="0" :step="1" :precision="4" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">元 / 1M tokens，支付给上游（成本）</span>
          </el-form-item>
          <el-form-item label="进价 · 输出">
            <el-input-number v-model="form.bp.output_cost_per_1m_tokens" :min="0" :step="1" :precision="4" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">元 / 1M tokens，支付给上游（成本）</span>
          </el-form-item>
          <el-form-item label="输入从响应取">
            <el-switch v-model="form.bp.input_from_response" />
            <span style="margin-left:8px;color:#999;font-size:12px">开启后输入 token 数从响应 usage 字段读取</span>
          </el-form-item>
          <el-divider content-position="left" style="margin:4px 0 12px">
            <span style="font-size:12px;color:#aaa">Prompt Cache 价格（Claude / OpenAI / Gemini，留空按默认倍率）</span>
          </el-divider>
          <el-form-item label="缓存写入售价">
            <el-input-number v-model="form.bp.cache_creation_price_per_1m_tokens" :min="0" :step="1" :precision="4" style="width:200px" :controls="true" placeholder="留空=1.25×输入价" />
            <el-button link style="margin-left:6px;font-size:12px;color:#c0c4cc" @click="form.bp.cache_creation_price_per_1m_tokens=null">清空</el-button>
            <span style="margin-left:4px;color:#aaa;font-size:11px">元 / 1M tokens，用户被扣费（留空 = 输入价 × 1.25，仅 Claude）</span>
          </el-form-item>
          <el-form-item label="缓存读取售价">
            <el-input-number v-model="form.bp.cache_read_price_per_1m_tokens" :min="0" :step="1" :precision="4" style="width:200px" :controls="true" placeholder="留空=0.5×输入价" />
            <el-button link style="margin-left:6px;font-size:12px;color:#c0c4cc" @click="form.bp.cache_read_price_per_1m_tokens=null">清空</el-button>
            <span style="margin-left:4px;color:#aaa;font-size:11px">元 / 1M tokens，用户被扣费（留空 = 输入价 × 0.5）</span>
          </el-form-item>
          <el-form-item label="缓存写入进价">
            <el-input-number v-model="form.bp.cache_creation_cost_per_1m_tokens" :min="0" :step="1" :precision="4" style="width:200px" :controls="true" />
            <el-button link style="margin-left:6px;font-size:12px;color:#c0c4cc" @click="form.bp.cache_creation_cost_per_1m_tokens=null">清空</el-button>
            <span style="margin-left:4px;color:#aaa;font-size:11px">元 / 1M tokens，支付给上游（留空 = 进价 × 1.25）</span>
          </el-form-item>
          <el-form-item label="缓存读取进价">
            <el-input-number v-model="form.bp.cache_read_cost_per_1m_tokens" :min="0" :step="1" :precision="4" style="width:200px" :controls="true" />
            <el-button link style="margin-left:6px;font-size:12px;color:#c0c4cc" @click="form.bp.cache_read_cost_per_1m_tokens=null">清空</el-button>
            <span style="margin-left:4px;color:#aaa;font-size:11px">元 / 1M tokens，支付给上游（留空 = 进价 × 0.5）</span>
          </el-form-item>
        </template>

        <!-- ===== 图片计费价格 ===== -->
        <template v-if="form.billing_type === 'image'">
          <el-divider content-position="left" style="margin:8px 0 12px">
            <span style="font-size:13px;color:#666">图片价格（元 / 张）</span>
          </el-divider>
          <!-- 按档位定价表格（size_prices / size_costs） -->
          <el-form-item label="按档位定价">
            <div style="font-size:12px;color:#999;margin-bottom:8px">填写后按 size 档位精确定价，覆盖下方基础价格；留空则使用基础价格 + 分辨率倍率模式。</div>
            <el-table :data="sizeTierRows" border size="small" style="width:480px">
              <el-table-column prop="label" label="档位" width="60" align="center" />
              <el-table-column label="售价（元）" align="center">
                <template #default="{ row }">
                  <el-input-number v-model="form.bp.size_prices[row.key]" :min="0" :step="0.01" :precision="4" size="small" style="width:140px" />
                </template>
              </el-table-column>
              <el-table-column label="进价（元）" align="center">
                <template #default="{ row }">
                  <el-input-number v-model="form.bp.size_costs[row.key]" :min="0" :step="0.01" :precision="4" size="small" style="width:140px" />
                </template>
              </el-table-column>
            </el-table>
            <div style="margin-top:8px;display:flex;gap:24px;align-items:center">
              <span style="font-size:12px;color:#666">兜底售价（元，size 不在表中时）：</span>
              <el-input-number v-model="form.bp.default_size_price" :min="0" :step="0.01" :precision="4" size="small" style="width:150px" />
              <span style="font-size:12px;color:#666">兜底进价（元）：</span>
              <el-input-number v-model="form.bp.default_size_cost" :min="0" :step="0.01" :precision="4" size="small" style="width:150px" />
            </div>
          </el-form-item>
          <el-divider content-position="left" style="margin:4px 0 12px">
            <span style="font-size:12px;color:#aaa">基础价格（像素分档模式，档位定价留空时生效）</span>
          </el-divider>
          <el-form-item label="售价 · 基础">
            <el-input-number v-model="form.bp.base_price" :min="0" :step="0.01" :precision="4" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">元 / 张，用户被扣费</span>
          </el-form-item>
          <el-form-item label="进价 · 基础">
            <el-input-number v-model="form.bp.base_cost" :min="0" :step="0.01" :precision="4" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">元 / 张，支付给上游（成本）</span>
          </el-form-item>
        </template>

        <!-- ===== 视频 / 音频计费价格 ===== -->
        <template v-if="form.billing_type === 'video' || form.billing_type === 'audio'">
          <el-divider content-position="left" style="margin:8px 0 12px">
            <span style="font-size:13px;color:#666">{{ form.billing_type === 'video' ? '视频' : '音频' }}价格（单位：元 / 秒）</span>
          </el-divider>
          <el-form-item label="售价 · 每秒">
            <el-input-number v-model="form.bp.price_per_second" :min="0" :step="0.01" :precision="4" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">元 / 秒，用户被扣费</span>
          </el-form-item>
          <el-form-item label="进价 · 每秒">
            <el-input-number v-model="form.bp.cost_per_second" :min="0" :step="0.01" :precision="4" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">元 / 秒，支付给上游（成本）</span>
          </el-form-item>
        </template>

        <!-- ===== 按次计费价格 ===== -->
        <template v-if="form.billing_type === 'count'">
          <el-divider content-position="left" style="margin:8px 0 12px">
            <span style="font-size:13px;color:#666">按次价格（单位：元 / 次）</span>
          </el-divider>
          <el-form-item label="售价 · 每次">
            <el-input-number v-model="form.bp.price_per_call" :min="0" :step="0.01" :precision="4" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">元 / 次，用户被扣费</span>
          </el-form-item>
          <el-form-item label="进价 · 每次">
            <el-input-number v-model="form.bp.cost_per_call" :min="0" :step="0.01" :precision="4" style="width:200px" />
            <span style="margin-left:8px;color:#999;font-size:12px">元 / 次，支付给上游（成本）</span>
          </el-form-item>
        </template>

        <el-form-item label="高级配置（JSON）">
          <el-input v-model="form.billingConfigStr" type="textarea" :rows="6"
            placeholder='{&#10;  "pricing_groups": {&#10;    "vip":     { "price_per_second": 6000 },&#10;    "premium": { "price_per_second": 4000 }&#10;  },&#10;  "metric_paths": { "size": "request.size" }&#10;}'
            style="font-family:monospace;font-size:12px" />
          <div style="font-size:11px;color:#aaa;margin-top:4px">
            上方价格字段优先级更高，保存时自动合并到此 JSON。<br>
            <b>分组定价</b>：在 <code>pricing_groups</code> 中按用户组覆盖任意价格字段，token 类型用
            <code>input/output_price_per_1m_tokens</code>，image/video/audio/count 类型分别用
            <code>size_prices</code> / <code>price_per_second</code> / <code>price_per_call</code>。
          </div>
        </el-form-item>

        <el-divider content-position="left" style="margin:8px 0 12px">
          <span style="font-size:13px;color:#666">号池绑定（多 Key 轮转）</span>
        </el-divider>
        <el-form-item label="绑定号池">
          <el-select v-model="form.key_pool_id" placeholder="不启用（使用 Headers 中的静态 Key）" clearable style="width:100%"
            :disabled="!editRow">
            <el-option :value="0" label="不启用" />
            <el-option v-for="p in channelPools" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
          <div style="font-size:11px;color:#aaa;margin-top:4px">
            <template v-if="!editRow">保存渠道后再编辑可选择号池</template>
            <template v-else>绑定后 Headers 中的 Authorization 将被号池中的 Sticky Key 覆盖</template>
          </div>
        </el-form-item>
        <el-form-item label="入参映射脚本">
          <el-input v-model="form.request_script" type="textarea" :rows="8" placeholder="// 将用户请求体转换为上游所需格式&#10;// input 为解析后的请求 JSON，返回值作为上游请求体&#10;function MapRequest(input) {&#10;    return input&#10;}" style="font-family:monospace;font-size:.82rem" />
        </el-form-item>
        <el-form-item label="出参映射脚本">
          <el-input v-model="form.response_script" type="textarea" :rows="8" placeholder="// 将上游响应映射为平台标准格式&#10;// 同步接口：返回 {code: 200, url: '...', status: 2}&#10;// 异步接口：返回 {upstream_task_id: 'xxx'} 即可触发轮询&#10;function MapResponse(input) {&#10;    return input&#10;}" style="font-family:monospace;font-size:.82rem" />
        </el-form-item>

        <el-divider content-position="left" style="margin:8px 0 12px">
          <span style="font-size:13px;color:#666">异步轮询配置（视频 / 音频等异步接口使用）</span>
        </el-divider>
        <el-form-item label="轮询 URL">
          <el-input v-model="form.query_url" placeholder="https://api.example.com/v1/tasks/{id}（{id} 为第三方任务 ID 占位符）" />
        </el-form-item>
        <el-form-item label="轮询方法">
          <el-select v-model="form.query_method" style="width:100px">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
          </el-select>
        </el-form-item>
        <el-form-item label="轮询超时（ms）">
          <el-input-number v-model="form.query_timeout_ms" :min="1000" :step="1000" />
          <div style="font-size:11px;color:#aaa;margin-top:4px">单次轮询 HTTP 请求的最大等待时间，视频/音频建议 60000ms</div>
        </el-form-item>
        <el-form-item label="轮询映射脚本">
          <el-input v-model="form.query_script" type="textarea" :rows="8" placeholder="// 将第三方轮询响应映射为标准格式&#10;// status: 2=成功 3=失败 其他=进行中&#10;function MapResponse(input) {&#10;    return input&#10;}" style="font-family:monospace;font-size:.82rem" />
        </el-form-item>
        <el-form-item label="错误检测脚本">
          <div style="font-size:12px;color:#999;margin-bottom:4px">
            checkError(response) — 返回非空字符串表示错误（平台将退费并标记失败），返回 null/false 表示正常。<br>
            未填时使用内置通用检测（自动识别 <code>{"error":{...}}</code> 等常见格式）。
          </div>
          <el-input v-model="form.error_script" type="textarea" :rows="8"
            placeholder="function checkError(resp) {&#10;    // 示例：ChatFire / OpenAI 格式&#10;    if (resp.error) return resp.error.code + ': ' + resp.error.message;&#10;    return null;&#10;}"
            style="font-family:monospace;font-size:.82rem" />
        </el-form-item>

        <!-- 认证扩展 -->
        <el-divider content-position="left" style="margin:16px 0 8px">认证方式</el-divider>
        <el-form-item label="认证类型">
          <el-select v-model="form.auth_type" style="width:180px">
            <el-option label="Bearer Token（默认）" value="bearer" />
            <el-option label="Query Param（如 Gemini 原生）" value="query_param" />
            <el-option label="HTTP Basic Auth" value="basic" />
            <el-option label="AWS SigV4" value="sigv4" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.auth_type === 'query_param'" label="参数名">
          <el-input v-model="form.auth_param_name" placeholder="key" style="width:200px" />
          <span style="margin-left:8px;color:#999;font-size:12px">API Key 将附加到 URL 查询参数，如 ?key=xxx</span>
        </el-form-item>
        <template v-if="form.auth_type === 'sigv4'">
          <el-form-item label="Region">
            <el-input v-model="form.auth_region" placeholder="us-east-1" style="width:200px" />
          </el-form-item>
          <el-form-item label="Service">
            <el-input v-model="form.auth_service" placeholder="execute-api" style="width:200px" />
          </el-form-item>
          <div style="margin:-4px 0 12px 120px;font-size:12px;color:#999">Key 格式：ACCESS_KEY_ID:SECRET_ACCESS_KEY</div>
        </template>

        <!-- 负载均衡 -->
        <el-divider content-position="left" style="margin:16px 0 8px">负载均衡</el-divider>
        <el-form-item label="优先级">
          <el-input-number v-model="form.priority" :min="0" :step="1" style="width:160px" />
          <span style="margin-left:8px;color:#999;font-size:12px">同模型多渠道时，优先级高的渠道优先被选中（默认 0）</span>
        </el-form-item>
        <el-form-item label="权重">
          <el-input-number v-model="form.weight" :min="1" :step="1" style="width:160px" />
          <span style="margin-left:8px;color:#999;font-size:12px">同优先级内按权重随机分流（默认 1）</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveChannel">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import { channelApi, keyPoolApi } from '@/api/admin'
import { ElMessage } from 'element-plus'

const channels = ref([])
const dialogVisible = ref(false)
const editRow = ref(null)
const iconFileInput = ref(null)

// 1 CNY = 1,000,000 credits（内部存储单位）
// 前端所有价格字段均以 CNY 显示，保存时自动乘以此系数转换为 credits
const CREDITS_PER_CNY = 1_000_000

// 价格字段分类（用于批量转换）
const TOKEN_PRICE_FIELDS = [
  'input_price_per_1m_tokens', 'output_price_per_1m_tokens',
  'input_cost_per_1m_tokens', 'output_cost_per_1m_tokens',
  'cache_creation_price_per_1m_tokens', 'cache_read_price_per_1m_tokens',
  'cache_creation_cost_per_1m_tokens', 'cache_read_cost_per_1m_tokens',
]
const SCALAR_PRICE_FIELDS = [
  'base_price', 'base_cost',
  'default_size_price', 'default_size_cost',
  'price_per_second', 'cost_per_second',
  'price_per_call', 'cost_per_call',
]

function creditsToCny(v) {
  if (v == null) return null
  return Math.round(v / CREDITS_PER_CNY * 10000) / 10000  // 保留4位小数
}

function cnyToCredits(v) {
  if (v == null) return null
  return Math.round(v * CREDITS_PER_CNY)
}

function onIconFile(e) {
  const file = e.target.files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = ev => { form.icon_url = ev.target.result }
  reader.readAsDataURL(file)
  e.target.value = ''
}

// 按渠道类型推荐的超时默认值
const defaultTimeoutByType      = { llm: 60000, image: 180000, video: 300000, audio: 180000 }
const defaultQueryTimeoutByType = { llm: 30000, image: 30000,  video: 60000,  audio: 60000  }

const emptyForm = () => ({
  name: '', model: '', type: 'llm', protocol: 'openai', base_url: '', method: 'POST',
  headersStr: '{}', timeout_ms: defaultTimeoutByType.llm,
  billing_type: 'token', billingConfigStr: '{}',
  request_script: '', response_script: '',
  query_url: '', query_method: 'GET', query_timeout_ms: defaultQueryTimeoutByType.llm, query_script: '', error_script: '',
  key_pool_id: 0,
  is_active: true,  // 展示字段
  icon_url: '', description: '',  // 负载均衡
  weight: 1, priority: 0,
  // 认证扫展
  auth_type: 'bearer', auth_param_name: '', auth_region: '', auth_service: '',
  bp: emptyBp(),
})

const SIZE_TIERS = ['1k', '2k', '3k', '4k']
const sizeTierRows = SIZE_TIERS.map(k => ({ key: k, label: k }))

function emptyBp() {
  return {
    // token 类型
    input_price_per_1m_tokens: 0, output_price_per_1m_tokens: 0,
    input_cost_per_1m_tokens: 0, output_cost_per_1m_tokens: 0,
    input_from_response: false,
    // token 缓存价格（null = 不设置，按默认倍率计费）
    cache_creation_price_per_1m_tokens: null,
    cache_read_price_per_1m_tokens: null,
    cache_creation_cost_per_1m_tokens: null,
    cache_read_cost_per_1m_tokens: null,
    // image - 基础价格（像素分档模式）
    base_price: 0, base_cost: 0,
    // image - 按模档直接定价（size_prices 模式，优先级更高）
    size_prices: { '1k': 0, '2k': 0, '3k': 0, '4k': 0 },
    size_costs:  { '1k': 0, '2k': 0, '3k': 0, '4k': 0 },
    default_size_price: 0,
    default_size_cost: 0,
    // video / audio 类型
    price_per_second: 0, cost_per_second: 0,
    // count 类型
    price_per_call: 0, cost_per_call: 0,
  }
}

// 从 billing_config JSON 中提取结构化价格字段（credits → CNY 转换）
function extractBp(cfg) {
  const bp = emptyBp()
  const keys = Object.keys(bp)
  for (const k of keys) {
    if (cfg[k] === undefined) continue
    if (k === 'size_prices' || k === 'size_costs') {
      // 对象类型：每个档位值单独转换
      const obj = typeof cfg[k] === 'object' && cfg[k] !== null ? cfg[k] : {}
      bp[k] = {}
      for (const tier of SIZE_TIERS) {
        bp[k][tier] = obj[tier] != null ? creditsToCny(obj[tier]) : 0
      }
    } else if (TOKEN_PRICE_FIELDS.includes(k) || SCALAR_PRICE_FIELDS.includes(k)) {
      // 数值价格字段：credits → CNY
      bp[k] = cfg[k] != null ? creditsToCny(cfg[k]) : (k.startsWith('cache_') ? null : 0)
    } else {
      bp[k] = cfg[k]
    }
  }
  return bp
}

// 将结构化价格字段合并回 billing_config（CNY → credits 转换，过滤零值避免污染 JSON）
function mergeBpToConfig(bp, baseConfigStr) {
  let cfg = {}
  try { cfg = JSON.parse(baseConfigStr || '{}') } catch { cfg = {} }
  const cacheKeys = ['cache_creation_price_per_1m_tokens', 'cache_read_price_per_1m_tokens',
                     'cache_creation_cost_per_1m_tokens', 'cache_read_cost_per_1m_tokens']
  for (const [k, v] of Object.entries(bp)) {
    if (cacheKeys.includes(k)) {
      // null 表示使用默认倍率，不写入 JSON
      if (v !== null && v !== undefined && v > 0) cfg[k] = cnyToCredits(v)
      else delete cfg[k]
      continue
    }
    if (k === 'size_prices' || k === 'size_costs') {
      const converted = {}
      let anyNonZero = false
      for (const tier of SIZE_TIERS) {
        const cv = cnyToCredits(v?.[tier] || 0)
        converted[tier] = cv
        if (cv > 0) anyNonZero = true
      }
      if (anyNonZero) cfg[k] = converted
      else delete cfg[k]
      continue
    }
    if (TOKEN_PRICE_FIELDS.includes(k) || SCALAR_PRICE_FIELDS.includes(k)) {
      const credits = cnyToCredits(v || 0)
      if (credits !== 0 || cfg[k] !== undefined) cfg[k] = credits
      continue
    }
    // 非价格字段（input_from_response 等布尔/其他字段）
    if (v !== false) cfg[k] = v
    else if (cfg[k] !== undefined) cfg[k] = v
  }
  return cfg
}
const form = reactive(emptyForm())
const channelPools = ref([]) // 供编辑弹窗中「号池绑定」下拉使用

onMounted(fetchChannels)

async function fetchChannels() {
  const res = await channelApi.list()
  channels.value = res.channels ?? []
}

function openDialog(row = null) {
  editRow.value = row
  if (row) {
    const cfg = row.billing_config ?? {}
    // 将 billing_config 中的价格字段提到 bp，剩余字段留在 billingConfigStr
    const bp = extractBp(cfg)
    const bpKeys = new Set(Object.keys(emptyBp()))
    const rest = Object.fromEntries(Object.entries(cfg).filter(([k]) => !bpKeys.has(k)))
    Object.assign(form, {
      ...row,
      headersStr: JSON.stringify(row.headers ?? {}, null, 2),
      billingConfigStr: JSON.stringify(rest, null, 2),
      bp,
    })
    // 加载该渠道可用的号池列表（channel_id 精确匹配）
    keyPoolApi.listPools(row.id).then(res => {
      channelPools.value = Array.isArray(res) ? res : (res.pools ?? [])
    })
  } else {
    Object.assign(form, emptyForm())
    channelPools.value = []
  }
  dialogVisible.value = true
}

async function saveChannel() {
  let headers, billingConfig
  try {
    headers = JSON.parse(form.headersStr || '{}')
    billingConfig = mergeBpToConfig(form.bp, form.billingConfigStr)
  } catch {
    return ElMessage.error('JSON 格式错误，请检查请求头或高级配置')
  }

  const payload = {
    name: form.name, model: form.model, type: form.type, protocol: form.protocol,
    base_url: form.base_url, method: form.method, headers,
    timeout_ms: form.timeout_ms, billing_type: form.billing_type,
    billing_config: billingConfig, request_script: form.request_script,
    response_script: form.response_script,
    query_url: form.query_url, query_method: form.query_method,
    query_timeout_ms: form.query_timeout_ms ?? 30000,
    query_script: form.query_script,
    error_script: form.error_script,
    key_pool_id: form.key_pool_id ?? 0,
    is_active: form.is_active,
    // 展示字段
    icon_url: form.icon_url || '',
    description: form.description || '',
    // 认证扪展
    auth_type: form.auth_type || 'bearer',
    auth_param_name: form.auth_param_name || '',
    auth_region: form.auth_region || '',
    auth_service: form.auth_service || '',
    // 负载均衡
    weight: form.weight ?? 1,
    priority: form.priority ?? 0,
  }

  if (editRow.value) {
    await channelApi.update(editRow.value.id, payload)
  } else {
    await channelApi.create(payload)
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetchChannels()
}

async function deleteRow(id) {
  await channelApi.delete(id)
  ElMessage.success('已删除')
  fetchChannels()
}

async function toggleActive(row) {
  await channelApi.patchActive(row.id, row.is_active)
}

function formatPrice(row) {
  const c = row.billing_config ?? {}
  const fmtCny = (v) => v != null ? `¥${(v / CREDITS_PER_CNY).toFixed(4).replace(/\.?0+$/, '')}` : '¥0'
  if (row.billing_type === 'token') {
    return `输入 ${fmtCny(c.input_price_per_1m_tokens)} / 输出 ${fmtCny(c.output_price_per_1m_tokens)}`
  }
  if (row.billing_type === 'image') {
    if (c.size_prices) {
      const parts = SIZE_TIERS.filter(k => c.size_prices[k]).map(k => `${k}:${fmtCny(c.size_prices[k])}`)
      return parts.length ? parts.join(' / ') : `基础 ${fmtCny(c.base_price)}`
    }
    return `基础 ${fmtCny(c.base_price)}`
  }
  if (row.billing_type === 'video' || row.billing_type === 'audio') return `${fmtCny(c.price_per_second)} /秒`
  if (row.billing_type === 'count') return `${fmtCny(c.price_per_call)} /次`
  return '—'
}

function formatCost(row) {
  const c = row.billing_config ?? {}
  const fmtCny = (v) => v != null ? `¥${(v / CREDITS_PER_CNY).toFixed(4).replace(/\.?0+$/, '')}` : '¥0'
  if (row.billing_type === 'token') {
    return `输入 ${fmtCny(c.input_cost_per_1m_tokens)} / 输出 ${fmtCny(c.output_cost_per_1m_tokens)}`
  }
  if (row.billing_type === 'image') {
    if (c.size_costs) {
      const parts = SIZE_TIERS.filter(k => c.size_costs[k]).map(k => `${k}:${fmtCny(c.size_costs[k])}`)
      return parts.length ? parts.join(' / ') : `基础 ${fmtCny(c.base_cost)}`
    }
    return `基础 ${fmtCny(c.base_cost)}`
  }
  if (row.billing_type === 'video' || row.billing_type === 'audio') return `${fmtCny(c.cost_per_second)} /秒`
  if (row.billing_type === 'count') return `${fmtCny(c.cost_per_call)} /次`
  return '—'
}
</script>

<style scoped>
.channels-page { max-width: 1320px; }
.hero-card { margin-bottom: 16px; }
.hero-row { display:flex;align-items:center;justify-content:space-between;gap:16px; }
.eyebrow { color:#1e66ff;font-size:.82rem;font-weight:700;text-transform:uppercase;letter-spacing:.08em; }
.hero-row h3 { margin:8px 0 10px;font-size:1.55rem; }
.hero-row p { margin:0;color:#617086; }
@media (max-width: 900px) { .hero-row { flex-direction:column;align-items:flex-start; } }
</style>
