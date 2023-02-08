package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/engine"
	"sort"
	"sync"
)

// BaseProvider 标准
type BaseProvider struct {
	dice int //骰子数
}

func (bp *BaseProvider) Renew(ctx *engine.RoundCtx) {

}

func (bp *BaseProvider) InitCtx(gc *api.GameConfigure, pc *api.PaymentConfigure) engine.RoundOpsCtx {
	//创建上下文处理器
	opsCtx := startRoundCtxHandler(engine.NewDice(), gc.Nums, mj.Library)
	opsCtx.gc = gc
	opsCtx.pc = pc
	return opsCtx
}

func startRoundCtxHandler(dice int, players int, libs mj.Cards) *BaseRoundCtxHandler {

	//掷骰子，洗牌，发牌

	//全量牌
	tiles := engine.Shuffle(libs)

	//发牌
	tb := engine.NewTable(dice, tiles)
	members := tb.Distribution(players)

	//添加到上下文
	ctxOps := &BaseRoundCtxHandler{
		lock:    sync.Mutex{},
		table:   tb,
		tiles:   make(map[int]*api.PlayerTiles, players),
		profits: make(map[int]*api.PlayerProfits, players),
		custom:  make(map[string]any, 0),

		recentAction: -1,
		recentIdx:    -1,
		recenter:     make(map[int]*BaseRecenter, 0),
	}

	//保存牌库 初始化
	for k, v := range members {
		ctxOps.tiles[k] = &api.PlayerTiles{
			Idx:   k,
			Hands: v,
			Races: make([]mj.Cards, 0),
			Outs:  make(mj.Cards, 0),
		}
		ctxOps.recenter[k] = &BaseRecenter{idx: k, put: 0, take: 0, race: nil}
	}
	return ctxOps
}

func (bp *BaseProvider) Finish() {

}

func (bp *BaseProvider) Quit() {

}

func (bp *BaseProvider) Handles() map[api.RaceType]RaceEvaluate {
	return map[api.RaceType]RaceEvaluate{
		api.DDDRace:  &dddEvaluation{},
		api.ABCRace:  &abcEvaluation{},
		api.EEEERace: &eeeeEvaluation{},
		api.WinRace:  &winEvaluation{},
	}
}

func newBaseProvider() GameDefine {
	return &BaseProvider{}
}

// 吃
type abcEvaluation struct {
}

func isUpperIdx(mineIdx, whoIdx, members int) bool {
	//只能吃上家出的牌
	limit := mineIdx - whoIdx
	if limit == 1 || limit == (members-1)*-1 {
		return true
	}
	return false
}

func (eval *abcEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {

	//只能吃上家出的牌
	if !isUpperIdx(raceIdx, whoIdx, ctx.Pos().Num()) {
		return false, nil
	}
	effects := make([]mj.Cards, 0)
	options := [][]int{{tile + 1, tile + 2}, {tile - 2, tile - 1}, {tile - 1, tile + 1}}
	for _, t := range options {
		if tiles.Index(t[0]) == -1 || tiles.Index(t[1]) == -1 {
			continue
		}
		effects = append(effects, mj.Cards{t[0], t[1]})
	}
	return len(effects) > 0, effects
}

// 碰
type dddEvaluation struct {
}

func (eval *dddEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	//只剩一张 且 不能碰自己打的牌
	if len(tiles) <= 1 || raceIdx == whoIdx {
		return false, nil
	}

	//正序后，查询是否存在
	sort.Ints(tiles)
	tIdx := tiles.Index(tile)
	if tIdx == -1 {
		return false, nil
	}
	//共2张牌一样
	ok := len(tiles) > tIdx+1 && tiles[tIdx+1] == tile
	if ok {
		return true, []mj.Cards{{tile, tile}}
	}
	return false, nil
}

// 杠
type eeeeEvaluation struct {
}

func (eval *eeeeEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {

	//自杠 从已判断中的牌检索
	tileCtx := ctx.HandlerCtx().GetTiles(raceIdx)
	if raceIdx == whoIdx {
		//检索 碰过的
		races := tileCtx.Races
		for i := 0; i < len(races); i++ {
			race := races[i]
			if len(race) == 3 && (race[0] == tile && race[1] == tile && race[2] == tile) {
				return true, []mj.Cards{{tile}}
			}
		}
		return false, nil
	}
	//杠别人 从手牌中检索
	if len(tiles) < 3 {
		return false, nil
	}

	sort.Ints(tiles)
	tIdx := tiles.Index(tile)
	if tIdx == -1 {
		return false, nil
	}
	//共3张牌一样
	ok := len(tiles) > tIdx+2 && tiles[tIdx+1] == tile && tiles[tIdx+2] == tile
	if ok {
		return true, []mj.Cards{{tile, tile, tile}}
	}
	return false, nil
}

// 胡
type winEvaluation struct {
}

func (eval *winEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	hands := ctx.HandlerCtx().GetTiles(raceIdx).Hands.Clone()
	hands = append(hands, tile)

	//只有两张,判断是否为将牌
	if len(hands) == 2 {
		return hands[0] == hands[1], []mj.Cards{hands}
	}

	comb := mj.NewWinChecker().Check(hands)
	if comb != nil {
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

func (eval *tingEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	//TODO implement me
	panic("implement me")
}
