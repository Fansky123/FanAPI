#!/bin/bash
# fanapi 一键启动脚本
set -e
cd "$(dirname "$0")/.."

echo "=== fanapi 启动脚本 ==="

# ---------- 1. PostgreSQL ----------
if ! pg_lsclusters 2>/dev/null | grep -q "online"; then
    echo "[1/4] 启动 PostgreSQL..."
    pg_ctlcluster 17 main start 2>/dev/null || true
else
    echo "[1/4] PostgreSQL 已运行 ✓"
fi

# ---------- 2. Redis ----------
if curl -s --max-time 1 telnet://127.0.0.1:6379 >/dev/null 2>&1; then
    echo "[2/4] Redis 已运行 ✓"
elif command -v redis-server >/dev/null 2>&1; then
    echo "[2/4] 启动 Redis..."
    redis-server --daemonize yes --logfile /tmp/redis.log
    sleep 1
else
    echo "[2/4] Redis 未运行且 redis-server 不可用，请先在宿主机启动 Redis (端口 6379)" >&2
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

# ---------- 4. Build ----------
echo "[4/4] 编译..."
go build -o /tmp/fanapi-server ./cmd/server
go build -o /tmp/fanapi-script ./cmd/script
echo "      编译完成 ✓"

# ---------- 5. Start services ----------
pkill -f "fanapi-server" 2>/dev/null || true
pkill -f "fanapi-script" 2>/dev/null || true
pkill -f "vite" 2>/dev/null || true
sleep 1

echo ""
echo ">>> 启动 API Server (port 8080)..."
/tmp/fanapi-server &>/tmp/server.log &
echo $! > /tmp/server.pid

echo ">>> 启动 Script Worker..."
/tmp/fanapi-script &>/tmp/script.log &
echo $! > /tmp/script.pid

# ---------- 前端（需要 Node.js） ----------
if command -v npm &>/dev/null; then
    echo ">>> 启动用户端前端 (port 3000)..."
    cd "$(dirname "$0")/../web/user"
    [ ! -d node_modules ] && npm install --silent
    npm run dev -- --host 0.0.0.0 &>/tmp/user-web.log &
    echo $! > /tmp/user-web.pid

    echo ">>> 启动管理端前端 (port 3001)..."
    cd "$(dirname "$0")/../web/admin"
    [ ! -d node_modules ] && npm install --silent
    npm run dev -- --host 0.0.0.0 &>/tmp/admin-web.log &
    echo $! > /tmp/admin-web.pid

    cd "$(dirname "$0")/.."
else
    echo "    [跳过前端] 未找到 npm，请手动运行:"
    echo "      cd web/user  && npm install && npm run dev"
    echo "      cd web/admin && npm install && npm run dev"
fi

sleep 2

# ---------- 6. Health check ----------
if curl -sf http://localhost:8080/health >/dev/null 2>&1; then
    echo ""
    echo "=== 全部启动成功 ==="
    echo "  API Server:    http://localhost:8080"
    echo "  API 文档:      http://localhost:8080/docs"
    if command -v npm &>/dev/null; then
    echo "  用户端:        http://localhost:3000"
    echo "  管理端:        http://localhost:3001"
    fi
    echo ""
    echo "  管理账号:      admin@test.com / Admin1234!"
    echo ""
    echo "  server 日志:   tail -f /tmp/server.log"
    echo "  worker 日志:   tail -f /tmp/script.log"
    if command -v npm &>/dev/null; then
    echo "  用户端日志:    tail -f /tmp/user-web.log"
    echo "  管理端日志:    tail -f /tmp/admin-web.log"
    fi
else
    echo "启动失败，查看日志: cat /tmp/server.log"
    exit 1
fi
