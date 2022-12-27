package ploy

import (
	"mahjong/server/api"
)

type GameDefine interface {
	Init(gc *api.GameConfigure, pc *api.PaymentConfigure)
	Finish() bool
	Quit()
}
