package api

import "encoding/json"

type WebEvent int

const (
	GameStart WebEvent = 100
	P         WebEvent = 100
)

type WebPacket[T any] struct {
	Event   int    `json:"event"`
	EventId string `json:"event_id"`
	Payload T      `json:"payload"`
}

// InitializeGamePayload 初始牌
type InitializeGamePayload struct {
	Cards []int `json:"cards"`
}

// UnPacket 解码
func UnPacket[T any](packet []byte) (int, *T, error) {
	var wp WebPacket[T]
	err := json.Unmarshal(packet, &wp)
	if err != nil {
		return -1, nil, err
	}
	return wp.Event, &wp.Payload, nil
}
