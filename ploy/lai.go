package ploy

import (
	"mahjong/server/api"
	"mahjong/server/engine"
)

type LaiProvider struct {
	lai int
	BaseProvider
}

func NewLaiProvider() GameDefine {
	return &LaiProvider{
		lai:          0,
		BaseProvider: BaseProvider{},
	}
}

func (lp *LaiProvider) Init(gc *api.GameConfigure, pc *api.PaymentConfigure) engine.TileHandle {

	//init
	handler := lp.initOps(gc, pc)

	//从前摸张牌，癞牌
	lp.lai = handler.table.Forward()

	return handler
}
