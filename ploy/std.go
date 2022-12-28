package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/engine"
	"mahjong/server/store"
)

type BaseProvider struct {
	dice  int //骰子数
	tiles mj.Cards
}

type BaseTileHandler struct {
	table   *engine.Table
	tiles   map[int]*PlayerTiles
	profits map[int]*PlayerProfit
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

func (b *BaseTileHandler) GetOuts(pIdx int) mj.Cards {
	return b.tiles[pIdx].outs
}

func (b *BaseTileHandler) GetHands(pIdx int) mj.Cards {
	return b.tiles[pIdx].hands
}

func (b *BaseTileHandler) GetRaces(pIdx int) []mj.Cards {
	return b.tiles[pIdx].races
}

func (b *BaseTileHandler) AddTake(pIdx int, tile int) {
	own := b.tiles[pIdx]

	own.lastedTake = tile
	own.hands = append(own.hands, tile)
}

func (b *BaseTileHandler) AddPut(pIdx int, tile int) {

	own := b.tiles[pIdx]

	idx := -1
	for i, t := range own.hands {
		if t == tile {
			idx = i
			break
		}
	}
	if idx == -1 {
		return
	}

	own.lastedPut = tile

	//update hands
	own.hands = append(own.hands[:idx], own.hands[idx+1:]...)
	own.outs = append(own.outs, tile)
}

func (b *BaseTileHandler) AddRace(pIdx int, tiles mj.Cards, whoIdx int, tile int) {
	own := b.tiles[pIdx]
	comb := append(tiles, tile)
	own.races = append(own.races, comb)
}

func (b *BaseTileHandler) Forward(pIdx int) int {
	return b.table.Forward()
}

func (b *BaseTileHandler) Backward(pIdx int) int {
	return b.table.Backward()
}

func (bp *BaseProvider) Init(gc *api.GameConfigure, pc *api.PaymentConfigure) engine.TileHandle {
	//创建上下文处理器
	return bp.initOps(gc, pc)
}

func (bp *BaseProvider) initOps(gc *api.GameConfigure, pc *api.PaymentConfigure) *BaseTileHandler {

	//掷骰子，洗牌，发牌
	bp.dice = engine.NewDice()

	//洗牌
	tiles := engine.Shuffle(bp.tiles)

	//发牌
	tb := engine.NewTable(bp.dice, tiles)
	members := tb.Distribution(gc.Nums)

	//添加到上下文
	opsCtx := &BaseTileHandler{
		table:   tb,
		tiles:   make(map[int]*PlayerTiles, gc.Nums),
		profits: make(map[int]*PlayerProfit, gc.Nums),
	}

	//保存牌库
	for k, v := range members {
		opsCtx.tiles[k] = &PlayerTiles{
			hands:      v,
			races:      make([]mj.Cards, 0),
			outs:       make(mj.Cards, 0),
			lastedTake: 0,
			lastedPut:  0,
		}
	}

	return opsCtx
}

func (bp *BaseProvider) Finish() bool {
	return false
}

func (bp *BaseProvider) Quit() {

}

func (bp *BaseProvider) Evaluate() map[api.RaceType]RaceEvaluate {
	return map[api.RaceType]RaceEvaluate{
		api.PairRace: &pairEvaluation{},
		api.EatRace:  &eatEvaluation{},
		api.GangRace: &gangEvaluation{},
		api.WinRace:  &winEvaluation{},
	}
}

func NewBaseProvider() GameDefine {
	return &BaseProvider{
		dice:  0,
		tiles: mj.Library,
	}
}

// 吃
type eatEvaluation struct{}

func (eval *eatEvaluation) Valid(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) bool {
	//只能吃上家出的牌
	if raceIdx-whoIdx != 1 || (raceIdx == (ctx.Position.Len()-1) && whoIdx != 0) {
		return false
	}
	return ctx.Handler.GetHands(raceIdx).HasList(tile)
}

func (eval *eatEvaluation) Plan(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) []mj.Cards {
	//TODO implement me
	panic("implement me")
}

// 碰
type pairEvaluation struct{}

func (eval *pairEvaluation) Valid(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) bool {
	//不能碰自己打的牌
	if raceIdx == whoIdx {
		return false
	}
	return ctx.Handler.GetHands(raceIdx).HasPair(tile)
}

func (eval *pairEvaluation) Plan(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) []mj.Cards {
	return nil
}

// 杠
type gangEvaluation struct {
}

func (eval *gangEvaluation) Valid(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) bool {
	//自杠 从已判断中的牌检索
	if raceIdx == whoIdx {
		races := ctx.Handler.GetRaces(raceIdx)
		for _, comb := range races {
			if comb.HasGang(tile) {
				return true
			}
		}
	}
	return ctx.Handler.GetHands(raceIdx).HasGang(tile)
}

func (eval *gangEvaluation) Plan(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) []mj.Cards {
	return nil
}

// 胡
type winEvaluation struct {
}

func (eval *winEvaluation) Valid(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) bool {
	hands := ctx.Handler.GetHands(raceIdx)

	return true
}

func (eval *winEvaluation) Plan(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) []mj.Cards {
	//TODO implement me
	panic("implement me")
}
