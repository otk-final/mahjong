package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/engine"
	"mahjong/server/store"
	"sort"
)

type BaseProvider struct {
	dice int //骰子数
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
	own.hands = own.hands.Append(tile)
}

func (b *BaseTileHandler) AddPut(pIdx int, tile int) {

	own := b.tiles[pIdx]
	own.lastedPut = tile

	//update hands
	own.hands = own.hands.Remove(tile)
	own.outs = own.outs.Append(tile)
}

func (b *BaseTileHandler) AddRace(pIdx int, tiles mj.Cards, whoIdx int, tile int) {
	own := b.tiles[pIdx]
	comb := append(tiles, tile)
	own.races = append(own.races, comb)

	//移交
	who := b.tiles[whoIdx]
	who.outs = who.outs[:1]
}

func (b *BaseTileHandler) Forward(pIdx int) int {
	return b.table.Forward()
}

func (b *BaseTileHandler) Backward(pIdx int) int {
	return b.table.Backward()
}

func (bp *BaseProvider) Init(gc *api.GameConfigure, pc *api.PaymentConfigure) engine.TileHandle {
	//创建上下文处理器
	return initTileHandler(engine.NewDice(), gc.Nums, mj.Library)
}

func initTileHandler(dice int, players int, libs mj.Cards) *BaseTileHandler {

	//掷骰子，洗牌，发牌

	//全量牌
	tiles := engine.Shuffle(libs)

	//发牌
	tb := engine.NewTable(dice, tiles)
	members := tb.Distribution(players)

	//添加到上下文
	opsCtx := &BaseTileHandler{
		table:   tb,
		tiles:   make(map[int]*PlayerTiles, players),
		profits: make(map[int]*PlayerProfit, players),
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
	return &BaseProvider{}
}

// 吃
type eatEvaluation struct{}

func (eval *eatEvaluation) Valid(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) bool {
	//只能吃上家出的牌
	if raceIdx-whoIdx != 1 || (raceIdx == (ctx.Position.Len()-1) && whoIdx != 0) {
		return false
	}
	effects := eval.Plan(ctx, raceIdx, whoIdx, tile)
	return len(effects) > 0
}

func (eval *eatEvaluation) Plan(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) []mj.Cards {
	hands := ctx.Handler.GetHands(raceIdx)

	temp := make(mj.Cards, len(hands))
	copy(temp, hands)

	effects := make([]mj.Cards, 0)
	u1, u2 := tile+1, tile+2
	if temp.Index(u1) != -1 && temp.Index(u2) != -1 {
		effects = append(effects, mj.Cards{tile, u1, u2})
	}

	l1, l2 := tile-2, tile-1
	if temp.Index(l1) != -1 && temp.Index(l2) != -1 {
		effects = append(effects, mj.Cards{l1, l2, tile})
	}
	return effects
}

// 碰
type pairEvaluation struct{}

func (eval *pairEvaluation) Valid(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) bool {
	hands := ctx.Handler.GetHands(raceIdx)

	//只剩一张
	if len(hands) < 1 {
		return false
	}

	//不能碰自己打的牌
	if raceIdx == whoIdx {
		return false
	}

	temp := make(mj.Cards, len(hands))
	copy(temp, hands)

	sort.Ints(temp)
	tIdx := temp.Index(tile)
	if tIdx == -1 {
		return false
	}
	return temp[tIdx+1] == tile
}

func (eval *pairEvaluation) Plan(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) []mj.Cards {
	return []mj.Cards{{tile, tile}}
}

// 杠
type gangEvaluation struct {
}

func (eval *gangEvaluation) Valid(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) bool {
	//自杠 从已判断中的牌检索
	if raceIdx == whoIdx {
		//检索 碰过的
		races := ctx.Handler.GetRaces(raceIdx)
		for i := 0; i < len(races); i++ {
			race := races[i]
			if len(race) == 3 && (race[0] == tile && race[1] == tile && race[2] == tile) {
				return true
			}
		}
		return false
	}
	//杠别人 从手牌中检索
	temp := ctx.Handler.GetHands(raceIdx)
	if len(temp) < 3 {
		return false
	}

	sort.Ints(temp)
	tIdx := temp.Index(tile)
	if tIdx == -1 {
		return false
	}
	return temp[tIdx+1] == tile && temp[tIdx+2] == tile
}

func (eval *gangEvaluation) Plan(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) []mj.Cards {
	//自杠
	if raceIdx == whoIdx {
		return []mj.Cards{{tile}}
	}
	return []mj.Cards{{tile, tile, tile}}
}

// 胡
type winEvaluation struct {
}

func (eval *winEvaluation) Valid(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) bool {
	hands := ctx.Handler.GetHands(raceIdx)

	temp := make(mj.Cards, len(hands))
	copy(temp, hands)
	temp = append(temp, tile)

	//只有两张,判断是否为将牌
	if len(temp) == 2 {
		return temp[0] == temp[1]
	}
	ok, _ := mj.NewWinChecker().Check(temp)
	return ok
}

func (eval *winEvaluation) Plan(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) []mj.Cards {

	hands := ctx.Handler.GetHands(raceIdx)
	temp := make(mj.Cards, len(hands))
	copy(temp, hands)
	temp = append(temp, tile)

	//默认只有一种胡牌 牌型
	return []mj.Cards{temp}
}
