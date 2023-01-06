package store

import (
	"errors"
	"mahjong/server/api"
	"mahjong/server/engine"
	"sync"
)

type roomConfig struct {
	roomId string
	gc     *api.GameConfigure
	pc     *api.PaymentConfigure
}

//缓存
var roomConfigMap = &sync.Map{}

func GetRoomConfig(roomId string) (*api.GameConfigure, *api.PaymentConfigure) {
	cfg, ok := roomConfigMap.Load(roomId)
	if ok {
		roomCfg := cfg.(*roomConfig)
		return roomCfg.gc, roomCfg.pc
	}
	return nil, nil
}

func CreateRoom(roomId string, gc *api.GameConfigure, pc *api.PaymentConfigure) {
	roomConfigMap.Store(roomId, &roomConfig{
		roomId: roomId,
		gc:     gc,
		pc:     pc,
	})
}

//缓存
var posMap = &sync.Map{}

func CreatePosition(roomId string, pos *engine.Position) {
	posMap.Store(roomId, pos)
}

func UpdatePosition(roomId string, pos *engine.Position) {
	posMap.Store(roomId, pos)
}

func GetPosition(roomId string) (*engine.Position, error) {
	pos, ok := posMap.Load(roomId)
	if !ok {
		return nil, errors.New("not found")
	}
	return pos.(*engine.Position), nil
}
