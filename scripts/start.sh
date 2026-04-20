#!/bin/bash
# fanapi 一键启动脚本
set -e
cd "$(dirname "$0")/.."
ROOT=$(pwd)

echo "=== fanapi 启动脚本 ==="

# ---------- 1. PostgreSQL ----------
if ! pg_lsclusters 2>/dev/null | grep -q "online"; then
    echo "[1/4] 启动 PostgreSQL..."
    pg_ctlcluster 17 main start 2>/dev/null || true
else
    echo "[1/4] PostgreSQL 已运行 ✓"
fi

# ---------- 2. Redis ----------
REDIS_ADDR=$(awk '/^redis:/{f=1;next} f && /addr:/{gsub(/.*addr:[[:space:]]*/,""); gsub(/"/, ""); print; exit} /^[^[:space:]]/{f=0}' config.yaml | tr -d '\r')
REDIS_ADDR=${REDIS_ADDR:-"localhost:6379"}
REDIS_HOST=${REDIS_ADDR%%:*}
REDIS_PORT=${REDIS_ADDR##*:}
if python3 -c "import socket; s=socket.socket(); s.settimeout(2); s.connect(('${REDIS_HOST}', ${REDIS_PORT})); s.close()" 2>/dev/null; then
    echo "[2/4] Redis 已就绪 ${REDIS_ADDR} ✓"
else
    echo "[2/4] 无法连接 Redis (${REDIS_ADDR})，请确认 Redis 容器已启动并可达" >&2
    exit 1
fi

# ---------- 3. NATS ----------
if ! cat /proc/net/tcp 2>/dev/null | awk '{print $2}' | grep -qi "107E"; then
    echo "[3/4] 启动 NATS..."
    nats-server -p 4222 &>/tmp/nats.log &
    sleep 1
else
    echo "[3/4] NATS 已运行 ✓"
fi

# ---------- 4. Stop old processes ----------
pkill -f "cmd/server" 2>/dev/null || true
pkill -f "cmd/script" 2>/dev/null || true
pkill -f "vite" 2>/dev/null || true
pkill -f "esbuild" 2>/dev/null || true
rm -f /tmp/server.pid /tmp/script.pid /tmp/user-web.pid
sleep 1

# ---------- 5. Start services ----------

echo ""
echo ">>> 启动 API Server (port 8080)..."
go run ./cmd/server &>/tmp/server.log &
echo $! > /tmp/server.pid

echo ">>> 启动 Script Worker..."
go run ./cmd/script &>/tmp/script.log &
echo $! > /tmp/script.pid

# ---------- 前端（需要 Node.js） ----------
if command -v pnpm &>/dev/null || command -v corepack &>/dev/null || command -v npm &>/dev/null; then
    echo ">>> 启动前端 (port 3000)..."
    cd "$ROOT/web/app"
    if ! command -v pnpm &>/dev/null && command -v corepack &>/dev/null; then
        corepack enable >/dev/null 2>&1 || true
    fi
    if command -v pnpm &>/dev/null; then
        [ ! -d node_modules ] && pnpm install --silent
        pnpm run dev &>/tmp/user-web.log &
    else
        [ ! -d node_modules ] && npm install --silent
        npm run dev -- --host 0.0.0.0 &>/tmp/user-web.log &
    fi
    echo $! > /tmp/user-web.pid

    cd "$ROOT"
else
    echo "    [跳过前端] 未找到 npm，请手动运行:"
    echo "      cd web/app && pnpm install && pnpm run dev"
    echo "      cd web/admin && npm install && npm run dev"
fi

sleep 8

# ---------- 6. Health check ----------
if curl -sf http://localhost:8080/health >/dev/null 2>&1; then
    echo ""
    echo "=== 全部启动成功 ==="
    echo "  API Server:    http://localhost:8080"
    echo "  API 文档:      http://localhost:8080/docs"
    if command -v npm &>/dev/null; then
    echo "  前端:          http://localhost:3000 (或 3001，见 /tmp/user-web.log)"
    echo "  管理端:        http://localhost:3000/admin"
    fi
    echo ""
    echo "  管理账号:      admin@fanapi.dev   / Admin@2026!"
  echo "  测试账号:      test@fanapi.dev    / Test@2026!"
    echo ""
    echo "  server 日志:   tail -f /tmp/server.log"
    echo "  worker 日志:   tail -f /tmp/script.log"
    if command -v npm &>/dev/null; then
    echo "  前端日志:      tail -f /tmp/user-web.log"
    fi
else
    echo "启动失败，查看日志: cat /tmp/server.log"
    exit 1
fi
