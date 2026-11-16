package models

import "time"

// [EN] RoomData represents the unified data structure for a smart home room.
// [RU] RoomData представляет унифицированную структуру данных для комнаты умного дома.
type RoomData struct {
    ID                int               `json:"id"`
    Name              map[string]string `json:"name"`
    Temperature       string            `json:"temperature"`
    LastUpdate        time.Time         `json:"last_update"`
    TargetTemperature float64           `json:"target_temperature" bson:"target_temperature"`
    AcStatus          bool              `json:"ac_status"`
}

// [EN] DeviceData represents a physical BLE sensor payload prepared for frontend rendering.
// [RU] DeviceData представляет физический BLE-датчик для передачи на фронтенд.
type DeviceData struct {
    DeviceID    string    `json:"device_id"`
    Temperature float64   `json:"temperature"`
    Humidity    float64   `json:"humidity"`
    Battery     int       `json:"battery"`
    SensorType  string    `json:"sensor_type"`
    RSSI        int       `json:"rssi"`
    LastUpdate  time.Time `json:"last_update"`
}

// [EN] BLETelemetryInput is the DTO for ingesting raw JSON telemetry from scripts/ble_bridge.py.
// [RU] BLETelemetryInput — DTO для приема сырого JSON с телеметрией от скрипта scripts/ble_bridge.py.
type BLETelemetryInput struct {
    DeviceID    string  `json:"device_id"`
    RSSI        int     `json:"rssi"`
    Timestamp   int64   `json:"timestamp"`
    SensorType  string  `json:"sensor_type"`
    Temperature float64 `json:"temperature"`
    Humidity    float64 `json:"humidity"`
    Battery     int     `json:"battery"`
    VoltageMV   int     `json:"voltage_mv"`
}