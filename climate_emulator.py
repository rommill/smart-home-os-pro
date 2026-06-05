import random
import time
from pymongo import MongoClient
from datetime import datetime, timezone

# MongoDB Docker configuration
MONGO_URI = "mongodb://localhost:27017/"
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

# English system configuration profiles for European/Estonian logging
ROOMS_CONFIG = {
    1: {"current_temp": 22.0, "name": "Living Room"},
    2: {"current_temp": 20.5, "name": "Bedroom"},
    3: {"current_temp": 24.5, "name": "Kitchen"},
    4: {"current_temp": 23.0, "name": "Bathroom"}
}

print(f"🚀 Intelligent Climate Balancing Engine initiated. Tracking 4 nodes...")

try:
    while True:
        for room_id, config in ROOMS_CONFIG.items():
            # Standard default fallback target value
            target_temp = 23.0 
            
            # Smart Lookup: Read the latest user slider interactions recorded by the Go Backend
            try:
                # Assuming target_temperature resides inside the same room record or latest telemetry mapping
                last_record = collection.find_one({"device_id": room_id}, sort=[("timestamp", -1)])
                if last_record and "target_temperature" in last_record:
                    target_temp = float(last_record["target_temperature"])
                elif last_record and "TargetTemperature" in last_record:
                    target_temp = float(last_record["TargetTemperature"])
            except Exception:
                pass # Fallback smoothly if DB schema layout shifts during transaction tasks

            current = config["current_temp"]

            # DEMO ENGINE: Smooth iterative transition logic mimicking hardware inertia transitions
            if current < target_temp - 0.2:
                # Heater effect: step up towards target parameters
                current += 0.2
            elif current > target_temp + 0.2:
                # AC effect: step down towards target parameters
                current -= 0.2
            else:
                # Idle micro-fluctuations around stable target profile
                current += random.uniform(-0.1, 0.1)

            # Round off floating precision scales safely
            current = round(current, 1)
            config["current_temp"] = current
            humidity = random.randint(45, 55)

            payload = {
                "device_id": room_id,
                "value": current,
                "humidity": humidity,
                "target_temperature": target_temp, # Keep tracking bounds synchronized
                "timestamp": datetime.now(timezone.utc)
            }

            collection.insert_one(payload)
            print(f"💾 {config['name']} (ID {room_id}) -> Actual: {current}°C | Target: {target_temp}°C | Hum: {humidity}%")

        print("--------------------------------------------------------------------------------")
        time.sleep(3)

except KeyboardInterrupt:
    print("\n🛑 Climate engine shut down.")
    client.close()