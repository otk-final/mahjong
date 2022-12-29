package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/engine"
	"mahjong/server/store"
)

type LaiProvider struct {
	lai int
	cao int
	BaseProvider
}

func NewLaiProvider() GameDefine {
	return &LaiProvider{
		lai:          0,
		BaseProvider: BaseProvider{},
	}
}

func (lp *LaiProvider) Init(gc *api.GameConfigure, pc *api.PaymentConfigure) engine.TileHandle {

	//牌库 只有万，条，筒
	laiLib := mj.LoadLibrary(mj.WanCard, mj.TiaoCard, mj.TongCard)

	//init
	handler := initTileHandler(engine.NewDice(), gc.Nums, laiLib)

	//从前摸张牌，当前牌为朝天，下一张为癞牌
	cao := handler.table.Forward()
	var lai int
	switch cao {
	case mj.W9:
		lai = mj.W1
	case mj.T9:
		lai = mj.T1
	case mj.L9:
		lai = mj.L1
	default:
		lai++
	}

	lp.cao = cao
	lp.lai = lai
	return handler
}

func (lp *LaiProvider) Evaluate() map[api.RaceType]RaceEvaluate {
	return map[api.RaceType]RaceEvaluate{
		api.PairRace: &pairEvaluation{},
		api.EatRace:  &eatEvaluation{},
		api.GangRace: &gangEvaluation{},
		api.WinRace:  &winEvaluation{},
	}
}

type laiWinEvaluation struct {
}

func (eval laiWinEvaluation) Valid(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) bool {

	//一癞到底，胡牌时最多只能有一个癞子
	return false
}

func (eval laiWinEvaluation) Plan(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) []mj.Cards {
	return nil
}
