# FanAPI 部署指南

本文档涵盖从**开发机打包**到**生产环境上线**的完整流程，提供两套部署方案：

| 方案 | 适用场景 |
|------|----------|
| **方案一：Docker 部署** | 推荐。环境隔离好，升级回滚方便 |
| **方案二：物理机部署** | 不想引入 Docker，或需要直接跑在宿主机上 |

---

## 目录

- [服务构成](#服务构成)
- [开发环境本地调试](#开发环境本地调试)
- [第一步：在开发机构建产物](#第一步在开发机构建产物)
- [方案一：Docker 部署](#方案一docker-部署)
- [方案二：物理机部署](#方案二物理机部署)
- [配置说明](#配置说明)
- [升级与重新部署](#升级与重新部署)
  - [场景一：只修改了渠道脚本](#场景一只修改了渠道脚本js-脚本字段)
  - [场景二：修改了后端 Go 代码](#场景二修改了后端-go-代码api-逻辑handleerservice-等)
  - [场景三：修改了前端代码](#场景三修改了前端代码vue-页面样式api-路径等)
- [常用维护命令](#常用维护命令)
- [日志说明与排查指南](#日志说明与排查指南)
- [任务全链路追踪（数据库查询）](#任务全链路追踪数据库查询)

---

## 服务构成

FanAPI 由两个独立进程组成，可以部署在同一台服务器，也可以分开：

| 进程 | 说明 | 必须运行 |
|------|------|----------|
| `fanapi-server` | API Server + 前端静态文件（含管理后台） | ✅ 始终需要 |
| `fanapi-script` | 异步任务 Worker（图片/视频/音频生成等） | 仅当使用异步任务功能时需要 |

> **提示**：如果只使用 LLM（文字对话）功能，只需部署 `fanapi-server`，不需要 `fanapi-script`。

两个进程均依赖以下中间件，需提前部署并可访问：

| 中间件 | 版本 | 备注 |
|--------|------|------|
| PostgreSQL | ≥ 14 | 主数据库 |
| Redis | ≥ 7 | 缓存 / 余额 |
| NATS | ≥ 2.10 | 消息队列（仅使用异步任务时需要） |

---

## 开发环境本地调试

### 前提条件

| 工具 | 版本 |
|------|------|
| Go | ≥ 1.26 |
| Node.js | ≥ 20 |
| PostgreSQL | ≥ 14 |
| Redis | ≥ 7 |
| NATS Server | ≥ 2.10 |

### 1. 配置文件

```bash
cp config.yaml config.local.yaml
```

编辑 `config.yaml` 填写本地服务地址（各字段含义见[配置说明](#配置说明)）：

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

```bash
bash scripts/start.sh
```

脚本会自动启动 PostgreSQL / NATS，检测 Redis，编译 Go 二进制，并以热重载方式运行前端 Vite dev server。

启动后访问：

| 地址 | 说明 |
|------|------|
| `http://localhost:3000` | 用户端 |
| `http://localhost:3000/admin` | 管理后台（管理员） |
| `http://localhost:3000/vendor/login` | 卡商端（号商登录） |
| `http://localhost:3000/agent/login` | 客服端（客服登录） |
| `http://localhost:8080` | API Server |
| `http://localhost:8080/docs` | 接口文档 |

### 3. 手动分步启动

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

## 第一步：在开发机构建产物

无论选择哪种部署方案，都从这一步开始。在**有代码的开发机**上执行：

### 1.1 编译前端

```bash
cd web/user
npm ci
npm run build
# 产物输出到 web/user/dist/
cd ../..
```

### 1.2 编译 Go 二进制（静态链接，无 CGO）

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -trimpath -o out/fanapi-server ./cmd/server

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -trimpath -o out/fanapi-script ./cmd/script
```

> 如果部署目标是 ARM 服务器（如 AWS Graviton），将 `GOARCH=amd64` 改为 `GOARCH=arm64`。

### 1.3 产物清单

构建完成后应有：

```
out/
  fanapi-server       # API Server 可执行文件
  fanapi-script       # Script Worker 可执行文件
web/user/dist/        # 前端静态资源目录
```

> **选择 Docker 部署可跳过此步骤**，Docker 镜像内部自动完成编译，直接从[方案一](#方案一docker-部署)开始。

---

## 方案一：Docker 部署

### 服务器环境要求

- Docker ≥ 24
- Docker Compose Plugin（`docker compose` 命令可用）

### 步骤 1：在开发机构建 Docker 镜像

```bash
# 构建 api 镜像（nginx + 前端 + fanapi-server）
docker build --target api -t fanapi-api:latest .

# 构建 script 镜像（仅 fanapi-script worker）
docker build --target script -t fanapi-script:latest .

# 或者一次构建两个
docker compose build
```

> `Dockerfile` 使用多阶段构建，Go 二进制使用静态链接（`CGO_ENABLED=0`），两个 target 共享编译缓存层，重复构建只会重跑发生变化的层。

### 步骤 2：将镜像传到服务器

**方式 A — 推送到镜像仓库（推荐）**

```bash
# 在开发机打 tag 并推送
docker tag fanapi-api:latest    registry.example.com/fanapi-api:latest
docker tag fanapi-script:latest registry.example.com/fanapi-script:latest
docker push registry.example.com/fanapi-api:latest
docker push registry.example.com/fanapi-script:latest

# 在服务器上拉取
docker pull registry.example.com/fanapi-api:latest
docker pull registry.example.com/fanapi-script:latest
```

**方式 B — 直接传文件（无镜像仓库时）**

```bash
# 在开发机打包
docker save fanapi-api:latest    | gzip > fanapi-api.tar.gz
docker save fanapi-script:latest | gzip > fanapi-script.tar.gz

# 上传到服务器
scp fanapi-api.tar.gz fanapi-script.tar.gz user@your-server:/opt/fanapi/

# 在服务器上导入
docker load < /opt/fanapi/fanapi-api.tar.gz
docker load < /opt/fanapi/fanapi-script.tar.gz
```

### 步骤 3：在服务器上部署中间件

> 如果 PostgreSQL / Redis / NATS 已经存在，跳到步骤 4。

以下命令使用 Docker 在宿主机本地启动中间件，端口绑定 `0.0.0.0` 以便 fanapi 容器可通过 `host-gateway` 访问。

**⚠️ 安全须知**：绑定 `0.0.0.0` 后端口对网卡可见，请在服务器防火墙 / 云安全组中禁止 5432、6379、4222 的公网入站访问，只允许内网或本机使用。

#### 3a. PostgreSQL

```bash
docker run -d \
  --name postgres \
  --restart unless-stopped \
  -e POSTGRES_PASSWORD=your-pg-password \
  -e POSTGRES_DB=fanapi \
  -p 0.0.0.0:5432:5432 \
  -v /opt/pgdata:/var/lib/postgresql/data \
  postgres:16
```

#### 3b. Redis

> Redis 7+ 默认启用 ACL，**必须设置密码**，否则连接报 `WRONGPASS`。

```bash
docker run -d \
  --name redis \
  --restart unless-stopped \
  -p 0.0.0.0:6379:6379 \
  redis:latest \
  --requirepass "your-redis-password"
```

#### 3c. NATS（JetStream 持久化 + 大消息支持）

`max_payload` 只能通过配置文件设置，先在宿主机创建配置文件，再挂载进容器：

```bash
# 创建配置文件（不需要提前克隆代码）
mkdir -p /opt/nats-data /opt/nats-conf
cat > /opt/nats-conf/nats.conf << 'EOF'
max_payload: 62914560
jetstream {
  store_dir: /data
}
EOF

# 启动 NATS 容器
docker run -d \
  --name nats \
  --restart unless-stopped \
  -p 0.0.0.0:4222:4222 \
  -v /opt/nats-data:/data \
  -v /opt/nats-conf/nats.conf:/etc/nats.conf:ro \
  nats:latest \
  -c /etc/nats.conf -a 0.0.0.0
```

> `max_payload: 62914560` 将单条消息上限设为 **60 MB**，防止图片生成渠道返回 base64 内联数据时超过 NATS 默认 1 MB 限制（报 "result too large" 错误）。`jetstream.store_dir` 开启 JetStream 持久化，容器重启后队列中未处理的任务不丢失。

#### 3d. 初始化数据库表

```bash
# 先克隆代码拿到 SQL 脚本（如已克隆可跳过）
git clone <你的仓库地址> /opt/fanapi
psql -h 127.0.0.1 -U postgres -d fanapi -f /opt/fanapi/scripts/migrate_*.sql
```

---

### 步骤 4：在服务器上准备项目文件

```bash
# 如果上一步已经 clone，跳过此命令
git clone <你的仓库地址> /opt/fanapi
cd /opt/fanapi
```

---

### 步骤 5：配置 config.yaml

> **重要**：中间件运行在宿主机 Docker 容器中，fanapi 容器需通过 `host-gateway`（Docker 内置别名，自动解析为宿主机 IP）访问它们，**不能用 `localhost`**。

编辑 `/opt/fanapi/config.yaml`：

```yaml
server:
  port: 8080
  jwt_secret: "替换为强随机字符串"  # openssl rand -hex 32
  jwt_expire_hours: 24

db:
  host: host-gateway        # ← 不要写 localhost
  port: 5432
  user: postgres
  password: your-pg-password
  dbname: fanapi
  sslmode: disable
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_idle_sec: 300

redis:
  addr: host-gateway:6379   # ← 不要写 localhost
  password: "your-redis-password"
  db: 0

nats:
  url: nats://host-gateway:4222  # ← 不要写 localhost
  # memory_storage: true  # 取消注释切换为内存存储（吞吐更高，重启丢未处理消息）
  replicas: 1              # JetStream 流副本数，生产集群建议 3

smtp:
  host: smtp.example.com
  port: 465
  user: no-reply@example.com
  password: SMTP密码
  from: "FanAPI <no-reply@example.com>"

worker:
  max_concurrent: 5000   # Script Worker 最大并发任务数，按服务器资源和上游限速调整
  # subjects:            # 订阅的任务类型，默认全部；专用 Worker 示例：["task.video.*"]
  #   - task.>
```

---

### 步骤 6：确认 docker-compose.yml 端口配置

`docker-compose.yml` 中 api 服务已配置为：

```yaml
ports:
  - "127.0.0.1:8088:80"  # 只绑定本机，宿主机 nginx 来接收外部流量
extra_hosts:
  - "host-gateway:host-gateway"  # 让容器可访问宿主机上的中间件
```

- 端口 `8088` 可按需修改，保证和宿主机 nginx 反代目标一致即可。
- 如果没有宿主机 nginx，改为 `"0.0.0.0:80:80"` 直接对外暴露。

### 步骤 7：启动服务

```bash
cd /opt/fanapi

# 启动所有服务（api + script）
docker compose up -d

# 查看启动状态
docker compose ps
```

### 步骤 8：验证服务正常

```bash
# 应返回 {"status":"ok"}
curl http://localhost:8088/health
```

---

### 步骤 9（可选）：宿主机 nginx 配置域名和 SSL

当宿主机已安装 nginx 负责域名管理和 SSL 时，fanapi 容器**不**直接对外暴露 80 端口（已绑定 `127.0.0.1:8088`），由宿主机 nginx 做 SSL 终止后反代。

创建 `/etc/nginx/sites-available/fanapi.conf`（CentOS 放到 `/etc/nginx/conf.d/fanapi.conf`）：

```nginx
# HTTP → HTTPS 跳转
server {
    listen 80;
    server_name your.domain.com;
    return 301 https://$host$request_uri;
}

# HTTPS 入口
server {
    listen 443 ssl;
    server_name your.domain.com;

    ssl_certificate     /etc/ssl/certs/your.domain.com.pem;
    ssl_certificate_key /etc/ssl/private/your.domain.com.key;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         HIGH:!aNULL:!MD5;

    # 反代到 fanapi 容器内的 nginx
    location / {
        proxy_pass         http://127.0.0.1:8088;
        proxy_http_version 1.1;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_read_timeout 180s;
    }
}
```

启用并重载：

```bash
# Debian / Ubuntu
sudo ln -sf /etc/nginx/sites-available/fanapi.conf /etc/nginx/sites-enabled/
sudo nginx -t && sudo nginx -s reload
```

流量链路：

```
外部 HTTPS:443 → 宿主机 nginx（SSL 终止）→ 127.0.0.1:8088 → 容器内 nginx:80 → fanapi-server:8080
```

---

### 分机部署（api 和 script 分开运行）

**服务器 A — 只运行 api：**

```bash
cd /opt/fanapi
docker compose up -d api
```

**服务器 B — 只运行 script Worker：**

两台服务器的 `config.yaml` 中 DB / Redis / NATS 地址必须一致，均指向共享中间件。

```bash
docker run -d \
  --name fanapi-script \
  --restart unless-stopped \
  -v /opt/fanapi/config.yaml:/app/config.yaml:ro \
  fanapi-script:latest
```

**水平扩容（多台机器同时跑 script）：**

多个 script 实例通过 NATS 竞争消费消息，天然负载均衡，在更多机器上执行同一条命令即可，无需任何额外配置。

---

## 方案二：物理机部署

不使用 Docker，直接将二进制和静态文件部署到宿主机，使用 **systemd** 管理进程生命周期。

### 服务器环境要求

- Linux（Debian / Ubuntu / CentOS 均可），systemd
- nginx ≥ 1.18
- 无需安装 Go / Node.js（产物已在开发机编译好）

### 步骤 1：在服务器上创建目录

```bash
sudo mkdir -p /opt/fanapi/web
sudo mkdir -p /var/log/fanapi
```

### 步骤 2：上传产物

在**开发机**上执行（先完成[第一步：构建产物](#第一步在开发机构建产物)）：

```bash
# 上传二进制
scp out/fanapi-server out/fanapi-script user@your-server:/opt/fanapi/

# 上传前端静态资源
scp -r web/user/dist user@your-server:/opt/fanapi/web/
```

在**服务器**上赋予执行权限：

```bash
sudo chmod +x /opt/fanapi/fanapi-server /opt/fanapi/fanapi-script
```

### 步骤 3：准备配置文件

在服务器上创建 `/opt/fanapi/config.yaml`：

```yaml
server:
  port: 8080
  jwt_secret: "替换为强随机字符串"  # openssl rand -hex 32
  jwt_expire_hours: 24

db:
  host: 数据库地址
  port: 5432
  user: postgres
  password: 数据库密码
  dbname: fanapi
  sslmode: disable
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_idle_sec: 300

redis:
  addr: Redis地址:6379
  password: ""
  db: 0

nats:
  url: nats://NATS地址:4222

smtp:
  host: smtp.example.com
  port: 465
  user: no-reply@example.com
  password: SMTP密码
  from: "FanAPI <no-reply@example.com>"
```

### 步骤 4：配置 nginx

安装 nginx（如未安装）：

```bash
# Debian / Ubuntu
sudo apt-get install -y nginx

# CentOS / RHEL
sudo yum install -y nginx
```

创建 `/etc/nginx/sites-available/fanapi`（CentOS 用户写到 `/etc/nginx/conf.d/fanapi.conf`）：

```nginx
server {
    listen 80;
    server_name _;

    root /opt/fanapi/web/dist;

    # ── API 反向代理 ──────────────────────────────────────
    location ~ ^/(auth|user|admin|v1|health|docs|pay)(/|$) {
        proxy_pass         http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header   Connection        "";
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;

        # LLM 流式响应超时适当放长
        proxy_connect_timeout  10s;
        proxy_read_timeout    180s;
        proxy_send_timeout    180s;

        # SSE / 流式响应禁用缓冲
        proxy_buffering             off;
        proxy_cache                 off;
        proxy_request_buffering     off;
    }

    # ── 前端 SPA ────────────────────────────────────────
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 静态资源长缓存（Vite 构建产物带 hash）
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff2?|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # index.html 不缓存
    location = /index.html {
        add_header Cache-Control "no-cache, no-store, must-revalidate";
    }
}
```

启用配置并重载：

```bash
# Debian / Ubuntu
sudo ln -sf /etc/nginx/sites-available/fanapi /etc/nginx/sites-enabled/fanapi
sudo rm -f /etc/nginx/sites-enabled/default   # 移除默认站点（可选）

# 检查配置语法
sudo nginx -t

# 启动并设置开机自启
sudo systemctl enable --now nginx
```

### 步骤 5：创建 systemd 服务

**API Server** — 创建 `/etc/systemd/system/fanapi-server.service`：

```ini
[Unit]
Description=FanAPI Server
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/fanapi
ExecStart=/opt/fanapi/fanapi-server
Restart=always
RestartSec=5
StandardOutput=append:/var/log/fanapi/server.log
StandardError=append:/var/log/fanapi/server.log

[Install]
WantedBy=multi-user.target
```

**Script Worker** — 创建 `/etc/systemd/system/fanapi-script.service`（不使用异步任务时可跳过）：

```ini
[Unit]
Description=FanAPI Script Worker
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/fanapi
ExecStart=/opt/fanapi/fanapi-script
Restart=always
RestartSec=5
StandardOutput=append:/var/log/fanapi/script.log
StandardError=append:/var/log/fanapi/script.log

[Install]
WantedBy=multi-user.target
```

### 步骤 6：启动服务

```bash
# 重新加载 systemd 配置
sudo systemctl daemon-reload

# 启动并设置开机自启
sudo systemctl enable --now fanapi-server
sudo systemctl enable --now fanapi-script   # 不使用异步任务时可跳过

# 查看状态
sudo systemctl status fanapi-server
sudo systemctl status fanapi-script
```

### 步骤 7：验证服务正常

```bash
# 应返回 {"status":"ok"}
curl http://localhost/health
```

浏览器访问各入口确认服务正常：

| 地址 | 说明 |
|------|------|
| `http://服务器IP` | 用户端 |
| `http://服务器IP/admin` | 管理后台（管理员） |
| `http://服务器IP/vendor/login` | 卡商端（号商登录） |
| `http://服务器IP/agent/login` | 客服端（客服登录） |

---

### 分机部署（物理机）

**服务器 A — 只运行 api：** 执行全部步骤，步骤 6 中跳过 `fanapi-script`。

**服务器 B — 只运行 script Worker：** 执行步骤 1、2（只上传 `fanapi-script`）、3、5（只创建 `fanapi-script.service`）、6（只启动 `fanapi-script`）。两台服务器的 `config.yaml` 中 DB / Redis / NATS 地址必须一致。

**水平扩容：** 在更多服务器上重复"服务器 B"的步骤，多个 script 实例通过 NATS 竞争消费，天然负载均衡。

---

## 配置说明

所有配置均通过 `config.yaml` 提供。Docker 部署通过卷挂载覆盖：

```
-v /host/path/config.yaml:/app/config.yaml:ro
```

| 字段 | 说明 |
|------|------|
| `server.jwt_secret` | JWT 签名密钥，**生产必须替换为强随机字符串**（`openssl rand -hex 32`） |
| `server.jwt_expire_hours` | JWT 有效期（小时），默认 24 |
| `db.host` / `db.port` / `db.user` / `db.password` / `db.dbname` | PostgreSQL 连接信息 |
| `db.sslmode` | PostgreSQL SSL 模式，内网可用 `disable` |
| `db.max_open_conns` | 最大打开连接数，建议与 pgBouncer pool_size 对齐，0 = 不限 |
| `db.max_idle_conns` | 最大空闲连接数，默认 2 |
| `db.conn_max_idle_sec` | 空闲连接超时秒数，防止被服务端踢掉，0 = 不限 |
| `redis.addr` | Redis 地址，格式 `host:port` |
| `redis.db` | Redis 数据库编号，默认 0 |
| `nats.url` | NATS 连接地址，格式 `nats://host:4222` |
| `nats.memory_storage` | `true` 切换为内存存储，吞吐更高但重启丢失队列中消息，默认 `false` |
| `nats.replicas` | JetStream 流副本数，单节点填 1，生产集群建议 3，默认 1 |
| `smtp.*` | 邮件服务配置，用于发送验证码 / 找回密码邮件 |
| `worker.max_concurrent` | Script Worker 最大同时执行的任务数，防止高并发打垮服务器或触发上游限速，默认 `100`，根据服务器资源和上游限速适当调大 |
| `worker.subjects` | Script Worker 订阅的 NATS 主题列表，默认 `["task.>"]`（全类型）；专用 Worker 示例：`["task.video.*"]` |

---

## 升级与重新部署

> **区分三种更新场景**：不同改动涉及的重建/重启范围不同，按需操作即可，无需每次全量重建。

| 改动类型 | 需要重编译 | 需要重建镜像 | 需要重启进程 |
|----------|-----------|-------------|-------------|
| 修改渠道脚本（JS 脚本字段） | ❌ | ❌ | 仅 script worker（`fanapi-script`）/ **或通过管理后台直接保存即生效** |
| 修改后端 Go 代码 | ✅ | Docker 需要 | ✅ `fanapi-server` 或 `fanapi-script` |
| 修改前端代码 | 前端 `npm build` | Docker 需要 | nginx reload 即可（无需重启 Go 进程） |

---

### 场景一：只修改了渠道脚本（JS 脚本字段）

渠道脚本存储在数据库中，`fanapi-script` 在每次处理任务时**实时从数据库读取**，通过管理后台保存后**立即生效，无需重启任何服务**。

若你直接修改了数据库或配置文件中的脚本，只需重启 script worker 使其刷新缓存：

**Docker：**

```bash
cd /opt/fanapi
docker compose restart script
```

**物理机：**

```bash
sudo systemctl restart fanapi-script
```

验证新脚本已被加载：

```bash
# Docker
docker compose logs -f script | grep -i "channel\|script"

# 物理机
sudo tail -f /var/log/fanapi/script.log
```

---

### 场景二：修改了后端 Go 代码（API 逻辑、handler、service 等）

**第一步：在开发机重新编译**

```bash
# 只编译 server（API 逻辑改动）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -trimpath -o out/fanapi-server ./cmd/server

# 只编译 script（Worker 逻辑改动）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -trimpath -o out/fanapi-script ./cmd/script
```

**第二步（Docker）：重建并重启**

```bash
# 只重建 api 镜像（server 改动）
docker compose build api
docker compose up -d api

# 只重建 script 镜像（worker 改动）
docker compose build script
docker compose up -d script
```

**第二步（物理机）：上传并重启**

```bash
# 上传新二进制（按实际改动选择）
scp out/fanapi-server user@your-server:/opt/fanapi/fanapi-server
scp out/fanapi-script user@your-server:/opt/fanapi/fanapi-script

# 在服务器上重启对应服务
sudo systemctl restart fanapi-server   # server 改动时
sudo systemctl restart fanapi-script   # script worker 改动时
```

**验证：**

```bash
# Docker
docker compose ps
curl http://localhost:8088/health   # 应返回 {"status":"ok"}

# 物理机
sudo systemctl status fanapi-server
curl http://localhost/health
```

---

### 场景三：修改了前端代码（Vue 页面、样式、API 路径等）

**第一步：在开发机重新构建前端**

```bash
cd web/user
npm ci          # 依赖有变化时执行，否则跳过
npm run build   # 产物输出到 web/user/dist/
cd ../..
```

**第二步（Docker）：重建 api 镜像并重启**

前端静态资源打包在 api 镜像里，需要重建镜像：

```bash
docker compose build api
docker compose up -d api
```

**第二步（物理机）：上传 dist 目录，reload nginx**

Go 进程无需重启，nginx 本身会直接读取文件系统，上传后 reload 即可：

```bash
# 上传新的前端静态资源
scp -r web/user/dist user@your-server:/opt/fanapi/web/dist

# reload nginx（不中断现有连接）
sudo nginx -t && sudo nginx -s reload
```

> 带 hash 的 JS/CSS 文件浏览器会自动更新；`index.html` 已配置 `no-cache`，用户刷新后立即加载新版本。

---

### 同时更新后端 + 前端

```bash
# 开发机：编译全部产物
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o out/fanapi-server ./cmd/server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o out/fanapi-script ./cmd/script
cd web/user && npm ci && npm run build && cd ../..

# Docker：全量重建
docker compose build
docker compose up -d

# 物理机：上传全部并重启
scp out/fanapi-server out/fanapi-script user@your-server:/opt/fanapi/
scp -r web/user/dist user@your-server:/opt/fanapi/web/dist
ssh user@your-server "sudo systemctl restart fanapi-server fanapi-script && sudo nginx -s reload"
```

### 数据库迁移

`fanapi-server` 启动时会自动执行 `xorm Sync2` 补充新字段，无需手动操作。

若升级说明中有额外迁移 SQL，手动执行：

```bash
psql -U postgres -d fanapi -f scripts/migrate_xxx.sql
```

---

## 常用维护命令

### Docker

```bash
# 查看运行状态
docker compose ps

# 实时查看日志
docker compose logs -f api
docker compose logs -f script

# 重启服务（不重建镜像）
docker compose restart api
docker compose restart script

# 停止所有服务
docker compose down
```

### 物理机

```bash
# 查看服务状态
sudo systemctl status fanapi-server
sudo systemctl status fanapi-script

# 实时查看日志
sudo tail -f /var/log/fanapi/server.log
sudo tail -f /var/log/fanapi/script.log

# 重启服务
sudo systemctl restart fanapi-server
sudo systemctl restart fanapi-script

# 停止服务
sudo systemctl stop fanapi-server
sudo systemctl stop fanapi-script
```

---

## 常见问题排查

### 502 Bad Gateway / fanapi-server 未启动

**症状**：`docker compose ps` 显示 `unhealthy`，日志出现 `connect() failed (111: Connection refused)`。

**排查步骤**：

```bash
# 1. 查看 fanapi-server 的实际报错
docker compose exec api /app/fanapi-server
# 或查看 supervisor 日志
docker compose logs api | grep -E "error|fatal|FATAL"
```

**常见原因及修复**：

| 报错关键字 | 原因 | 修复 |
|---|---|---|
| `WRONGPASS` | Redis 密码错误或为空（Redis 7+ 默认需要密码） | 在 `config.yaml` 填写 `redis.password`，或重建 Redis 容器时加 `--requirepass` |
| `connection refused` (5432/6379/4222) | 中间件只监听 `127.0.0.1`，容器无法通过 `host-gateway` 访问 | 用 `ss -tlnp` 检查端口，重建中间件容器改为 `-p 0.0.0.0:端口:端口` |
| `no such host: host-gateway` | `docker-compose.yml` 缺少 `extra_hosts` 配置 | 确保两个服务都有 `extra_hosts: ["host-gateway:host-gateway"]` |
| `database "fanapi" does not exist` | 数据库未创建 | `psql -U postgres -c "CREATE DATABASE fanapi;"` |

---

### 前端页面路由 404（如 /admin、/dashboard）

**症状**：浏览器访问 `http://域名/admin` 返回 404。

**原因**：nginx 将 `/admin` 代理到 Go 后端，但后端没有对应的页面路由，返回 404。`/admin` 是 Vue Router 的前端路由，应回落到 `index.html`。

**状态**：项目 `docker/nginx.conf` 已修复，直连后端规则只保留 `v1`、`health`、`pay`，`/admin` 等前端路由自动走 SPA fallback。重新构建镜像即可：

```bash
docker compose build api && docker compose up -d api
```

---

### 接口请求返回 HTML 页面

**症状**：调用 `/api/xxx` 接口，响应内容是 `<!DOCTYPE html>` 而不是 JSON。

**原因**：前端 `baseURL` 为 `/api`，但 nginx 未配置 `/api/` 前缀的 rewrite。

**状态**：项目中 `docker/nginx.conf` 已包含 `/api/` rewrite 规则，确保使用最新代码构建镜像即可：

```bash
docker compose build api
docker compose up -d api
```

---

### 中间件端口检查

```bash
# 检查 PG / Redis / NATS 是否绑定在 0.0.0.0（容器可访问）
ss -tlnp | grep -E '5432|6379|4222'

# 期望输出（0.0.0.0 或 * 开头表示可访问，127.0.0.1 开头表示容器无法访问）
# LISTEN  0.0.0.0:5432  ← 正确
# LISTEN  127.0.0.1:6379 ← 容器无法访问，需要重建
```

---

### 防火墙安全建议

中间件端口绑定 `0.0.0.0` 后，务必在防火墙层面限制访问：

```bash
# 使用 iptables：拒绝公网访问 Redis / NATS
iptables -A INPUT -p tcp --dport 6379 ! -s 127.0.0.1 -j DROP
iptables -A INPUT -p tcp --dport 4222 ! -s 127.0.0.1 -j DROP

# 或使用阿里云/腾讯云安全组：不开放 5432 / 6379 / 4222 入站规则（推荐）
```

---

## 任务全链路追踪（数据库查询）

当一个异步任务出现问题（卡住、失败、结果不对）时，可通过以下 SQL 追踪它从 **API 创建 → Worker 执行 → 结果写回** 的完整过程。所有数据存储在三张表中：

| 表 | 存储内容 |
|---|---|
| `tasks` | 任务主体：请求参数、上游请求/响应、执行结果、错误信息 |
| `billing_transactions` | 计费流水：预扣 / 结算 / 退款每步操作 |
| `channels` | 渠道配置：任务用的哪个上游渠道 |

---

### 第一步：查任务主体

```sql
SELECT
    id,
    type,                 -- image / video / audio / music
    status,               -- pending / processing / done / failed
    progress,
    channel_id,
    api_key_id,
    user_id,
    corr_id,              -- 关联计费流水的唯一 ID
    upstream_task_id,     -- 异步渠道：第三方任务 ID（如有，说明走了轮询）
    error_msg,            -- 失败时的错误描述
    credits_charged,      -- 实际扣除的 credits
    request,              -- 用户提交的原始参数（model、prompt 等）
    upstream_request,     -- Worker 实际发给上游的请求（含 _url、_headers）
    upstream_response,    -- 上游原始返回的 JSON
    result,               -- response_script 映射后的标准结果（url、items 等）
    created_at,           -- 任务创建时间
    updated_at            -- 最后更新时间（完成 / 失败时间）
FROM tasks
WHERE id = <任务ID>;
```

**关键字段说明：**

| 字段 | 排查用途 |
|---|---|
| `request` | 确认用户提交的参数是否正确 |
| `upstream_request._url` | Worker 实际请求的上游 URL（可直接用于 curl 复现） |
| `upstream_request._headers` | 请求头（`Authorization` 已脱敏为 `Bearer ****`） |
| `upstream_response` | 上游原始返回，失败时看具体报错内容 |
| `result` | 脚本映射后的最终结果，成功时应含 `url` 或 `items` |
| `error_msg` | 平台侧错误描述（脚本报错、上游超时等） |
| `upstream_task_id` | 非空说明走了异步轮询；若任务卡在 `processing` 检查此字段 |
| `corr_id` | 用于关联下方计费流水查询 |

> 在 psql 命令行用 `\x` 切换垂直展示模式，JSON 字段更易阅读。GUI 工具（DBeaver / pgAdmin / TablePlus）直接点开 JSON 列即可。

---

### 第二步：查计费流水

```sql
SELECT
    id,
    type,          -- hold=预扣  settle=结算  refund=退款  charge=直扣
    credits,       -- 操作的 credits 数（正=扣出，负=退回）
    cost,          -- 上游进价成本
    balance_after, -- 操作后用户余额快照
    metrics,       -- 计费详情（token 数 / 分辨率 / 时长等）
    created_at
FROM billing_transactions
WHERE corr_id = (SELECT corr_id FROM tasks WHERE id = <任务ID>)
ORDER BY created_at;
```

**正常流水模式：**

| 结果 | 流水条数 | 顺序 |
|---|---|---|
| 任务成功 | 2 条 | `hold`（提交时预扣）→ `settle`（完成后按实际用量结算，多退少补） |
| 任务失败 | 2 条 | `hold`（提交时预扣）→ `refund`（失败后全额退款） |

若只有 `hold` 没有后续流水，说明任务还在处理中，或 Worker 结果尚未被 `result-writer` 写回。

---

### 第三步：查关联渠道配置

```sql
SELECT
    id,
    name,
    type,
    base_url,
    billing_type,
    billing_config,
    request_script,
    response_script,
    query_url
FROM channels
WHERE id = (SELECT channel_id FROM tasks WHERE id = <任务ID>);
```

用于确认任务使用的是哪个渠道、脚本内容是否符合预期、`query_url` 是否配置正确（异步轮询任务）。

---

### 三步合并（快速一览）

将 `<任务ID>` 替换为实际值，依次执行：

```sql
-- 步骤 1：任务主体（psql 用户建议先执行 \x 开启垂直展示）
SELECT id, type, status, error_msg, upstream_task_id, corr_id,
       request, upstream_request, upstream_response, result,
       created_at, updated_at
FROM tasks WHERE id = <任务ID>;

-- 步骤 2：计费流水
SELECT type, credits, cost, balance_after, metrics, created_at
FROM billing_transactions
WHERE corr_id = (SELECT corr_id FROM tasks WHERE id = <任务ID>)
ORDER BY created_at;

-- 步骤 3：渠道配置
SELECT id, name, type, base_url, billing_type, billing_config,
       request_script, response_script, query_url
FROM channels
WHERE id = (SELECT channel_id FROM tasks WHERE id = <任务ID>);
```

---

### 同步查看 Worker 日志

结合容器日志可以看到 Worker 处理该任务的实时过程：

```bash
# 在日志中过滤指定任务 ID（将 12 替换为实际 ID）
docker compose logs api script --no-log-prefix 2>&1 | grep "task 12\b"
```

**成功任务的典型日志顺序：**

```
[script worker] subscribed to task.>              ← Worker 启动订阅
（无报错日志 = request_script / upstream 调用正常）
[result-proc] subscribed to result.>              ← API 侧结果处理器
[result-writer] batch done update (1 rows): <nil> ← 写入 DB 成功（nil = 无错误）
```

**失败任务的典型日志：**

```
[worker] task 12: request mapping error: ...      ← request_script 报错
[worker] task 12: response mapping error: ...     ← response_script 报错
[worker] task 12: upstream error: ...             ← 上游 HTTP 调用失败
[worker] task 12: result too large (...), ...     ← 上游响应体超过 55MB 限制
```

---

## 日志说明与排查指南

### 日志输出位置

FanAPI 使用 Go 标准 `log` 包，所有日志直接写到 **stdout / stderr**，不产生本地日志文件。查看方式如下：

**Docker 部署：**

```bash
# 实时追踪 API Server（含 nginx + fanapi-server + result-processor）
docker compose logs -f api

# 实时追踪 Script Worker
docker compose logs -f script

# 只看最近 200 行，不跟踪
docker compose logs --tail=200 api

# 关键字过滤（例：只看 error）
docker compose logs -f api 2>&1 | grep -i error
```

**物理机（systemd）部署：**

```bash
# 实时查看 API Server 日志（写到文件）
sudo tail -f /var/log/fanapi/server.log

# 实时查看 Script Worker 日志
sudo tail -f /var/log/fanapi/script.log

# 使用 journalctl（如未配置写文件）
sudo journalctl -u fanapi-server -f
sudo journalctl -u fanapi-script -f

# 只看最近 100 行
sudo journalctl -u fanapi-server -n 100 --no-pager
```

**NATS Server：**

```bash
docker logs nats -f --tail=100
```

---

### 日志前缀速查表

所有模块的日志都带有固定前缀，便于快速定位问题来源：

| 前缀 | 所在模块 | 说明 |
|---|---|---|
| `[worker]` | `internal/script/worker.go` | 任务消息解析失败、上游 HTTP 调用失败、NATS 发布结果失败 |
| `[script worker]` | `internal/script/worker.go` | Worker 启动时的订阅信息 |
| `[result-proc]` | `internal/taskresult/handler.go` | 结果消息解析失败 |
| `[result-writer]` | `internal/taskresult/writer.go` | 批量写入数据库失败 |
| `[register]` `[login]` `[apikey]` | `internal/service/auth.go` | 用户注册/登录/密钥相关 DB 错误 |
| `[ocpc/...]` | `internal/service/ocpc.go` | OCPC 广告回传失败 |
| _(无前缀)_ | `cmd/server/main.go` `cmd/script/main.go` | 启动阶段连接失败（DB / Redis / NATS）|

---

### 常见问题日志特征与修复

#### ① 任务卡在「排队中」不动

**现象**：提交任务后状态长期为 `pending`，`/v1/tasks/:id` 返回 `code=150`。

**排查：**

```bash
# 检查 script worker 是否正在运行并消费消息
docker compose logs script -f --tail=50

# 应能看到类似输出：
# [script worker] subscribed to task.> (consumer: workers-all)
```

如果没有上述输出，说明 script worker 未运行或启动失败，查看启动报错：

```bash
docker compose logs script 2>&1 | grep -i "fatal\|error\|panic"
```

---

#### ② 任务失败，想查上游返回了什么

Script Worker 会把**上游请求**（含目标 URL、请求头脱敏版本）和**上游响应**存入数据库的 `tasks` 表：

```sql
SELECT id, status, error_msg, upstream_task_id,
       upstream_request, upstream_response
FROM tasks
WHERE id = <任务ID>;
```

- `upstream_request._url` — 实际请求的上游 URL
- `upstream_request._headers` — 请求头（Authorization 已脱敏为 `Bearer ****`）
- `upstream_response` — 上游原始响应 JSON
- `error_msg` — 平台侧错误描述

可用 `upstream_request` 的内容直接复现 curl 请求：

```bash
curl -X POST "<upstream_request._url>" \
  -H "Authorization: Bearer <真实Key>" \
  -H "Content-Type: application/json" \
  -d '<upstream_request 去掉 _url/_headers 字段后的 JSON>'
```

---

#### ③ JS 脚本执行报错

每个渠道有四段可配置的 JS 脚本，报错时日志前缀和错误信息各不相同：

| 脚本字段 | 触发阶段 | 日志前缀 | 错误样式 |
|---|---|---|---|
| `request_script` | Worker 发起上游请求前 | `[worker]` | `request mapping error: ...` |
| `response_script` | Worker 收到上游响应后 | `[worker]` | `response mapping error: ...` |
| `error_script` | Worker 检测上游错误时 | `[worker]` | `error_script failed: ...` |
| `query_script` | Poller 每次轮询后 | `[poller]` | `query_script error: ...` |

---

**日志特征速查：**

```
# request_script 语法错误（脚本无法编译）
[worker] task 123: request mapping error: script compile error: Unexpected token at script:5:10

# request_script 运行时错误（函数内部抛出异常）
[worker] task 123: request mapping error: "mapRequest" execution error: TypeError: ...

# response_script 运行时错误
[worker] task 123: response mapping error: "mapResponse" execution error: ReferenceError: xxx is not defined at script:12:5

# response_script 返回类型错误（没有返回对象）
[worker] task 123: response mapping error: function "mapResponse" must return an object, got string

# error_script 报错（不影响任务继续，只打印日志）
[worker] task 123: error_script failed: "checkError" execution error: ...

# query_script 报错（轮询停止，upstream_response 已写入 DB）
[poller] task 123: query_script error: "mapResponse" execution error: TypeError: ...
```

---

**排查步骤：**

**第一步：从日志确认是哪段脚本出错，以及错误位置（行号:列号）**

```bash
docker compose logs script api --no-log-prefix 2>&1 | grep "task 123"
```

**第二步：查数据库，拿到脚本的输入数据**

```sql
-- 查任务，拿到上游原始返回（response_script / query_script 的输入）
SELECT
    upstream_request,   -- request_script 的输入
    upstream_response,  -- response_script / query_script 的输入
    error_msg           -- 脚本报错的完整错误信息
FROM tasks
WHERE id = 123;
```

> **重要**：`upstream_response` 在脚本执行前已写入数据库，即使脚本报错也能查到上游真实返回了什么。

**第三步：在本地复现错误**

把数据库中拿到的 `upstream_response` 内容粘贴到浏览器控制台，本地运行脚本：

```javascript
// 把 upstream_response 的内容粘贴为 output 的值
const output = { /* 粘贴 upstream_response JSON */ };

// 粘贴渠道中配置的脚本内容，然后调用对应函数
function mapResponse(output) {
    // ... 渠道中配置的脚本 ...
}

console.log(mapResponse(output));  // 查看输出或报错
```

**第四步：修复脚本并验证**

进入**管理后台 → 渠道管理 → 对应渠道**，修改对应脚本字段，保存后**立即生效，无需重启任何服务**。

再重新提交一次任务验证是否正常。

---

**常见脚本错误原因：**

| 错误信息 | 原因 | 修复方向 |
|---|---|---|
| `script compile error: Unexpected token` | JS 语法错误，如少了括号、逗号 | 检查脚本语法，在浏览器控制台粘贴测试 |
| `function "mapRequest" not found` | 脚本里没有定义 `mapRequest` 函数 | 确保函数名拼写正确 |
| `function "mapResponse" must return an object` | 脚本 `return` 了字符串/null 而不是对象 | 确保返回 `{}` 格式的对象 |
| `ReferenceError: xxx is not defined` | 脚本里引用了不存在的变量或字段 | 检查字段名是否与上游返回的 JSON 一致 |
| `TypeError: Cannot read properties of undefined` | 访问了 `undefined` 的子字段，如 `output.data[0].url` 但 `data` 为空 | 加防御判断：`output.data && output.data[0] && output.data[0].url` |

---

#### ④ NATS 消息体过大（上游返回了 base64 图片）

**日志特征：**

```
[worker] task 123: result too large (58623102 bytes), stripping upstream_response
[worker] task 123: result still too large (57001234 bytes), marking failed
```

**原因**：上游 API 在响应中内联了 base64 图片数据，体积超过 55MB NATS 软限制。

**修复**：修改渠道的 `response_script`，只提取图片 URL，不透传整个响应体：

```javascript
function mapResponse(output) {
    // ✅ 只返回 URL，不透传整个响应
    return { url: output.data[0].url, status: 2 };
}
```

---

#### ⑤ 启动失败（连接中间件报错）

**日志特征（无前缀，出现在进程启动时）：**

```
db: failed to connect to `host=host-gateway`: dial error: ...
redis: WRONGPASS invalid username-password pair
nats: nats: no servers available for connection
```

对照[502 排查表](#502-bad-gateway--fanapi-server-未启动)逐项检查。

---

#### ⑥ 持久化日志到文件（Docker）

默认 Docker 日志存在内存中，重启后丢失较旧的记录。如需持久化，在 `docker-compose.yml` 中添加：

```yaml
services:
  api:
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
        max-file: "10"
  script:
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
        max-file: "10"
```

日志文件位于宿主机 `/var/lib/docker/containers/<容器ID>/<容器ID>-json.log`，`max-file: "10"` 表示最多保留 10 个轮转文件（共 500MB）。
