package models

import "time"

// RoomData represents the unified data structure for a smart home room
type RoomData struct {
	ID                int               `json:"id"`
	Name              map[string]string `json:"name"`
	Temperature       string            `json:"temperature"`
	LastUpdate        time.Time         `json:"last_update"`
	TargetTemperature float64           `json:"target_temperature" bson:"target_temperature"`
	AcStatus          bool              `json:"ac_status"`
}

// BLETelemetryInput — DTO для приема JSON от scripts/ble_bridge.py
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