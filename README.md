# FanAPI

多渠道 LLM & AI 生成服务聚合平台，统一接口代理多个第三方 AI API（OpenAI、Claude 等），内置计费、用户和频道管理系统。

## 功能特性

- **多渠道代理** — 通过 goja（JS 运行时）动态脚本映射请求/响应格式，灵活接入各类上游 API
- **多协议支持** — 同时支持 OpenAI、Claude、Gemini 三种协议格式（含 SSE 流式）
- **LLM 对话** — 支持流式（SSE）和非流式代理，双阶段计费（预扣 + 结算），用户中断时按实际输出字符兜底估算
- **请求追踪** — LLM 响应头返回 `X-Corr-Id`，可与计费流水 `corr_id` 字段精确对应，用户可查询哪笔对话扣了多少费
- **异步任务** — 图片、视频、音频生成任务，支持异步轮询状态查询，失败自动退款
- **计费系统** — 多维度计费模型（按 token / 图片 / 视频 / 音频 / 自定义脚本），余额管理与交易记录
- **自动退费** — 任务失败（HTTP 错误、第三方业务失败、NATS 发布失败）均自动退还已扣 credits 并写退费流水
- **卡密充值** — 管理员生成卡密，用户凭码充值
- **用户系统** — 用户名+密码注册（邮箱可选，用于找回密码）、JWT 登录、API Key 管理
- **管理后台** — 渠道 CRUD、号池管理、用户充值、交易查询、卡密管理，与用户端共享同一前端入口

## 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.26 |
| Web 框架 | Gin |
| 数据库 | PostgreSQL + xorm |
| 缓存 | Redis |
| 消息队列 | NATS |
| 认证 | JWT + API Key |
| 动态脚本 | goja (JavaScript) |
| 前端 | Vue 3 + Vite |

## 依赖服务

- PostgreSQL（默认端口 5433）
- Redis（默认端口 6379）
- NATS（默认端口 4222）
- SMTP 邮件服务

## 快速开始

### 1. 配置

复制并编辑配置文件：

```bash
cp config.yaml config.local.yaml
# 编辑数据库、Redis、NATS、SMTP 等连接信息
```

### 2. 启动（开发环境）

```bash
bash scripts/start.sh
```

启动后访问地址：
- 用户端：`http://localhost:3000`
- 管理后台：`http://localhost:3000/admin`
- API 文档：`http://localhost:8080/docs`

### 3. 默认账号

服务首次启动时，数据库会自动创建以下账号：

| 角色 | 用户名 | 邮箱 | 密码 | 说明 |
|------|--------|------|------|------|
| 管理员 | `admin` | `admin@fanapi.dev` | `Admin@2026!` | 拥有全部管理接口权限 |
| 测试用户 | `test` | `test@fanapi.dev` | `Test@2026!` | 普通用户权限，用于接口调试 |

> **生产环境请立即修改默认密码。**

### 4. 数据库种子数据（可选）

```bash
# ChatFire 渠道预置数据
psql -U <user> -d <db> -f scripts/seed_chatfire.sql
```

### 5. 数据库迁移（非首次部署）

若数据库由旧版升级，需执行迁移脚本补充新字段（新部署由 xorm `Sync2` 自动处理，无需手动执行）：

```bash
psql -U <user> -d <db> -f scripts/migrate_20260405_add_error_script_corr_id.sql
```

高并发场景建议同时执行索引优化脚本（使用 `CONCURRENTLY`，不锁表，可在线执行）：

```bash
psql -U <user> -d <db> -f scripts/migrate_20260405_add_indexes.sql
```

## 渠道脚本系统

每个渠道可配置最多 4 个 JS 脚本，均通过管理后台编辑：

