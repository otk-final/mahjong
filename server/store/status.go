package store

import (
	"mahjong/server/api"
	"mahjong/server/engine"
)

type RoundCtx struct {
	Round     int
	Position  *engine.Position
	Exchanger *engine.Exchanger
	Handler   engine.TileHandle
}

func (ctx *RoundCtx) Player(acctId string) (*api.Player, error) {
	return ctx.Position.Index(acctId)
}

func LoadRoundCtx(roomId string, acctId string) (*RoundCtx, error) {
	return nil, nil
}

func RegisterRoundCtx(roomId string, pos *engine.Position, exchanger *engine.Exchanger, handler engine.TileHandle) {

}
