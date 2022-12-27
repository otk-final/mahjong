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

type RaceType int

var RaceNames = map[RaceType]string{
	WinRace:  "胡",
	PairRace: "碰",
	EatRace:  "吃",
	GangRace: "杠",
}

const (
	// WinRace 胡
	WinRace RaceType = iota + 1
	// PairRace 碰
	PairRace
	// EatRace 吃
	EatRace
	// GangRace 杠
	GangRace
)

type TakePayload struct {
	Who   int `json:"who"`
	Round int `json:"round"`
	Take  int `json:"take"`
}
type PutPayload struct {
	Who   int `json:"who"`
	Round int `json:"round"`
	Put   int `json:"put"`
}

type PutAckPayload struct {
	AckId int `json:"ackId"`
	PutPayload
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