| 字段 | 函数名 | 说明 |
|------|--------|------|
| `request_script` | `mapRequest(input)` | 将平台标准请求转换为第三方 API 格式 |
| `response_script` | `mapResponse(output)` | 将第三方同步响应映射为平台标准格式（同步任务）或提取 `upstream_task_id`（异步任务） |
| `query_script` | `mapResponse(output)` | 将异步轮询响应映射为平台标准格式（`status`: 2=成功, 3=失败, 其他=进行中） |
| `error_script` | `checkError(response)` | 自定义错误检测，返回非空字符串=错误消息（触发退费），返回 `null`/`false`=正常 |

### error_script 示例

**ChatFire / OpenAI 错误格式：**
```js
function checkError(resp) {
    if (resp.error) return resp.error.code + ': ' + resp.error.message;
    return null;
}
```

**自定义 code+message 格式：**
```js
function checkError(resp) {
    if (resp.code !== 0 && resp.code !== 200) return resp.message || 'error code: ' + resp.code;
    return null;
}
```

> 未填写 `error_script` 时，平台会使用内置通用检测（自动识别 `{"error":{...}}` 和字符串类型错误码格式）。

## 计费说明

1 CNY = 1,000,000 credits

### LLM 双阶段计费

| 阶段 | 时机 | 说明 |
|------|------|------|
| `hold`（预扣） | 请求发出前 | 按最大上下文 + 最大输出 token 保守估算，原子扣除避免超额 |
| `settle`（结算） | 响应完成后 | 用精确 usage 重新计算，退还多扣或补扣差额 |

- 用户中断流式响应时，按实时累计字符数估算，不全额退款
- 每次 LLM 请求响应头携带 `X-Corr-Id`，可在计费记录中通过 `corr_id` 字段追溯

### 异步任务计费

| 事件 | 流水类型 | 说明 |
|------|----------|------|
| 任务创建成功 | `charge` | 任务参数已知，一次性精确扣费 |
| 任务失败（任意原因）| `refund` | 自动退还全部已扣 credits，`metrics.reason` 记录失败原因 |

失败场景覆盖：NATS 发布失败、上游 HTTP 错误、`error_script` 检测到错误、`response_script/query_script` 输出 `status=3`、任务超时（>2小时）。

## API 文档

### 认证接口（无需鉴权）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/auth/send-code` | 发送邮件验证码 |
| POST | `/auth/register` | 注册账号 |
| POST | `/auth/login` | 登录，返回 JWT |

### 用户接口（Bearer JWT 或 API Key）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/user/profile` | 查询个人资料 |
| GET | `/user/balance` | 查询余额 |
| GET | `/user/transactions` | 交易记录 |
| GET | `/user/channels` | 可用频道列表（含 `routing_model` 字段） |
| GET/POST/DELETE | `/user/apikeys` | API Key 管理 |
| PUT | `/user/password` | 修改密码 |
| POST | `/user/bind-email` | 绑定邮筱 |
| POST | `/user/cards/redeem` | 兑换卡密（需 JWT） |

### AI 调用接口（API Key）

