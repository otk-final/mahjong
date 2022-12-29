package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/engine"
	"mahjong/server/store"
	"sort"
)

// BaseProvider 标准
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
	own.hands = append(own.hands, tile)
}

func (b *BaseTileHandler) AddPut(pIdx int, tile int) {

	own := b.tiles[pIdx]
	own.lastedPut = tile

	//update hands
	tIdx := own.hands.Index(tile)
	own.hands = own.hands.Remove(tIdx)
	own.outs = append(own.outs, tile)
}

func (b *BaseTileHandler) AddRace(pIdx int, tiles mj.Cards, whoIdx int, tile int) {
	own := b.tiles[pIdx]
	race := append(tiles, tile)
	own.races = append(own.races, race)

	//移交
	who := b.tiles[whoIdx]
	who.outs = who.outs[:len(who.outs)-1]
}

func (b *BaseTileHandler) Forward(pIdx int) int {
	return b.table.Forward()
}

func (b *BaseTileHandler) Backward(pIdx int) int {
	return b.table.Backward()
}

func (bp *BaseProvider) Init(gc *api.GameConfigure, pc *api.PaymentConfigure) engine.TileHandle {
	//创建上下文处理器
	return startTileHandler(engine.NewDice(), gc.Nums, mj.Library)
}

func startTileHandler(dice int, players int, libs mj.Cards) *BaseTileHandler {

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
		api.DDDRace:  &dddEvaluation{},
		api.ABCRace:  &abcEvaluation{},
		api.EEEERace: &eeeeEvaluation{},
		api.WinRace:  &winEvaluation{},
	}
}

func NewBaseProvider() GameDefine {
	return &BaseProvider{}
}

// 吃
type abcEvaluation struct {
}

func (eval *abcEvaluation) Eval(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) (bool, []mj.Cards) {
	//只能吃上家出的牌
	if raceIdx-whoIdx != 1 || (raceIdx == (ctx.Position.Len()-1) && whoIdx != 0) {
		return false, nil
	}

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
	if len(effects) > 0 {
		return true, effects
	}
	return false, nil
}

// 碰
type dddEvaluation struct {
}

func (eval *dddEvaluation) Eval(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) (bool, []mj.Cards) {
	hands := ctx.Handler.GetHands(raceIdx)
	//只剩一张
	if len(hands) < 1 {
		return false, nil
	}
	//不能碰自己打的牌
	if raceIdx == whoIdx {
		return false, nil
	}

	temp := make(mj.Cards, len(hands))
	copy(temp, hands)

	sort.Ints(temp)
	tIdx := temp.Index(tile)
	if tIdx == -1 {
		return false, nil
	}
	//共2张牌一样
	ok := temp[tIdx+1] == tile
	if ok {
		return true, []mj.Cards{{tile, tile}}
	}
	return false, nil
}

// 杠
type eeeeEvaluation struct {
}

func (eval *eeeeEvaluation) Eval(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) (bool, []mj.Cards) {

	//自杠 从已判断中的牌检索
	if raceIdx == whoIdx {
		//检索 碰过的
		races := ctx.Handler.GetRaces(raceIdx)
		for i := 0; i < len(races); i++ {
			race := races[i]
			if len(race) == 3 && (race[0] == tile && race[1] == tile && race[2] == tile) {
				return true, []mj.Cards{{tile}}
			}
		}
		return false, nil
	}
	//杠别人 从手牌中检索
	temp := ctx.Handler.GetHands(raceIdx)
	if len(temp) < 3 {
		return false, nil
	}

	sort.Ints(temp)
	tIdx := temp.Index(tile)
	if tIdx == -1 {
		return false, nil
	}
	//共3张牌一样
	ok := temp[tIdx+1] == tile && temp[tIdx+2] == tile
	if ok {
		return true, []mj.Cards{{tile, tile, tile}}
	}
	return false, nil
}

// 胡
type winEvaluation struct {
}

func (eval *winEvaluation) Eval(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) (bool, []mj.Cards) {
	hands := ctx.Handler.GetHands(raceIdx)

	temp := make(mj.Cards, len(hands))
	copy(temp, hands)
	temp = append(temp, tile)

	//只有两张,判断是否为将牌
	if len(temp) == 2 {
		return temp[0] == temp[1], []mj.Cards{temp}
	}

	ok, comb := mj.NewWinChecker().Check(temp)
	if ok {
		//有效组合
		out := make([]mj.Cards, 0)
		out = append(out, comb.ABC...)
		out = append(out, comb.DDD...)
		out = append(out, comb.EE)
		return true, out
	}
	return false, nil
}

//听
type tingEvaluation struct {
}

func (eval *tingEvaluation) Eval(ctx *store.RoundCtx, raceIdx, whoIdx, tile int) (bool, []mj.Cards) {
	//TODO implement me
	panic("implement me")
}
