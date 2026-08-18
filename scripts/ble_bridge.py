import asyncio
import logging
import os
import struct
import time
from typing import Dict, Any, Optional
import httpx
from bleak import BleakScanner

# ---------------------------------------------------------------------------
# CONFIGURATION
# ---------------------------------------------------------------------------
LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO").upper()
BTHOME_SERVICE_UUID = "0000fcd2-0000-1000-8000-00805f9b34fb"
SEND_INTERVAL_SEC = int(os.getenv("SEND_INTERVAL_SEC", "30"))
TEMP_DELTA_THRESHOLD = float(os.getenv("TEMP_DELTA_THRESHOLD", "0.2"))
GO_BACKEND_URL = os.getenv("GO_BACKEND_URL", "http://127.0.0.1:8080/api/v1/telemetry")

DEVICE_ROOM_MAP = {
    "60B6BFF5-3829-63A6-CB3F-6D14B38A0BDE": "1",
}
DEFAULT_ROOM_ID = "1"

logging.basicConfig(
    level=LOG_LEVEL,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[logging.StreamHandler()]
)
logger = logging.getLogger("BLEBridge")

device_cache: Dict[str, Dict[str, Any]] = {}
http_client: Optional[httpx.AsyncClient] = None


def parse_bthome_v2(payload: bytes) -> Optional[Dict[str, Any]]:
    """Парсер BTHome V2 TLV с отладкой"""
    if len(payload) < 2:
        return None

    # Выводим сырые байты в DEBUG/INFO, чтобы увидеть точную структуру пакета
    logger.debug(f"Raw BTHome payload (hex): {payload.hex()}")

    # В BTHome v2 payload[0] - это BTHome Device Info (флаги)
    # Данные начинается с payload[1:]
    data = payload[1:]
    metrics: Dict[str, Any] = {}
    idx = 0

    try:
        while idx < len(data):
            sensor_id = data[idx]
            idx += 1

            if sensor_id == 0x02 and idx + 2 <= len(data):  # Temperature (2 bytes, signed int16, scale 0.01)
                temp_raw = struct.unpack("<h", data[idx:idx+2])[0]
                metrics["temperature"] = round(temp_raw * 0.01, 2)
                idx += 2
            elif sensor_id == 0x03 and idx + 2 <= len(data):  # Humidity (2 bytes, uint16, scale 0.01)
                hum_raw = struct.unpack("<H", data[idx:idx+2])[0]
                metrics["humidity"] = round(hum_raw * 0.01, 2)
                idx += 2
            elif sensor_id == 0x01 and idx + 1 <= len(data):  # Battery % (1 byte)
                metrics["battery"] = data[idx]
                idx += 1
            elif sensor_id == 0x0C and idx + 2 <= len(data):  # Voltage (2 bytes, uint16)
                volt_raw = struct.unpack("<H", data[idx:idx+2])[0]
                metrics["voltage_mv"] = volt_raw
                idx += 2
            else:
                # Если встретился неизвестный sensor_id, мы не знаем его размерность,
                # поэтому прекращаем разбор текущего пакета
                break

        return metrics if metrics else None
    except Exception as err:
        logger.error(f"Failed to parse BTHome payload ({payload.hex()}): {err}")
        return None


def should_publish(device_id: str, new_metrics: Dict[str, Any]) -> bool:
    """Фильтр частых отправлений (Throttling)"""
    now = time.time()
    if device_id not in device_cache:
        device_cache[device_id] = {"last_sent": now, "metrics": new_metrics}
        return True

    cached = device_cache[device_id]
    time_passed = now - cached["last_sent"]

    if time_passed >= SEND_INTERVAL_SEC:
        cached["last_sent"] = now
        cached["metrics"] = new_metrics
        return True

    last_temp = cached["metrics"].get("temperature")
    curr_temp = new_metrics.get("temperature")
    if last_temp is not None and curr_temp is not None:
        if abs(curr_temp - last_temp) >= TEMP_DELTA_THRESHOLD:
            logger.info(f"⚡ Temp jump for {device_id}: {last_temp}°C -> {curr_temp}°C")
            cached["last_sent"] = now
            cached["metrics"] = new_metrics
            return True

    return False


async def publish_telemetry(device_id: str, rssi: int, metrics: Dict[str, Any]):
    global http_client
    if http_client is None:
        http_client = httpx.AsyncClient(timeout=3.0)

    payload = {
        "device_id": device_id,
        "rssi": rssi,
        "timestamp": int(time.time()),
        "sensor_type": "LYWSD03MMC",
        "temperature": metrics.get("temperature"),
        "humidity": metrics.get("humidity"),
        "battery": metrics.get("battery"),
        "voltage_mv": metrics.get("voltage_mv")
    }

    try:
        response = await http_client.post(GO_BACKEND_URL, json=payload)
        if response.status_code in (200, 201):
            logger.info(f"✅ Sent to Go API [{response.status_code}]: {device_id} -> T:{metrics.get('temperature')}°C | H:{metrics.get('humidity')}%")
        else:
            logger.warning(f"⚠️ Go API returned HTTP {response.status_code}: {response.text}")
    except Exception as err:
        logger.error(f"❌ Failed to connect to Go Backend ({GO_BACKEND_URL}): {err}")


def handle_device(device, advertisement_data):
    service_data = advertisement_data.service_data
    if BTHOME_SERVICE_UUID in service_data:
        payload = service_data[BTHOME_SERVICE_UUID]
        
        # Логируем каждый входящий BTHome пакет для отладки
        logger.info(f"🔍 BTHome packet from {device.address}: {payload.hex()}")
        
        metrics = parse_bthome_v2(payload)

        # Проверяем, что парсер реально извлек температуру
        if not metrics or metrics.get("temperature") is None:
            logger.warning(f"⚠️ Could not extract temperature from payload: {payload.hex()}")
            return

        room_id = DEVICE_ROOM_MAP.get(device.address.upper(), DEFAULT_ROOM_ID)

        if should_publish(room_id, metrics):
            asyncio.create_task(publish_telemetry(room_id, advertisement_data.rssi, metrics))


async def main():
    global http_client
    logger.info("🚀 Starting BLE Bridge Service (Production Mode)...")
    logger.info(f"⚙️ Config: Target URL={GO_BACKEND_URL}, Interval={SEND_INTERVAL_SEC}s, Delta={TEMP_DELTA_THRESHOLD}°C")

    scanner = BleakScanner(detection_callback=handle_device)
    await scanner.start()

    try:
        while True:
            await asyncio.sleep(1)
    finally:
        logger.info("🛑 Stopping BLE scanner and HTTP client...")
        await scanner.stop()
        if http_client:
            await http_client.aclose()


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except (KeyboardInterrupt, SystemExit):
        logger.info("👋 BLE Bridge Service stopped cleanly.")