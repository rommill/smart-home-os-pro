#!/bin/bash

cleanup() {
    echo ""
    echo "🛑 Stopping all background processes..."
    jobs -p | xargs -r kill 2>/dev/null
    exit 0
}
trap cleanup SIGINT SIGTERM EXIT

echo "🚀 Starting Smart Home OS Pro environment..."

# Освобождаем порты 8080 и 5500
for PORT in 8080 5500; do
    PORT_PID=$(lsof -ti:$PORT)
    if [ -n "$PORT_PID" ]; then
        echo "🧹 Freeing port $PORT..."
        echo "$PORT_PID" | xargs kill -9 2>/dev/null
    fi
done

# 1. Go Backend
echo "📦 Starting Go Backend..."
(cd backend && go run main.go) &
sleep 2

# 2. BLE Bridge
echo "📡 Starting BLE Bridge..."
python ble_bridge.py &

# 3. Frontend via pnpm
if [ -d "frontend" ]; then
    echo "💻 Starting React Frontend via pnpm..."
    (cd frontend && pnpm dev) &
fi

echo "✅ All services started! Press Ctrl+C to stop all."

wait