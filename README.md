# FanAPI

多渠道 LLM & AI 生成服务聚合平台，统一接口代理多个第三方 AI API（OpenAI、Claude 等），内置计费、用户和频道管理系统。

## 功能特性

- **多渠道代理** — 通过 goja（JS 运行时）动态脚本映射请求/响应格式，灵活接入各类上游 API
- **多协议支持** — 同时支持 OpenAI、Claude、Gemini 三种协议格式（含 SSE 流式）
- **LLM 对话** — 支持流式（SSE）和非流式代理，双阶段计费（预扣 + 结算）
- **异步任务** — 图片、视频、音频生成任务，支持异步轮询状态查询
- **计费系统** — 多维度计费模型（按 token / 图片 / 视频 / 音频 / 自定义脚本），余额管理与交易记录
- **卡密充值** — 管理员生成卡密，用户凭码充值
- **用户系统** — 邮件验证码注册、JWT 登录、API Key 管理
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

| 角色 | 邮箱 | 密码 | 说明 |
|------|------|------|------|
| 管理员 | `admin@fanapi.dev` | `Admin@2026!` | 拥有全部管理接口权限 |
| 测试用户 | `test@fanapi.dev` | `Test@2026!` | 普通用户权限，用于接口调试 |

> **生产环境请立即修改默认密码。**

### 4. 数据库种子数据（可选）

```bash
psql -U <user> -d <db> -f scripts/seed_chatfire.sql
```

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
| GET | `/user/balance` | 查询余额 |
| GET | `/user/transactions` | 交易记录 |
| GET | `/user/channels` | 可用频道列表 |
| GET/POST/DELETE | `/user/apikeys` | API Key 管理 |

### AI 调用接口（API Key）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/chat/completions?channel_id=X` | LLM 对话（OpenAI 标准格式，支持 SSE） |
| POST | `/v1/messages?channel_id=X` | LLM 对话（Claude 原生格式，支持 SSE） |
| POST | `/v1/gemini?channel_id=X` | LLM 对话（Gemini 原生格式，支持 SSE） |
| POST | `/v1/image` | 图片生成（异步） |
| POST | `/v1/video` | 视频生成（异步） |
| POST | `/v1/audio` | 音频生成（异步） |
| GET | `/v1/tasks` | 任务列表 |
| GET | `/v1/tasks/:id` | 任务状态查询 |

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
| GET | `/admin/stats` | 平台数据统计 |
| POST | `/admin/cards/generate` | 批量生成卡密 |
| GET | `/admin/cards` | 卡密列表 |
| DELETE | `/admin/cards/:id` | 删除卡密 |
| POST | `/user/cards/redeem` | 用户兑换卡密（需 JWT）|

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
│   ├── script/       # 异步任务 worker
│   └── service/      # 业务逻辑层
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