渠道路由通过请求体的 `model` 字段指定——将其设为渠道**名称**（即 `/user/channels` 返回的 `routing_model` 字段的值），服务端会自动解析并替换为真实的上游模型名。兼容旧客户端，也可以使用 `?channel_id=X` 查询参数指定渠道。

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/chat/completions` | LLM 对话（OpenAI 标准格式，支持 SSE） |
| POST | `/v1/messages` | LLM 对话（Claude 原生格式，支持 SSE） |
| POST | `/v1/gemini` | LLM 对话（Gemini 原生格式，支持 SSE） |
| POST | `/v1/image` | 图片生成（异步） |
| POST | `/v1/video` | 视频生成（异步） |
| POST | `/v1/audio` | 音频生成（异步） |
| GET | `/v1/tasks` | 任务列表 |
| GET | `/v1/tasks/:id` | 任务状态查询 |
| GET | `/v1/llm-logs` | LLM 请求日志 |

### 管理接口（JWT + admin 角色）

| 方法 | 路径 | 说明 |
|------|------|------|
| CRUD | `/admin/channels` | 频道管理 |
| CRUD | `/admin/key-pools` | 号池管理 |
| GET/POST/DELETE | `/admin/key-pools/:id/keys` | 号池 Key 管理 |
| GET | `/admin/users` | 用户列表 |
| POST | `/admin/users/:id/recharge` | 用户充值 |
| PUT | `/admin/users/:id/password` | 重置用户密码 |
| GET | `/admin/transactions` | 全部交易记录 |
| GET | `/admin/tasks` | 全部任务查询 |
| GET | `/admin/tasks/:id` | 任务详情 |
| GET | `/admin/stats` | 平台数据统计 |
| POST | `/admin/cards/generate` | 批量生成卡密 |
| GET | `/admin/cards` | 卡密列表 |
| DELETE | `/admin/cards/:id` | 删除卡密 |
| POST | `/user/cards/redeem` | 用户兑换卡密（需 JWT）|
| GET | `/admin/llm-logs` | LLM 请求日志 |
| GET | `/admin/llm-logs/:id` | LLM 请求日志详情 |

## 项目结构

```
fanapi/
├── cmd/
│   ├── server/       # HTTP 服务入口
│   └── script/       # 脚本执行入口
├── internal/
│   ├── billing/      # 计费引擎（提取器、定价器）
│   ├── cache/        # Redis 缓存
│   ├── config/       # 配置加载
│   ├── db/           # 数据库连接
│   ├── handler/      # HTTP 路由处理器
│   ├── middleware/   # 认证、鉴权中间件
│   ├── model/        # 数据模型
│   ├── mq/           # NATS 消息队列
│   ├── script/       # 异步任务 worker（仅依赖 NATS，无需 DB/Redis）
│   ├── service/      # 业务逻辑层
│   └── taskresult/   # 结果处理器、批量写入器、异步轮询器
├── pkg/
│   └── mailer/       # 邮件发送
├── web/
│   └── user/         # 前端（Vue 3 + Vite，用户端 + 管理后台）
│       ├── src/views/         # 页面组件
│       │   ├── admin/         # 管理后台页面（路由前缀 /admin）
│       │   ├── auth/          # 登录 / 注册
│       │   ├── billing/       # 充值与账单
│       │   ├── dashboard/     # 布局与渠道列表
│       │   ├── docs/          # API 文档
│       │   ├── keys/          # API Key 管理
│       │   ├── playground/    # 在线调试
│       │   └── tasks/         # 任务中心
│       └── src/api/           # API 封装
│           ├── index.js       # 用户端 API
│           ├── http.js        # 用户端 axios 实例
│           ├── admin.js       # 管理端 API
│           └── admin-http.js  # 管理端 axios 实例
└── scripts/          # 数据库初始化脚本
```

---

## 管理员操作手册

> 访问地址：`http://localhost:3000/admin`（或生产域名 `/admin`）
> 需使用拥有 admin 角色的账号登录。

---

### 一、渠道管理

路径：**管理后台 → Channels**

每个渠道代表一个第三方 API 接入点。字段说明如下：

#### 基础信息

| 字段 | 说明 |
|------|------|
| 模型名称（路由键） | 用户调用 API 时在请求体 `model` 字段填写的值，如 `gpt-4o`、`nano-1001`。必须唯一 |
| 标准模型名 | 用于前端分组展示，可以是同一个上游模型的多个渠道共享同一个值 |
| 接口类型 | `llm`（对话）/ `image`（图片）/ `video`（视频）/ `audio`（音频） |
| API 协议 | `openai`（默认）/ `claude`（Anthropic 原生）/ `gemini`（Google 原生）。无入参脚本时平台自动转换格式；有入参脚本时脚本优先 |
| 上游 URL | 第三方 API 完整地址，如 `https://api.openai.com/v1/chat/completions` |
| 请求头（JSON）| 固定请求头，通常用于写 API Key，如 `{"Authorization": "Bearer sk-xxx"}` |
| 超时（ms）| 请求提交超时，LLM 建议 60000，图片建议 180000，视频建议 300000 |

