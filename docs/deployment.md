# FanAPI 部署指南

本文档说明如何在开发环境进行本地调试，以及如何在生产环境使用 Docker 打包和部署。

---

## 目录

- [开发环境](#开发环境)
- [生产环境 — 打包镜像](#生产环境--打包镜像)
- [生产环境 — 单机部署](#生产环境--单机部署（api--script-同机）)
- [生产环境 — 分机部署](#生产环境--分机部署（api-和-script-分开）)
- [配置说明](#配置说明)
- [升级与重新部署](#升级与重新部署)

---

## 开发环境

### 前提条件

| 依赖 | 版本 | 说明 |
|------|------|------|
| Go | ≥ 1.22 | 编译后端 |
| Node.js | ≥ 20 | 编译前端 |
| PostgreSQL | ≥ 14 | 数据库 |
| Redis | ≥ 7 | 缓存 |
| NATS | ≥ 2.10 | 消息队列 |

### 1. 配置文件

```bash
cp config.yaml config.local.yaml
```

编辑 `config.yaml`（或 `config.local.yaml`）填写本地服务地址：

```yaml
db:
  host: localhost
  port: 5432
  user: postgres
  password: yourpassword
  dbname: fanapi

redis:
  addr: localhost:6379

nats:
  url: nats://localhost:4222

smtp:
  host: smtp.example.com
  port: 465
  user: no-reply@example.com
  password: yoursmtppassword
```

### 2. 一键启动

脚本会自动启动 PostgreSQL / NATS，检测 Redis，编译 Go 二进制，并以热重载方式跑前端 Vite dev server。

```bash
bash scripts/start.sh
```

启动后访问：

| 地址 | 说明 |
|------|------|
| `http://localhost:3000` | 用户端前端（Vite dev server） |
| `http://localhost:3000/admin` | 管理后台 |
| `http://localhost:8080` | API Server |
| `http://localhost:8080/docs` | 接口文档（Swagger） |

### 3. 手动分步启动

如不使用一键脚本，可分别启动各进程：

```bash
# 编译
go build -o /tmp/fanapi-server ./cmd/server
go build -o /tmp/fanapi-script ./cmd/script

# API Server
/tmp/fanapi-server

# Script Worker（另一个终端）
/tmp/fanapi-script

# 前端（另一个终端）
cd web/user && npm install && npm run dev
```

---

## 生产环境 — 打包镜像

项目提供两个独立的 Docker build target，对应两应用角色：

| Target | 内容 | 对外端口 |
|--------|------|----------|
| `api` | nginx + 前端静态文件 + `fanapi-server` | `80` |
| `script` | 仅 `fanapi-script` worker | 无（仅对外连接 DB / Redis / NATS） |

### 构建 api 镜像

```bash
docker build --target api -t fanapi-api:latest .
```

### 构建 script 镜像

```bash
docker build --target script -t fanapi-script:latest .
```

### 同时构建两个（使用 docker compose）

```bash
docker compose build
```

> **构建说明**：`Dockerfile` 使用多阶段构建，Go 二进制使用静态链接（`CGO_ENABLED=0`），最终镜像基于 `debian:bookworm-slim`，两个 target 共享 Stage 1（node-builder）和 Stage 2（go-builder）的缓存层，重复构建只会重跑发生变化的层。

---

## 生产环境 — 单机部署（api + script 同机）

适用于初期流量较小，一台服务器承载所有服务。

### 前提条件

服务器上需已部署并可访问：
- PostgreSQL
- Redis
- NATS

### 1. 准备配置文件

在服务器上创建 `config.yaml`：

```yaml
server:
  port: 8080
  jwt_secret: "替换为强随机字符串"
  jwt_expire_hours: 24

db:
  host: 数据库地址
  port: 5432
  user: postgres
  password: 数据库密码
  dbname: fanapi
  sslmode: disable

redis:
  addr: Redis地址:6379
  password: ""

nats:
  url: nats://NATS地址:4222

smtp:
  host: smtp.example.com
  port: 465
  user: no-reply@example.com
  password: SMTP密码
  from: "FanAPI <no-reply@example.com>"
```

### 2. 使用 docker compose 启动

将项目中的 `docker-compose.yml` 和 `config.yaml` 放到同一目录，然后：

```bash
# 启动所有服务（api + script）
docker compose up -d

# 查看运行状态
docker compose ps

# 查看日志
docker compose logs -f api
docker compose logs -f script
```

### 3. 确认服务正常

```bash
# 检查 API 健康状态
curl http://localhost/health

# 应返回 {"status":"ok"} 或 200
```

---

## 生产环境 — 分机部署（api 和 script 分开）

适用于规模扩大后，将 Worker 独立部署到另一台服务器：

### 服务器 A — 部署 api（前端 + API Server）

```bash
# 只启动 api 容器，不启动 script
docker compose up -d api
```

或不依赖 compose，直接 run：

```bash
docker run -d \
  --name fanapi-api \
  --restart unless-stopped \
  -p 80:80 \
  -v /path/to/config.yaml:/app/config.yaml:ro \
  fanapi-api:latest
```

### 服务器 B — 部署 script（Worker）

将 `fanapi-script:latest` 镜像推送到镜像仓库（或直接 `docker save` / `scp` 传输），然后在 B 机上：

```bash
docker run -d \
  --name fanapi-script \
  --restart unless-stopped \
  -v /path/to/config.yaml:/app/config.yaml:ro \
  fanapi-script:latest
```

> **注意**：两台服务器的 `config.yaml` 中 DB / Redis / NATS 地址要一致，均指向共享的中间件实例。

### 水平扩容 script

如果任务队列积压，可以在多台服务器上同时跑 `fanapi-script`，它们会通过 NATS 竞争消费消息，天然实现负载均衡，无需其他配置：

```bash
# 在多台机器上分别执行
docker run -d \
  --name fanapi-script \
  --restart unless-stopped \
  -v /path/to/config.yaml:/app/config.yaml:ro \
  fanapi-script:latest
```

---

## 配置说明

所有配置均通过 `config.yaml` 提供，生产环境通过 Docker 卷挂载覆盖：

```
-v /host/path/config.yaml:/app/config.yaml:ro
```

| 字段 | 说明 |
|------|------|
| `server.jwt_secret` | JWT 签名密钥，**生产必须替换为强随机字符串** |
| `server.jwt_expire_hours` | JWT 有效期（小时），默认 24 |
| `db.*` | PostgreSQL 连接信息 |
| `redis.*` | Redis 连接信息 |
| `nats.url` | NATS 连接地址 |
| `smtp.*` | 邮件服务配置，用于发送验证码 / 找回密码邮件 |
| `worker.*` | Script Worker 并发数等参数 |

---

## 升级与重新部署

### 拉取新代码后重新构建

```bash
git pull

# 重新构建并重启
docker compose build
docker compose up -d
```

### 仅重建 api 或 script 其中一个

```bash
docker compose build api
docker compose up -d api

# 或
docker compose build script
docker compose up -d script
```

### 数据库迁移

新部署时 `fanapi-server` 启动时会自动执行 `xorm Sync2` 补充新字段，无需手动操作。

若是从旧版升级，需要额外执行迁移 SQL（如有）：

```bash
psql -U postgres -d fanapi -f scripts/migrate_xxx.sql
```

---

## 常用维护命令

```bash
# 查看所有容器状态
docker compose ps

# 实时查看 API 日志
docker compose logs -f api

# 实时查看 Worker 日志
docker compose logs -f script

# 重启 api（不重建镜像）
docker compose restart api

# 停止所有服务
docker compose down

# 停止并删除数据卷（危险）
docker compose down -v
```
