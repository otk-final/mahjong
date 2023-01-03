package api

import (
	"github.com/google/uuid"
)

type WebEvent int

type WebPacket[T any] struct {
	Event   WebEvent `json:"event"`
	EventId string   `json:"event_id"`
	Payload T        `json:"payload"`
}

type TakePayload struct {
	Who   int `json:"who"`
	Round int `json:"round"`
	Tile  int `json:"take"`
}
type PutPayload struct {
	Who   int `json:"who"`
	Round int `json:"round"`
	Tile  int `json:"put"`
}

type RacePayload struct {
	Who       int      `json:"who"`
	Round     int      `json:"round"`
	RaceType  RaceType `json:"raceType"`
	HandTiles []int    `json:"handTiles"`
	Tile      int      `json:"tile"`
}
type AckPayload struct {
	Who   int `json:"who"`
	Round int `json:"round"`
	AckId int `json:"ackId"`
}

type NextPayload struct {
	Who int `json:"who"`
}

type JoinPayload struct {
	Member *Player `json:"member"`
	Round  int     `json:"round"`
}

func Packet[T any](code WebEvent, event T) *WebPacket[T] {
	return &WebPacket[T]{Event: code, EventId: uuid.New().String(), Payload: event}
}
