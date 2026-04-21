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

ENV http_proxy= \
    https_proxy= \
    HTTP_PROXY= \
    HTTPS_PROXY= \
    all_proxy= \
    ALL_PROXY= \
    no_proxy=* \
    NO_PROXY=*

WORKDIR /web
RUN unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY all_proxy ALL_PROXY && \
    npm config delete proxy || true && \
    npm config delete https-proxy || true && \
    npm install -g pnpm@9.15.9

# 先复制 package 文件利用缓存层
COPY web/app/package.json ./
COPY web/app/pnpm-lock.yaml ./
RUN unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY all_proxy ALL_PROXY && \
    npm_config_proxy= npm_config_https_proxy= pnpm install --frozen-lockfile

COPY web/app/ ./
RUN unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY all_proxy ALL_PROXY && \
    npm_config_proxy= npm_config_https_proxy= pnpm build

# ─────────────────────────────────────────────────────────────
# Stage 2: 编译 Go 二进制（静态链接，无 CGO）
# ─────────────────────────────────────────────────────────────
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/golang:1.26.2-alpine AS go-builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=sum.golang.google.cn \
    http_proxy= \
    https_proxy= \
    HTTP_PROXY= \
    HTTPS_PROXY= \
    all_proxy= \
    ALL_PROXY= \
    no_proxy=* \
    NO_PROXY=*

WORKDIR /src

# 先下载依赖（利用 Docker 层缓存）
COPY go.mod go.sum ./
RUN unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY all_proxy ALL_PROXY && \
    HTTP_PROXY= HTTPS_PROXY= ALL_PROXY= GOPROXY=https://goproxy.cn,direct GOSUMDB=sum.golang.google.cn go mod download

COPY . .
RUN unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY all_proxy ALL_PROXY && \
    HTTP_PROXY= HTTPS_PROXY= ALL_PROXY= GOPROXY=https://goproxy.cn,direct GOSUMDB=sum.golang.google.cn \
    go build -ldflags="-s -w" -trimpath -o /out/fanapi-server ./cmd/server && \
    HTTP_PROXY= HTTPS_PROXY= ALL_PROXY= GOPROXY=https://goproxy.cn,direct GOSUMDB=sum.golang.google.cn \
    go build -ldflags="-s -w" -trimpath -o /out/fanapi-script ./cmd/script

# ─────────────────────────────────────────────────────────────
# Stage 3a: api — nginx + 前端 + fanapi-server
# ─────────────────────────────────────────────────────────────
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/debian:bookworm-slim AS api

ENV http_proxy= \
    https_proxy= \
    HTTP_PROXY= \
    HTTPS_PROXY= \
    all_proxy= \
    ALL_PROXY= \
    no_proxy=* \
    NO_PROXY=*

RUN printf 'Acquire::http::Proxy "false";\nAcquire::https::Proxy "false";\n' > /etc/apt/apt.conf.d/99no-proxy && \
    apt-get -o Acquire::http::Proxy=false -o Acquire::https::Proxy=false update && \
    apt-get -o Acquire::http::Proxy=false -o Acquire::https::Proxy=false install -y --no-install-recommends \
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

ENV http_proxy= \
    https_proxy= \
    HTTP_PROXY= \
    HTTPS_PROXY= \
    all_proxy= \
    ALL_PROXY= \
    no_proxy=* \
    NO_PROXY=*

RUN printf 'Acquire::http::Proxy "false";\nAcquire::https::Proxy "false";\n' > /etc/apt/apt.conf.d/99no-proxy && \
    apt-get -o Acquire::http::Proxy=false -o Acquire::https::Proxy=false update && \
    apt-get -o Acquire::http::Proxy=false -o Acquire::https::Proxy=false install -y --no-install-recommends \
        ca-certificates \
        tzdata && \
    rm -rf /var/lib/apt/lists/*

ENV TZ=Asia/Shanghai

COPY --from=go-builder /out/fanapi-script /app/fanapi-script
COPY config.yaml /app/config.yaml

WORKDIR /app
CMD ["/app/fanapi-script"]
