package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/service/engine"
	"sort"
	"sync"
)

// BaseProvider 标准
type BaseProvider struct {
}

func (bp *BaseProvider) Renew(ctx *engine.RoundCtx) GameDefine {
	return bp
}

func (bp *BaseProvider) InitOperation(setting *api.GameConfigure) engine.RoundOperation {

	//初始牌库 全量牌
	mjLib := mj.LoadLibrary()

	//创建上下文处理器
	opsCtx := startRoundCtxHandler(setting.Nums, mjLib)
	opsCtx.setting = setting
	return opsCtx
}

func startRoundCtxHandler(players int, libs mj.Cards) *BaseRoundCtxHandler {

	//掷骰
	dice := engine.NewDice()

	//洗牌
	tiles := engine.Shuffle(libs)

	//开桌
	tb := engine.NewTable(dice, tiles)

	//发牌
	members := tb.Distribution(players)

	//添加到上下文
	ctxOps := &BaseRoundCtxHandler{
		lock: sync.Mutex{},
		//玩家数据
		table:   tb,
		tiles:   make(map[int]*api.PlayerTiles, players),
		profits: make(map[int]*api.PlayerProfits, players),
		//最新数据
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

func (bp *BaseProvider) Handles() map[api.RaceType]RaceEvaluator {
	return map[api.RaceType]RaceEvaluator{
		api.DDDRace:         &dddEvaluation{},
		api.ABCRace:         &abcEvaluation{},
		api.EEEERace:        &eeeeEvaluation{},
		api.EEEEOwnRace:     &eeeeOwnEvaluation{},
		api.EEEEUpgradeRace: &eeeeUpgradeEvaluation{},
		api.WinRace:         &winEvaluation{},
	}
}

func newBaseProvider() GameDefine {
	return &BaseProvider{}
}

// 吃
type abcEvaluation struct {
	illegals mj.Cards
}

func isUpperIdx(mineIdx, whoIdx, members int) bool {
	//只能吃上家出的牌
	limit := mineIdx - whoIdx
	if limit == 1 || limit == (members-1)*-1 {
		return true
	}
	return false
}

func RaceTilesMerge(race api.RaceType, tiles mj.Cards, tile int) mj.Cards {
	switch race {
	case api.CaoRace, api.ABCRace, api.DDDRace:
		//合并
		result := append(tiles.Clone(), tile)
		sort.Ints(result)
		return result
	case api.EEEERace, api.EEEEOwnRace, api.EEEEUpgradeRace:
		return mj.Cards{tile, tile, tile, tile}
	case api.GuiRace, api.LaiRace:
		//单牌
		return mj.Cards{tile}
	}
	return mj.Cards{}
}

func (eval *abcEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {

	//只能吃上家出的牌
	if !isUpperIdx(raceIdx, whoIdx, ctx.Pos().Num()) || eval.illegals.Index(tile) != -1 {
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
	illegals mj.Cards
}

func (eval *dddEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	//不能碰自己打的牌
	if raceIdx == whoIdx || eval.illegals.Index(tile) != -1 {
		return false, nil
	}
	existLen := len(tiles.Indexes(tile))
	if existLen == 2 {
		return true, []mj.Cards{{tile, tile}}
	}
	return false, nil
}

// 杠（碰升级）
type eeeeUpgradeEvaluation struct {
	illegals mj.Cards
}

func (eval *eeeeUpgradeEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	//自杠 从已判断中的牌检索
	if raceIdx != whoIdx || eval.illegals.Index(tile) != -1 {
		return false, nil
	}
	tileCtx := ctx.Operating().GetTiles(raceIdx)
	//检索 碰过的
	races := tileCtx.Races
	for i := 0; i < len(races); i++ {
		existIdx := races[i].Indexes(tile)
		if len(existIdx) == 3 {
			return true, []mj.Cards{{tile}}
		}
	}
	return false, nil
}

// 杠（自己）
type eeeeOwnEvaluation struct {
	illegals mj.Cards
}

func (eval *eeeeOwnEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	if raceIdx != whoIdx || eval.illegals.Index(tile) != -1 {
		return false, nil
	}
	existLen := len(tiles.Indexes(tile))
	if existLen == 4 {
		return true, []mj.Cards{{tile, tile, tile, tile}}
	}
	return false, nil
}

// 杠（别人）
type eeeeEvaluation struct {
	illegals mj.Cards
}

func (eval *eeeeEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	//不能杠自己打的牌
	if raceIdx == whoIdx || eval.illegals.Index(tile) != -1 {
		return false, nil
	}
	existLen := len(tiles.Indexes(tile))
	if existLen == 3 {
		return true, []mj.Cards{{tile, tile, tile}}
	}
	return false, nil
}

// 胡
type winEvaluation struct {
}

func (eval *winEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {

	hands := tiles.Clone()
	//非自己手牌 合并目标牌后进行判定
	if raceIdx != whoIdx {
		hands = append(hands, tile)
	}

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
