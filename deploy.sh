#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIR="$SCRIPT_DIR/server"
CLIENT_DIR="$SCRIPT_DIR/client"
WEB_DIR="$SERVER_DIR/web"
PORT="${PORT:-8088}"
export PATH="$HOME/flutter/bin:$PATH"

echo "=== SpaceGame Deploy ==="

# 1. Stop existing server
echo "[1/5] Stopping existing server..."
SERVER_PID=$(lsof -ti:"$PORT" 2>/dev/null || true)
if [ -n "$SERVER_PID" ]; then
    kill "$SERVER_PID" 2>/dev/null || true
    sleep 2
    kill -9 "$SERVER_PID" 2>/dev/null || true
    echo "  Stopped (PID: $SERVER_PID)"
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
go build -o server ./cmd/server/
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

nohup "$SERVER_DIR/server" > "$SCRIPT_DIR/spacegame.log" 2>&1 &
SERVER_PID=$!
echo "  Started (PID: $SERVER_PID)"

echo ""
echo "=== Deploy complete ==="
echo "Server: http://localhost:$PORT/"
echo "Logs:   tail -f $SCRIPT_DIR/spacegame.log"
