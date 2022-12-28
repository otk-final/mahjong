package ploy

import (
	"mahjong/server/api"
	"mahjong/server/engine"
)

type GameDefine interface {
	Init(gc *api.GameConfigure, pc *api.PaymentConfigure) engine.TileHandle
	Finish() bool
	Quit()
}
