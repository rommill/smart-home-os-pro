package models

import "time"

// RoomData represents the unified data structure for a smart home room
type RoomData struct {
	ID                int               `json:"id"`
	Name              map[string]string `json:"name"`
	Temperature       string            `json:"temperature"`
	LastUpdate        time.Time         `json:"last_update"`
	TargetTemperature float64           `json:"target_temperature"` // This field was missing!
	AcStatus          bool              `json:"ac_status"`          // Ready for AC automation
}