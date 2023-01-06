package store

import (
	"errors"
	"mahjong/server/engine"
	"sync"
)

var roundCtxMap = &sync.Map{}

func LoadRoundCtx(roomId string, acctId string) (*engine.RoundCtx, error) {
	v, ok := roundCtxMap.Load(roomId)
	if !ok {
		return nil, errors.New("not found")
	}
	ctx := v.(*engine.RoundCtx)

	//check 用户
	_, err := ctx.Player(acctId)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func RegisterRoundCtx(roomId string, pos *engine.Position, exchanger *engine.Exchanger, handler engine.RoundOpsCtx) *engine.RoundCtx {

	v, ok := roundCtxMap.Load(roomId)
	var ctx *engine.RoundCtx
	if ok {
		ctx = v.(*engine.RoundCtx)
		ctx.Round++
	} else {
		ctx = &engine.RoundCtx{Round: 1}
	}

	ctx.Position = pos
	ctx.Handler = handler
	ctx.Exchanger = exchanger
	roundCtxMap.Store(roomId, ctx)
	return ctx
}
