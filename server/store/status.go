package store

import (
	"errors"
	"mahjong/server/api"
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

func CreateRoundCtx(roomId string, setting *api.GameConfigure, pos *engine.Position, exchanger *engine.Exchanger, handler engine.RoundOperation) *engine.RoundCtx {
	v, _ := roundCtxMap.LoadOrStore(roomId, engine.NewRoundCtx(0, setting, pos, exchanger, handler))
	return v.(*engine.RoundCtx)
}
