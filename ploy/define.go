package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/engine"
	"mahjong/server/store"
)

// GameDefine 游戏规则
type GameDefine interface {
	Init(gc *api.GameConfigure, pc *api.PaymentConfigure) engine.TileHandle
	Evaluate() map[api.RaceType]RaceEvaluate
	Finish() bool
	Quit()
}

// RaceEvaluate 碰，吃，杠，胡...评估
type RaceEvaluate interface {
	// Valid 可行判断
	Valid(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) bool
	// Plan 方案
	Plan(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) []mj.Cards
}
