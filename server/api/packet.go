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
	QuitEvent
)

type WebPacket[T any] struct {
	Event     WebEvent `json:"event"`
	EventName string   `json:"eventName"`
	EventId   string   `json:"eventId"`
	Payload   T        `json:"payload"`
}

type QuitPayload struct {
	Reason string `json:"reason"`
}

type TakePayload struct {
	Who      int `json:"who"`
	Tile     int `json:"tile"`
	Remained int `json:"remained"`
}

type PutPayload struct {
	Who  int `json:"who"`
	Tile int `json:"tile"`
}

type RacePayload struct {
	RaceType RaceType `json:"raceType"`
	Who      int      `json:"who"`
	Target   int      `json:"target"`
	Tiles    mj.Cards `json:"tiles"`
	Tile     int      `json:"tile"`
	Interval int      `json:"interval"`
}
type WinPayload struct {
	Who    int          `json:"who"`
	Target int          `json:"target"`
	Tile   int          `json:"tile"`
	Tiles  *PlayerTiles `json:"tiles"`
	Effect RaceType     `json:"effect"`
}

type AckPayload struct {
	Who   int `json:"who"`
	AckId int `json:"ackId"`
}

type TurnPayload struct {
	Pre      int `json:"pre"`
	Who      int `json:"who"`
	Interval int `json:"interval"`
}

type JoinPayload struct {
	NewPlayer *Player `json:"newPlayer"`
}

type GamePayload struct {
	TurnIdx   int             `json:"turnIdx"`
	Interval  int             `json:"interval"`
	RecentIdx int             `json:"recentIdx"`
	Remained  int             `json:"remained"`
	Players   []*PlayerTiles  `json:"players"`
	Extras    []*mj.CardExtra `json:"extras"`
}

func Packet[T any](code WebEvent, name string, event T) *WebPacket[T] {
	return &WebPacket[T]{Event: code, EventName: name, EventId: uuid.New().String(), Payload: event}
}
