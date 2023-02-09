package store

import (
	"errors"
	"mahjong/server/api"
	engine2 "mahjong/service/engine"
	"sync"
)

var roundCtxMap = &sync.Map{}

func LoadRoundCtx(roomId string, acctId string) (*engine2.RoundCtx, error) {
	v, ok := roundCtxMap.Load(roomId)
	if !ok {
		return nil, errors.New("not found")
	}
	ctx := v.(*engine2.RoundCtx)

	//check 用户
	_, err := ctx.Player(acctId)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func CreateRoundCtx(roomId string, setting *api.GameConfigure, pos *engine2.Position, exchanger *engine2.Exchanger, handler engine2.RoundOperation) *engine2.RoundCtx {
	v, _ := roundCtxMap.LoadOrStore(roomId, engine2.NewRoundCtx(0, setting, pos, exchanger, handler))
	return v.(*engine2.RoundCtx)
}