---

#### 计费类型与价格配置

> **单位换算：1 元 = 1,000,000 credits**
> 所有价格字段均为 credits 数值。

##### token 计费（LLM 对话）

| 字段 | 含义 |
|------|------|
| 售价 · 输入 | 用户每消耗 100 万输入 token 被扣多少 credits |
| 售价 · 输出 | 用户每消耗 100 万输出 token 被扣多少 credits |
| 进价 · 输入 / 输出 | 平台支付给上游的成本，仅用于利润统计，不影响用户扣费 |
| 输入从响应取 | 开启后输入 token 数从响应 `usage` 字段读取（更精确），适合上游不在请求中返回 token 计数的场景 |

示例（¥15/M 输入，¥60/M 输出）：
```
售价 · 输入 = 15000000
售价 · 输出 = 60000000
```

##### image 计费（图片生成）

有两种模式，**档位定价优先级高于基础价格**：

**模式一：按档位定价（推荐）**
在表格中按 `1k`/`2k`/`3k`/`4k` 档位填入售价和进价（credits/张）。如果档位不在表中，使用"兜底价格"。

**模式二：基础价格 + 分辨率倍率**
填写"售价 · 基础"，在"高级配置（JSON）"中配置 `resolution_tiers` 倍率表。

##### video / audio 计费（视频 / 音频）

| 字段 | 含义 |
|------|------|
| 售价 · 每秒 | 用户每生成 1 秒内容被扣多少 credits |
| 进价 · 每秒 | 平台成本，仅用于统计 |

##### count 计费（按次）

| 字段 | 含义 |
|------|------|
| 售价 · 每次 | 每次调用扣多少 credits |
| 进价 · 每次 | 平台成本 |

##### custom 计费（自定义脚本）

在"高级配置（JSON）"旁边的脚本框中填写 JS 脚本，函数签名：
```js
function calcBilling(request) {
    // request 为请求体 JSON
    // 返回值为整数 credits 数
    return 10000;
}
```

---

#### 高级配置（JSON）

这个文本框用于配置**无法用上方表单表达**的高级参数，保存时会自动和上方价格字段合并。

常用字段：

```json
{
  "metric_paths": {
    "input_tokens":  "response.usage.prompt_tokens",
    "output_tokens": "response.usage.completion_tokens",
    "size":          "request.size",
    "duration":      "request.duration"
  },
  "resolution_tiers": [
    { "max_pixels": 1048576, "multiplier": 1.0 },
    { "max_pixels": 4194304, "multiplier": 2.0 },
    { "max_pixels": 99999999, "multiplier": 4.0 }
  ],
  "input_from_response": true,
  "pricing_groups": {
    "vip": {
      "input_price_per_1m_tokens":  8000000,
      "output_price_per_1m_tokens": 32000000
    },
    "premium": {
      "price_per_second": 6000
    }
  }
}
```

| 字段 | 说明 |
|------|------|
| `metric_paths` | 告诉计费引擎从请求/响应的哪个 JSON 路径取字段值，格式 `"来源.字段"` |
| `resolution_tiers` | 图片分辨率分档倍率，按像素总数从小到大排列 |
| `input_from_response` | 同表单中"输入从响应取"开关，二选一 |
| `pricing_groups` | **分组定价**，见下一节 |

---

#### 分组定价（pricing_groups）

`pricing_groups` 支持对不同用户群体设置不同价格。

**原理**：`pricing_groups` 下的 key 是用户 group 名，value 是想覆盖的价格字段（浅合并到基础配置上）。用户 group 为空时使用顶层基础价格。

**各计费类型对应的 key：**

