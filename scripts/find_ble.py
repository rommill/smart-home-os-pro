import asyncio
import logging
from bleak import BleakScanner

logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s")

def detection_callback(device, advertisement_data):
    # Log any device that matches ATC, Mi, or has Service Data
    name = device.name or "Unknown"
    if "ATC" in name or "LYWSD03MMC" in name or advertisement_data.service_data:
        logging.info("Found Device: Name=%s, Address=%s, Services=%s, Data=%s", 
                     name, device.address, list(advertisement_data.service_uuids), advertisement_data.service_data)

async def main():
    logging.info("Scanning for ALL local BLE devices for 10 seconds...")
    scanner = BleakScanner(detection_callback=detection_callback)
    await scanner.start()
    await asyncio.sleep(10.0)
    await scanner.stop()
    logging.info("Scan complete.")

if __name__ == "__main__":
    asyncio.run(main())