# =============================================================
# FanAPI — 生产镜像（多阶段构建）
#
# 两个独立 target，可按需分开部署：
#
#   api    — nginx(80) + 前端静态文件 + fanapi-server
#            docker build --target api    -t fanapi-api .
#
#   script — 仅 fanapi-script worker（无 nginx / 无前端）
#            docker build --target script -t fanapi-script .
#
# 挂载说明：
#   -v /host/config.yaml:/app/config.yaml  覆盖默认配置
#
# =============================================================

# ─────────────────────────────────────────────────────────────
# Stage 1: 构建前端静态资源
# ─────────────────────────────────────────────────────────────
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/node:20.19.6-alpine3.23 AS node-builder

WORKDIR /web

# 先复制 package 文件利用缓存层
COPY web/app/package*.json ./
RUN npm ci --prefer-offline

COPY web/app/ ./
RUN npm run build

# ─────────────────────────────────────────────────────────────
# Stage 2: 编译 Go 二进制（静态链接，无 CGO）
# ─────────────────────────────────────────────────────────────
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/golang:1.26.2-alpine AS go-builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=sum.golang.google.cn

WORKDIR /src

# 先下载依赖（利用 Docker 层缓存）
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -trimpath -o /out/fanapi-server ./cmd/server && \
    go build -ldflags="-s -w" -trimpath -o /out/fanapi-script ./cmd/script

# ─────────────────────────────────────────────────────────────
# Stage 3a: api — nginx + 前端 + fanapi-server
# ─────────────────────────────────────────────────────────────
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/debian:bookworm-slim AS api

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        nginx \
        supervisor \
        curl \
        ca-certificates \
        tzdata && \
    rm -rf /var/lib/apt/lists/* && \
    mkdir -p /var/log/supervisor

ENV TZ=Asia/Shanghai

COPY --from=go-builder /out/fanapi-server /app/fanapi-server
COPY --from=node-builder /web/dist /app/web/dist
COPY config.yaml /app/config.yaml
COPY docker/nginx.conf /etc/nginx/nginx.conf
COPY docker/supervisord-api.conf /etc/supervisor/conf.d/fanapi.conf

WORKDIR /app
EXPOSE 80
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD curl -fsS http://localhost/health || exit 1
CMD ["/usr/bin/supervisord", "-n", "-c", "/etc/supervisor/supervisord.conf"]

# ─────────────────────────────────────────────────────────────
# Stage 3b: script — 仅 fanapi-script worker
# ─────────────────────────────────────────────────────────────
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/debian:bookworm-slim AS script

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        tzdata && \
    rm -rf /var/lib/apt/lists/*

ENV TZ=Asia/Shanghai

COPY --from=go-builder /out/fanapi-script /app/fanapi-script
COPY config.yaml /app/config.yaml

WORKDIR /app
CMD ["/app/fanapi-script"]
