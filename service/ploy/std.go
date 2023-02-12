package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/service/engine"
	"sync"
)

// BaseProvider 标准
type BaseProvider struct {
}

func (bp *BaseProvider) CanPut(pIdx int, tile int) bool {
	return true
}

func (bp *BaseProvider) Extras() []*mj.CardExtra {
	return []*mj.CardExtra{}
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

func (eval *abcEvaluation) Valid(ctx *engine.RoundCtx, raceIdx int, racePart mj.Cards, whoIdx int, whoTile int) bool {
	return len(racePart) == 2
}

func (eval *abcEvaluation) Next(ctx *engine.RoundCtx, raceIdx int, whoIdx int) RaceNext {
	return NextPut
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
		tp, tn := t[0], t[1]
		//非法
		if eval.illegals.Index(tp) != -1 || eval.illegals.Index(tn) != -1 {
			continue
		}
		//存在
		if tiles.Index(tp) != -1 && tiles.Index(tn) != -1 {
			effects = append(effects, t)
		}
	}
	return len(effects) > 0, effects
}

// 碰
type dddEvaluation struct {
	illegals mj.Cards
}

func (eval *dddEvaluation) Valid(ctx *engine.RoundCtx, raceIdx int, racePart mj.Cards, whoIdx int, whoTile int) bool {
	return len(racePart) == 2
}

func (eval *dddEvaluation) Next(ctx *engine.RoundCtx, raceIdx int, whoIdx int) RaceNext {
	return NextPut
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
}

func (eval *eeeeUpgradeEvaluation) Valid(ctx *engine.RoundCtx, raceIdx int, racePart mj.Cards, whoIdx int, whoTile int) bool {
	return len(racePart) == 2
}

func (eval *eeeeUpgradeEvaluation) Next(ctx *engine.RoundCtx, raceIdx int, whoIdx int) RaceNext {
	return NextTake
}

func (eval *eeeeUpgradeEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	//自杠 从已判断中的牌检索
	if raceIdx != whoIdx {
		return false, nil
	}
	tileCtx := ctx.Operating().GetTiles(raceIdx)
	//检索 已判定的牌
	races := tileCtx.Races
	hands := tileCtx.Hands

	plans := make([]mj.Cards, 0)
	for i := 0; i < len(races); i++ {
		raceItem := races[i]
		if raceItem.IsDDD() && hands.Index(raceItem[0]) != -1 {
			plans = append(plans, mj.Cards{raceItem[0]})
		}
	}
	return len(plans) > 0, plans
}

// 杠（自己）
type eeeeOwnEvaluation struct {
}

func (eval *eeeeOwnEvaluation) Valid(ctx *engine.RoundCtx, raceIdx int, racePart mj.Cards, whoIdx int, whoTile int) bool {
	return len(racePart) == 4
}

func (eval *eeeeOwnEvaluation) Next(ctx *engine.RoundCtx, raceIdx int, whoIdx int) RaceNext {
	return NextTake
}

func (eval *eeeeOwnEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	if raceIdx != whoIdx {
		return false, nil
	}
	//数量分组
	counts := make(map[int]int, 0)
	for _, t := range tiles {
		counts[t] = counts[t] + 1
	}

	plans := make([]mj.Cards, 0)
	for k, v := range counts {
		//有4张
		if v == 4 {
			plans = append(plans, mj.Cards{k, k, k, k})
		}
	}
	return len(plans) > 0, plans
}

// 杠（别人）
type eeeeEvaluation struct {
}

func (eval *eeeeEvaluation) Valid(ctx *engine.RoundCtx, raceIdx int, racePart mj.Cards, whoIdx int, whoTile int) bool {
	return len(racePart) == 3
}

func (eval *eeeeEvaluation) Next(ctx *engine.RoundCtx, raceIdx int, whoIdx int) RaceNext {
	return NextTake
}

func (eval *eeeeEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	//不能杠自己打的牌
	if raceIdx == whoIdx {
		return false, nil
	}
	//自己手牌中存在3张
	if len(tiles.Indexes(tile)) == 3 {
		return true, []mj.Cards{{tile, tile, tile}}
	}
	return false, nil
}

// 胡
type winEvaluation struct {
}

func (eval *winEvaluation) Valid(ctx *engine.RoundCtx, raceIdx int, racePart mj.Cards, whoIdx int, whoTile int) bool {
	return true
}

func (eval *winEvaluation) Next(ctx *engine.RoundCtx, raceIdx int, whoIdx int) RaceNext {
	return NextQuit
}

func (eval *winEvaluation) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {

	var winIntact mj.Cards
	if raceIdx == whoIdx {
		//自摸
		winIntact = tiles
	} else {
		//别人点炮
		winIntact = append(tiles, tile)
	}

	//只有两张,判断是否为将牌
	if len(winIntact) == 2 {
		return winIntact[0] == winIntact[1], []mj.Cards{{tile}}
	}

	//是否能胡牌
	comb := mj.NewWinChecker().Check(winIntact)
	if comb != nil {
		return true, []mj.Cards{{tile}}
	}
	return false, nil
}
