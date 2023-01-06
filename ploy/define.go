package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/engine"
)

// GameDefine 游戏规则
type GameDefine interface {
	// InitCtx 初始化
	InitCtx(gc *api.GameConfigure, pc *api.PaymentConfigure) engine.RoundOpsCtx
	// Handles 策略集
	Handles() map[api.RaceType]RaceEvaluate
	// Renew 从上下文中恢复
	Renew(ctx *engine.RoundCtx)
	// Finish 结束
	Finish() bool
	// Quit 退出
	Quit()
}

// RaceEvaluate 碰，吃，杠，胡...评估
type RaceEvaluate interface {
	// Eval 可行判断
	Eval(ctx *engine.RoundCtx, raceIdx, whoIdx, tile int) (bool, []mj.Cards)
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

type BaseRoundCtxHandler struct {
	gc      *api.GameConfigure
	pc      *api.PaymentConfigure
	table   *engine.Table
	tiles   map[int]*PlayerTiles
	profits map[int]*PlayerProfit
	custom  map[string]any
}

// PlayerTiles 玩家牌库
type PlayerTiles struct {
	hands      mj.Cards
	races      []mj.Cards
	outs       mj.Cards
	lastedTake int
	lastedPut  int
}

//PlayerProfit 玩家收益
type PlayerProfit struct {
}

func (b *BaseRoundCtxHandler) WithConfig() (*api.GameConfigure, *api.PaymentConfigure) {
	return b.gc, b.pc
}

func (b *BaseRoundCtxHandler) GetOuts(pIdx int) mj.Cards {
	return b.tiles[pIdx].outs
}

func (b *BaseRoundCtxHandler) GetHands(pIdx int) mj.Cards {
	return b.tiles[pIdx].hands
}

func (b *BaseRoundCtxHandler) GetRaces(pIdx int) []mj.Cards {
	return b.tiles[pIdx].races
}

func (b *BaseRoundCtxHandler) AddTake(pIdx int, tile int) {
	own := b.tiles[pIdx]

	own.lastedTake = tile
	own.hands = append(own.hands, tile)
}

func (b *BaseRoundCtxHandler) AddPut(pIdx int, tile int) {

	own := b.tiles[pIdx]
	own.lastedPut = tile

	//update hands
	tIdx := own.hands.Index(tile)
	own.hands = own.hands.Remove(tIdx)
	own.outs = append(own.outs, tile)
}

func (b *BaseRoundCtxHandler) AddRace(pIdx int, tiles mj.Cards, whoIdx int, tile int) {
	own := b.tiles[pIdx]
	race := append(tiles, tile)
	own.races = append(own.races, race)

	//移交
	who := b.tiles[whoIdx]
	who.outs = who.outs[:len(who.outs)-1]
}

func (b *BaseRoundCtxHandler) Forward(pIdx int) int {
	return b.table.Forward()
}

func (b *BaseRoundCtxHandler) Backward(pIdx int) int {
	return b.table.Backward()
}
