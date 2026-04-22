#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIR="$SCRIPT_DIR/server"
CLIENT_DIR="$SCRIPT_DIR/client"
WEB_DIR="$SERVER_DIR/web"
PORT="${PORT:-8088}"
export PATH="$HOME/flutter/bin:$PATH"

echo "=== SpaceGame Deploy ==="

# Check dependencies
echo "Checking dependencies..."
for cmd in go flutter lsof; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "Error: '$cmd' is not installed or not in PATH"
        exit 1
    fi
done
echo "  All dependencies found"

# 1. Stop existing server
echo "[1/5] Stopping existing server..."
SERVER_PIDS=$(lsof -ti :"$PORT" 2>/dev/null || true)
if [ -n "$SERVER_PIDS" ]; then
    for pid in $SERVER_PIDS; do
        kill "$pid" 2>/dev/null || true
    done
    sleep 2
    for pid in $SERVER_PIDS; do
        kill -9 "$pid" 2>/dev/null || true
    done
    echo "  Waiting for server to stop..."
    STOP_WAIT=10
    STOPPED=0
    while [ $STOP_WAIT -gt 0 ]; do
        if ! lsof -ti :"$PORT" &>/dev/null; then
            STOPPED=1
            break
        fi
        sleep 1
        STOP_WAIT=$((STOP_WAIT - 1))
    done
    if [ $STOPPED -eq 1 ]; then
        echo "  Stopped (PIDs: $SERVER_PIDS)"
    else
        echo "  ERROR: Server failed to stop"
        exit 1
    fi
else
    echo "  No running server found"
fi

# 2. Build Flutter web
echo "[2/5] Building Flutter web..."
cd "$CLIENT_DIR"
flutter build web --release
echo "  Done"

# 3. Copy frontend to server web directory
echo "[3/5] Copying frontend to $WEB_DIR..."
rm -rf "$WEB_DIR"
mkdir -p "$WEB_DIR"
cp -r "$CLIENT_DIR/build/web/." "$WEB_DIR"
echo "  Done"

# 4. Build Go server
echo "[4/5] Building Go server..."
cd "$SERVER_DIR"
go build -o SpaceGameServer ./cmd/server/
echo "  Done"

# 5. Start server in background
echo "[5/5] Starting server..."
export DB_HOST="${DB_HOST:-localhost}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-spacegame}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"
export PORT="$PORT"

nohup "$SERVER_DIR/SpaceGameServer" > "$SCRIPT_DIR/spacegame.log" 2>&1 &
SERVER_PID=$!
echo "  Started (PID: $SERVER_PID)"

# Wait for server to be ready
echo "  Waiting for server to be ready..."
MAX_WAIT=30
WAITED=0
SERVER_READY=0
while [ $WAITED -lt $MAX_WAIT ]; do
    if curl -s -o /dev/null -w "%{http_code}" "http://localhost:$PORT/" 2>/dev/null | grep -q "200"; then
        SERVER_READY=1
        break
    fi
    sleep 1
    WAITED=$((WAITED + 1))
done

if [ $SERVER_READY -eq 1 ]; then
    echo "  Server is ready"
else
    echo "  ERROR: Server failed to start within ${MAX_WAIT}s"
    echo "  Check logs: tail -f $SCRIPT_DIR/spacegame.log"
    kill "$SERVER_PID" 2>/dev/null || true
    exit 1
fi

echo ""
echo "=== Deploy complete ==="
echo "Server: http://localhost:$PORT/"
echo "Logs:   tail -f $SCRIPT_DIR/spacegame.log"
