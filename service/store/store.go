package store

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"hash/crc32"
	"mahjong/server/api"
	"mahjong/service/engine"
	"sync"
)

type roomConfig struct {
	roomId  string
	setting *api.GameConfigure
}

//缓存
var roomConfigMap = &sync.Map{}

func GetRoomConfig(roomId string) *api.GameConfigure {
	cfg, ok := roomConfigMap.Load(roomId)
	if ok {
		roomCfg := cfg.(*roomConfig)
		return roomCfg.setting
	}
	return nil
}

// 生成房间号
func roomIdGen() string {
	id := crc32.ChecksumIEEE([]byte(uuid.New().String()))
	return fmt.Sprintf("%d", id)
}

func CreateRoom(setting *api.GameConfigure) string {
	roomId := roomIdGen()
	roomConfigMap.Store(roomId, &roomConfig{
		roomId:  roomId,
		setting: setting,
	})
	return roomId
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