| 计费类型 | 可覆盖的字段 |
|----------|-------------|
| `token` | `input_price_per_1m_tokens`、`output_price_per_1m_tokens` |
| `image`（档位模式） | `size_prices`（需完整 map）、`default_size_price` |
| `image`（基础价格模式） | `base_price` |
| `video` / `audio` | `price_per_second` |
| `count` | `price_per_call` |

**示例（LLM token 渠道）：**
```json
{
  "metric_paths": {
    "input_tokens":  "response.usage.prompt_tokens",
    "output_tokens": "response.usage.completion_tokens"
  },
  "pricing_groups": {
    "vip": {
      "input_price_per_1m_tokens":  8000000,
      "output_price_per_1m_tokens": 32000000
    }
  }
}
```

**示例（图片渠道，size_prices 模式）：**
```json
{
  "pricing_groups": {
    "vip": {
      "size_prices": { "1k": 3000, "2k": 9000, "4k": 30000 }
    }
  }
}
```

> ⚠️ 注意：`size_prices` 是浅合并，分组里必须写**完整的 map**，不能只写想改的档位。

---

#### 脚本配置

每个渠道最多可配置 4 个 JS 脚本（均通过管理后台编辑）：

| 字段 | 函数签名 | 触发时机 | 说明 |
|------|----------|----------|------|
| 入参映射脚本 | `function MapRequest(input)` | 请求发出前 | 将平台标准格式转换为第三方 API 所需格式；`input` 为请求体 JSON，返回值作为实际发送的请求体 |
| 出参映射脚本 | `function MapResponse(input)` | 响应返回后 | 同步任务：返回 `{code:200, url:'...', status:2}`；异步任务：返回 `{upstream_task_id:'xxx'}` 触发轮询 |
| 轮询映射脚本 | `function MapResponse(input)` | 每次轮询响应后 | 将第三方轮询响应映射为平台标准格式，`status` 字段：`2`=成功，`3`=失败，其他=仍在处理中 |
| 错误检测脚本 | `function checkError(resp)` | 每次响应后 | 返回非空字符串=错误（触发退费），返回 `null`/`false`=正常。未填时使用内置通用检测 |

**错误检测脚本示例：**
```js
// OpenAI / ChatFire 格式
function checkError(resp) {
    if (resp.error) return resp.error.code + ': ' + resp.error.message;
    return null;
}

// 自定义 code 格式
function checkError(resp) {
    if (resp.code !== 0 && resp.code !== 200) return resp.message || 'error: ' + resp.code;
    return null;
}
```

---

#### 认证方式

| 类型 | 适用场景 | 说明 |
|------|----------|------|
| `bearer`（默认） | 大多数 OpenAI 兼容 API | Header 中 `Authorization: Bearer <key>` |
| `query_param` | Gemini 原生格式等 | Key 附加到 URL 查询参数，需填写"参数名"（如 `key`） |
| `basic` | HTTP Basic Auth | Key 格式为 `user:password`（或仅密码，user 为空） |
| `sigv4` | AWS Bedrock 等 | Key 格式为 `ACCESS_KEY_ID:SECRET_ACCESS_KEY`，需填写 Region 和 Service |

---

#### 异步轮询配置（视频 / 音频）

适用于接口只返回任务 ID、需要轮询查询结果的场景。

1. **出参映射脚本**返回 `{ upstream_task_id: "xxx" }` → 触发异步模式
2. **轮询 URL** 填写轮询地址，支持 `{id}` 占位符（会被 `upstream_task_id` 替换），如 `https://api.example.com/v1/tasks/{id}`
3. **轮询映射脚本**将第三方响应转换为标准格式（`status: 2/3/其他`）
4. 超时 2 小时后任务自动标记失败并退款

---

#### 负载均衡（多渠道分流）

同一个"模型名称"可以对应多个渠道，系统按以下规则选择：

1. **先按优先级**（Priority）降序排列，优先级高的先选
2. **同优先级内**按权重（Weight）加权随机分流
3. **近期错误率过高**的渠道自动跳过（错误率 > 50% 且请求数 ≥ 5 次时降级）
4. 请求失败时自动换下一个渠道重试

