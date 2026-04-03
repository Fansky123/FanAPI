# FanAPI

多渠道 LLM & AI 生成服务聚合平台，统一接口代理多个第三方 AI API（OpenAI、Claude 等），内置计费、用户和频道管理系统。

## 功能特性

- **多渠道代理** — 通过 yaegi 动态脚本映射请求/响应格式，灵活接入各类上游 API
- **LLM 对话** — 支持流式（SSE）和非流式代理，双阶段计费（预扣 + 结算）
- **异步任务** — 图片、视频、音频生成任务，支持异步轮询状态查询
- **计费系统** — 多维度计费模型（按 token / 图片 / 视频 / 音频 / 自定义脚本），余额管理与交易记录
- **用户系统** — 邮件验证码注册、JWT 登录、API Key 管理
- **管理后台** — 渠道 CRUD、用户充值、交易查询

## 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.26 |
| Web 框架 | Gin |
| 数据库 | PostgreSQL + xorm |
| 缓存 | Redis |
| 消息队列 | NATS |
| 认证 | JWT + API Key |
| 动态脚本 | yaegi |
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
# 使用 Docker 启动依赖服务
docker compose -f Dockerfile.dev up -d

# 启动服务
go run cmd/server/main.go
```

### 3. 数据库初始化

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
| POST | `/v1/llm?channel_id=X` | LLM 对话代理（支持 SSE） |
| POST | `/v1/image` | 图片生成（异步） |
| POST | `/v1/video` | 视频生成（异步） |
| POST | `/v1/audio` | 音频生成（异步） |
| GET | `/v1/tasks` | 任务列表 |
| GET | `/v1/tasks/:id` | 任务状态查询 |

### 管理接口（JWT + admin 角色）

| 方法 | 路径 | 说明 |
|------|------|------|
| CRUD | `/admin/channels` | 频道管理 |
| GET | `/admin/users` | 用户列表 |
| POST | `/admin/users/:id/recharge` | 用户充值 |
| GET | `/admin/transactions` | 全部交易记录 |
| GET | `/admin/tasks` | 全部任务查询 |

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
│   ├── admin/        # 管理后台前端（Vue 3）
│   └── user/         # 用户前端（Vue 3）
└── scripts/          # 数据库初始化脚本
```

## 参与贡献

1. Fork 本仓库
2. 新建 `feat/xxx` 分支
3. 提交代码
4. 发起 Pull Request
