package api

import (
	"github.com/google/uuid"
	"mahjong/mj"
)

type WebEvent int

const (
	JoinEvent WebEvent = iota + 100
	ExitEvent
	BeginEvent
	TakeEvent
	PutEvent
	RaceEvent
	WinEvent
	AckEvent
	TurnEvent
)

type WebPacket[T any] struct {
	Event     WebEvent `json:"event"`
	EventName string   `json:"eventName"`
	EventId   string   `json:"eventId"`
	Payload   T        `json:"payload"`
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
	Other     int      `json:"other"`
	Round     int      `json:"round"`
	RaceType  RaceType `json:"raceType"`
	HandTiles []int    `json:"handTiles"`
	Tile      int      `json:"tile"`
}

type WinPayload struct {
	Who       int   `json:"who"`
	Other     int   `json:"other"`
	Round     int   `json:"round"`
	HandTiles []int `json:"handTiles"`
	Tile      int   `json:"tile"`
}

type AckPayload struct {
	Who   int `json:"who"`
	Round int `json:"round"`
	AckId int `json:"ackId"`
}

type TurnPayload struct {
	Who int `json:"who"`
}

type JoinPayload struct {
	Members []*Player `json:"members"`
	Round   int       `json:"round"`
}

type BeginPayload struct {
	Who   int      `json:"who"`
	Turn  bool     `json:"turn"`
	Hands mj.Cards `json:"hands"`
}

func Packet[T any](code WebEvent, name string, event T) *WebPacket[T] {
	return &WebPacket[T]{Event: code, EventName: name, EventId: uuid.New().String(), Payload: event}
}
