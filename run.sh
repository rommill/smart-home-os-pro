#!/bin/bash

cleanup() {
    echo ""
    echo "🛑 Stopping all background processes..."
    jobs -p | xargs -r kill 2>/dev/null
    exit 0
}
trap cleanup SIGINT SIGTERM EXIT

echo "🚀 Starting Smart Home OS Pro environment..."

# 1. Ждем здоровья баз
echo "⏳ Waiting for PostgreSQL & MongoDB to be fully healthy..."
until [ "$(docker inspect -f '{{.State.Health.Status}}' smart_postgres 2>/dev/null)" == "healthy" ] && \
      [ "$(docker inspect -f '{{.State.Health.Status}}' smart_mongo 2>/dev/null)" == "healthy" ]; do
  sleep 1
done
echo "✅ Databases are healthy and ready!"

# 2. Пауза для стабилизации маппинга портов на macOS
echo "⏳ Waiting for Docker ports to stabilize on host..."
sleep 3

for PORT in 8080 5500; do
    PORT_PID=$(lsof -ti:$PORT)
    if [ -n "$PORT_PID" ]; then
        echo "🧹 Freeing port $PORT..."
        echo "$PORT_PID" | xargs kill -9 2>/dev/null
    fi
done

# 3. Go Backend (используем 127.0.0.1, так как go run запускается локально на Mac)
echo "📦 Starting Go Backend..."
(cd backend && JWT_SECRET=super_secret_smart_home_jwt_key_2026 POSTGRES_URL="postgresql://admin:smart_password@127.0.0.1:5432/smart_home_db?sslmode=disable" MONGO_URI="mongodb://admin:smart_password@127.0.0.1:27017" go run main.go) &

# 4. Ожидание реальной готовности Go-сервера (health check) перед стартом остальных сервисов
echo "⏳ Waiting for Go Backend to accept connections..."
until curl -s http://127.0.0.1:8080/healthz > /dev/null; do
  sleep 1
done
echo "✅ Go Backend is up and running!"

# 5. BLE Bridge
echo "📡 Starting BLE Bridge..."
python ble_bridge.py &

# 6. Frontend via pnpm
if [ -d "frontend" ]; then
    echo "💻 Starting React Frontend via pnpm..."
    (cd frontend && pnpm dev) &
fi

# 7. Climate Emulator
if [ -f "climate_emulator.py" ]; then
    echo "🌡️ Starting Climate Emulator..."
    ./venv/bin/python3 climate_emulator.py &
fi

echo "✅ All services started! Press Ctrl+C to stop all."

wait