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

## 参与贡献

1. Fork 本仓库
2. 新建 `feat/xxx` 分支
3. 提交代码
4. 发起 Pull Request
