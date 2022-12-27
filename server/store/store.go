package store

import (
	"mahjong/server/api"
	"mahjong/server/engine"
)

func GetRoomConfig(roomId string) (*api.GameConfigure, *api.PaymentConfigure) {
	return nil, nil
}

func CreateRoom(roomId string, gc *api.GameConfigure, pc *api.PaymentConfigure) {

}

func CreatePosition(roomId string, pos *engine.Position) {

}

func UpdatePosition(roomId string, pos *engine.Position) error {
	return nil
}

func GetPosition(roomId string) (*engine.Position, error) {
	return nil, nil
}

func GetCountdown(roomId string) (*engine.Countdown, error) {
	return nil, nil
}
