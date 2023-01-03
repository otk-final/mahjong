package store

import (
	"errors"
	"mahjong/server/api"
	"mahjong/server/engine"
	"sync"
)

type RoundCtx struct {
	Round     int
	Position  *engine.Position
	Exchanger *engine.Exchanger
	Handler   engine.RoundCtxHandle
}

func (ctx *RoundCtx) Player(acctId string) (*api.Player, error) {
	return ctx.Position.Index(acctId)
}

var roundCtxMap = &sync.Map{}

func LoadRoundCtx(roomId string, acctId string) (*RoundCtx, error) {
	v, ok := roundCtxMap.Load(roomId)
	if !ok {
		return nil, errors.New("not found")
	}
	ctx := v.(*RoundCtx)

	//check 用户
	_, err := ctx.Player(acctId)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func RegisterRoundCtx(roomId string, pos *engine.Position, exchanger *engine.Exchanger, handler engine.RoundCtxHandle) {

	v, ok := roundCtxMap.Load(roomId)
	ctx := v.(*RoundCtx)
	if ok {
		ctx.Round++
	} else {
		ctx = &RoundCtx{Round: 1}
	}

	ctx.Position = pos
	ctx.Handler = handler
	ctx.Exchanger = exchanger
	roundCtxMap.Store(roomId, ctx)
}
