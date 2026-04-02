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
if ! redis-cli ping 2>/dev/null | grep -q "PONG"; then
    echo "[2/4] 启动 Redis..."
    redis-server --daemonize yes --logfile /tmp/redis.log
    sleep 1
else
    echo "[2/4] Redis 已运行 ✓"
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
sleep 1

echo ""
echo ">>> 启动 API Server (port 8080)..."
/tmp/fanapi-server &>/tmp/server.log &
echo $! > /tmp/server.pid

echo ">>> 启动 Script Worker..."
/tmp/fanapi-script &>/tmp/script.log &
echo $! > /tmp/script.pid

sleep 2

# ---------- 6. Health check ----------
if curl -sf http://localhost:8080/auth/login -X POST \
    -H "Content-Type: application/json" -d '{}' >/dev/null 2>&1; then
    echo ""
    echo "=== 全部启动成功 ==="
    echo "  API Server:    http://localhost:8080"
    echo "  API 文档:      http://localhost:8080/docs"
    echo "  管理账号:      admin@test.com / Admin1234!"
    echo ""
    echo "  server 日志:   tail -f /tmp/server.log"
    echo "  worker 日志:   tail -f /tmp/script.log"
else
    echo "启动失败，查看日志: cat /tmp/server.log"
    exit 1
fi
