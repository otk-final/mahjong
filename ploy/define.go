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
	Finish()
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
	tiles   map[int]*api.PlayerTiles
	profits map[int]*api.PlayerProfits
	custom  map[string]any
}

func (b *BaseRoundCtxHandler) WithConfig() (*api.GameConfigure, *api.PaymentConfigure) {
	return b.gc, b.pc
}

func (b *BaseRoundCtxHandler) GetTiles(pIdx int) *api.PlayerTiles {
	return b.tiles[pIdx]
}

func (b *BaseRoundCtxHandler) GetProfits(pIdx int) *api.PlayerProfits {
	return b.profits[pIdx]
}

func (b *BaseRoundCtxHandler) withHands(pIdx int) mj.Cards {
	return b.tiles[pIdx].Hands
}
func (b *BaseRoundCtxHandler) withRaces(pIdx int) []mj.Cards {
	return b.tiles[pIdx].Races
}
func (b *BaseRoundCtxHandler) withOuts(pIdx int) mj.Cards {
	return b.tiles[pIdx].Outs
}

func (b *BaseRoundCtxHandler) AddTake(pIdx int, tile int) {
	own := b.tiles[pIdx]

	own.LastedTake = tile
	own.Hands = append(own.Hands, tile)
}

func (b *BaseRoundCtxHandler) AddPut(pIdx int, tile int) {

	own := b.tiles[pIdx]
	own.LastedPut = tile

	//update hands
	tIdx := own.Hands.Index(tile)
	own.Hands = own.Hands.Remove(tIdx)
	own.Outs = append(own.Outs, tile)
}

func (b *BaseRoundCtxHandler) AddRace(pIdx int, tiles mj.Cards, whoIdx int, tile int) {
	own := b.tiles[pIdx]
	race := append(tiles, tile)
	own.Races = append(own.Races, race)

	//移交
	who := b.tiles[whoIdx]
	who.Outs = who.Outs[:len(who.Outs)-1]
}

func (b *BaseRoundCtxHandler) Forward(pIdx int) int {
	return b.table.Forward()
}

func (b *BaseRoundCtxHandler) Backward(pIdx int) int {
	return b.table.Backward()
}

func (b *BaseRoundCtxHandler) Remained() int {
	return b.table.Remains()
}