---

#### 号池绑定（多 Key 轮转）

适用于同一渠道需要使用多个 API Key 轮转的场景（如防止单 Key 限速）。

1. 先在**号池管理**中创建号池并添加 Key
2. 编辑渠道时在"绑定号池"下拉中选择
3. 绑定后，Headers 中的 `Authorization` 字段会被号池中分配的 Key 覆盖
4. 系统使用**粘性分配**（Sticky Assignment）：同一用户/任务 ID 固定分配同一个 Key，Key 被标记为耗尽时自动轮转到下一个

> 新建渠道时"绑定号池"不可选，需先保存渠道后再编辑绑定。

---

### 二、号池管理

路径：**管理后台 → Key Pools**

号池是多个第三方 API Key 的集合，供渠道轮转使用。

| 操作 | 说明 |
|------|------|
| 新增号池 | 填写名称，绑定到渠道（在渠道编辑页操作） |
| 添加 Key | 在号池详情中添加，填写 Key 值（明文，加密存储） |
| 删除 Key | 软删除，不影响历史记录 |

---

### 三、用户管理

路径：**管理后台 → Users**

| 操作 | 说明 |
|------|------|
| 查看用户列表 | 显示 ID、用户名、邮箱、余额、分组、注册时间 |
| 充值 | 点击"充值"，输入 credits 数量（1 元 = 1,000,000 credits） |
| 重置密码 | 点击"重置密码"，直接设置新密码（无需旧密码） |
| 设置分组 | 点击"设置分组"，输入分组名（需与渠道 `pricing_groups` 中的 key 一致） |

**分组功能说明：**
- 分组名区分大小写，必须与 `billing_config.pricing_groups` 中的 key 完全一致
- 用户 group 为空 = 使用默认价格（顶层 billing_config 字段）
- 修改分组立即生效，不影响已完成的历史扣费

---

### 四、卡密管理

路径：**管理后台 → Cards**

| 操作 | 说明 |
|------|------|
| 生成卡密 | 填写数量、每张 credits 数、备注，批量生成 |
| 查看列表 | 可按状态筛选（未使用 / 已使用） |
| 删除卡密 | 只能删除未使用的卡密 |

用户在**充值页面**输入卡号兑换，格式：`FANAPI-XXXXXXXXXXXXXXXX`（16 位大写 hex）。

---

### 五、账单管理

路径：**管理后台 → Billing**

查看全平台所有用户的交易流水，支持分页。流水类型说明：

| 类型 | 触发时机 | credits 方向 |
|------|----------|-------------|
| `hold` | LLM 预扣（请求发出前） | 扣除（Redis 原子扣，仅记录流水） |
| `settle` | LLM 结算（响应完成后） | 补扣或退还差额 |
| `charge` | 异步任务（图片/视频/音频）一次性扣费 | 扣除 |
| `refund` | 任务失败自动退款 | 退还 |
| `recharge` | 管理员充值 / 用户卡密兑换 | 增加 |

---

### 六、任务管理

路径：**管理后台 → Tasks**

查看所有异步任务（图片/视频/音频生成请求），支持按状态、用户筛选。

| 字段 | 说明 |
|------|------|
| 状态 | `pending`=等待处理，`processing`=处理中，`done`=完成，`failed`=失败 |
| upstream_task_id | 第三方返回的任务 ID |
| 结果 | 任务完成后的标准化响应（含 `url`、`status` 等） |

---

### 七、统计面板

路径：**管理后台 → Dashboard**

实时展示：
- 总用户数、活跃渠道数
- 今日总收入（credits）、今日总成本（credits）、今日利润
- 今日请求数（LLM + 异步任务）

---

## 参与贡献

1. Fork 本仓库
2. 新建 `feat/xxx` 分支
3. 提交代码
4. 发起 Pull Request
