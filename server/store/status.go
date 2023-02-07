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
	v, _ := roundCtxMap.LoadOrStore(roomId, engine.NewRoundCtx(0, pos, exchanger, handler))
	return v.(*engine.RoundCtx)
}
