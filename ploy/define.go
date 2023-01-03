package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/engine"
	"mahjong/server/store"
)

// GameDefine 游戏规则
type GameDefine interface {
	// Init 初始化
	Init(gc *api.GameConfigure, pc *api.PaymentConfigure) engine.RoundCtx
	// HandleMapping 策略集
	HandleMapping() map[api.RaceType]RaceEvaluate
	// Renew 从上下文中恢复
	Renew(ctx *store.RoundCtx)
	// Finish 结束
	Finish() bool
	// Quit 退出
	Quit()
}

// RaceEvaluate 碰，吃，杠，胡...评估
type RaceEvaluate interface {
	// Eval 可行判断
	Eval(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) (bool, []mj.Cards)
}

func NewProvider(mode string) GameDefine {
	switch mode {
	case "std": //标准
		return newBaseProvider()
	case "lai": //赖子
		return newLaiProvider()
	case "k5x": //卡5星
		break
	case "7d": //七对
		break
	case "sc": //四川
		break
	case "gz": //广东
		break
	}
	return nil
}
