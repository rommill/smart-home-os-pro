import random
import time
import requests
from datetime import datetime, timezone
from pymongo import MongoClient

# RU: Конфигурация подключений и эндпоинтов сервисного API
# EN: Connection configuration and service API endpoints
MONGO_URI = "mongodb://localhost:27017/"
BACKEND_URL = "http://localhost:8080"
API_URL = f"{BACKEND_URL}/api/v1/telemetry"
TARGETS_URL = f"{BACKEND_URL}/api/v1/service/targets"

# RU: Заголовки для авторизации системного сервиса по статическому API-ключу
# EN: Headers for system service authorization via static API key
SERVICE_HEADERS = {
    "X-API-Key": "secret_emulator_key_123",
    "Content-Type": "application/json"
}

DB_NAME = "smart_home"
COLLECTION_NAME = "telemetry"

try:
    client = MongoClient(MONGO_URI, serverSelectionTimeoutMS=2000)
    db = client[DB_NAME]
    collection = db[COLLECTION_NAME]
    client.server_info()
    print("✅ Successfully connected to MongoDB database engine!")
except Exception as e:
    print(f"❌ Database connection error: {e}")
    exit(1)

ROOMS_CONFIG = {
    "1": {"current_temp": 22.0, "name": "Living Room"},
    "2": {"current_temp": 20.5, "name": "Bedroom"},
    "3": {"current_temp": 24.5, "name": "Kitchen"},
    "4": {"current_temp": 23.0, "name": "Bathroom"}
}

print("🚀 Intelligent Climate Balancing Engine initiated. Tracking 4 nodes...")

try:
    while True:
        # RU: Получаем актуальные target_temperature с бэкенда через защищенный API-ключ
        # EN: Fetch current target_temperature from backend via protected API key
        targets_map = {}
        try:
            resp = requests.get(TARGETS_URL, headers=SERVICE_HEADERS, timeout=1)
            if resp.status_code == 200:
                data = resp.json()
                rooms_list = data if isinstance(data, list) else data.get("rooms", [])
                for room in rooms_list:
                    r_id = str(room.get("id"))
                    if "target_temperature" in room:
                        targets_map[r_id] = float(room["target_temperature"])
        except Exception as e:
            pass

        for device_id, config in ROOMS_CONFIG.items():
            target_temp = targets_map.get(str(device_id), 23.0)

            current = config["current_temp"]

            if current < target_temp - 0.2:
                current += 0.2
            elif current > target_temp + 0.2:
                current -= 0.2
            else:
                current += random.uniform(-0.1, 0.1)

            current = round(current, 1)
            config["current_temp"] = current
            humidity = random.randint(45, 55)

            payload = {
                "device_id": str(device_id),
                "rssi": -65,
                "timestamp": int(datetime.now(timezone.utc).timestamp()),
                "sensor_type": "TH_SENSOR",
                "temperature": float(current),
                "humidity": float(humidity),
                "battery": 98,
                "voltage_mv": 3000
            }

            try:
                resp = requests.post(API_URL, json=payload, timeout=1)
                print("DEBUG RESP:", resp.status_code, resp.text)
                if resp.status_code not in (200, 201, 204):
                    print(f"⚠️ API status {resp.status_code} for ID {device_id}: {resp.text}")
            except Exception as e:
                print(f"❌ API post error for ID {device_id}: {e}")

            mongo_doc = payload.copy()
            mongo_doc["target_temperature"] = target_temp
            collection.insert_one(mongo_doc)

            print(f"💾 {config['name']} (ID {device_id}) -> Temp: {current}°C | Target: {target_temp}°C | Hum: {humidity}%")

        print("--------------------------------------------------------------------------------")
        time.sleep(3)

except KeyboardInterrupt:
    print("\n🛑 Climate engine shut down.")
    client.close()