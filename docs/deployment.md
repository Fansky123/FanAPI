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
- [常用维护命令](#常用维护命令)

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
| `http://localhost:3000` | 用户端前端 |
| `http://localhost:3000/admin` | 管理后台 |
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

#### 3c. NATS（需开启 JetStream 持久化）

```bash
docker run -d \
  --name nats \
  --restart unless-stopped \
  -p 0.0.0.0:4222:4222 \
  -v /opt/nats-data:/data \
  nats:latest \
  -js -sd /data -a 0.0.0.0
```

#### 3d. 初始化数据库表

```bash
# 执行项目根目录下的 SQL 脚本（按文件名日期顺序）
psql -h 127.0.0.1 -U postgres -d fanapi -f scripts/migrate_*.sql
```

---

### 步骤 4：在服务器上准备项目文件

```bash
# 直接在服务器上克隆代码（推荐，后续 git pull 即可升级）
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

smtp:
  host: smtp.example.com
  port: 465
  user: no-reply@example.com
  password: SMTP密码
  from: "FanAPI <no-reply@example.com>"
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

浏览器访问 `http://服务器IP` 打开用户端，`http://服务器IP/admin` 打开管理后台。

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
| `worker.subjects` | Script Worker 订阅的 NATS 主题列表，默认 `["task.>"]`（全类型）；专用 Worker 示例：`["task.video.*"]` |

---

## 升级与重新部署

### Docker 升级

```bash
# 1. 在开发机重新构建镜像
docker compose build

# 2. 推送到镜像仓库（或用 docker save / scp 方式传输）
docker push registry.example.com/fanapi-api:latest
docker push registry.example.com/fanapi-script:latest

# 3. 在服务器上拉取并重启
cd /opt/fanapi
docker compose pull
docker compose up -d
```

仅升级其中一个：

```bash
docker compose build api    && docker compose up -d api
docker compose build script && docker compose up -d script
```

### 物理机升级

```bash
# 1. 在开发机重新构建（见第一步：构建产物）

# 2. 上传新产物到服务器
scp out/fanapi-server out/fanapi-script user@your-server:/opt/fanapi/
scp -r web/user/dist user@your-server:/opt/fanapi/web/

# 3. 在服务器上重启服务
sudo systemctl restart fanapi-server
sudo systemctl restart fanapi-script
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
