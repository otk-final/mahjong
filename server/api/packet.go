package api

import "encoding/json"

type WebEvent int

const (
	GameReadyEvent WebEvent = 100
	TakeCardEvent  WebEvent = 101
	PutCardEvent   WebEvent = 102
)

type WebPacket[T any] struct {
	Event   WebEvent `json:"event"`
	EventId string   `json:"event_id"`
	Payload T        `json:"payload"`
}

// UnPacket 解码
func UnPacket[T any](packet []byte) (WebEvent, *T, error) {
	var wp WebPacket[T]
	err := json.Unmarshal(packet, &wp)
	if err != nil {
		return -1, nil, err
	}
	return wp.Event, &wp.Payload, nil
}
