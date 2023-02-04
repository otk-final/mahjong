package api

import (
	"github.com/google/uuid"
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
	Who      int `json:"who"`
	Duration int `json:"duration"`
}

type JoinPayload struct {
	NewPlayer *Player `json:"newPlayer"`
	Round     int     `json:"round"`
}

type BeginPayload struct {
	TurnIdx  int            `json:"turnIdx"`
	Remained int            `json:"remained"`
	Tiles    []*PlayerTiles `json:"tiles"`
}

func Packet[T any](code WebEvent, name string, event T) *WebPacket[T] {
	return &WebPacket[T]{Event: code, EventName: name, EventId: uuid.New().String(), Payload: event}
}
