package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/service/engine"
)

// LaiProvider 癞子
type LaiProvider struct {
	BaseProvider

	tileLai int
	tileCao int
	tileGui int

	hasGui bool //是否有鬼牌（红中）
	noABC  bool //不能吃
	noWin  bool //不能放茺
	unique bool //一癞到底 有且只能有一个癞子胡牌
}

func newLaiProvider() GameDefine {
	return &LaiProvider{
		BaseProvider: BaseProvider{},
	}
}

type LaiRoundCtxHandler struct {
	*BaseRoundCtxHandler
	Cao int
	Lai int
	Gui int
}

func (lp *LaiProvider) CanPut(pIdx int, tile int) bool {
	return lp.tileGui != tile && lp.tileLai != tile
}

func (lp *LaiProvider) Renew(ctx *engine.RoundCtx) GameDefine {
	ctxHandler := ctx.Operating().(*LaiRoundCtxHandler)
	lp.tileLai = ctxHandler.Lai
	lp.tileCao = ctxHandler.Cao
	lp.tileGui = ctxHandler.Gui
	return lp
}

func (lp *LaiProvider) InitOperation(setting *api.GameConfigure) engine.RoundOperation {

	//牌库 只有万，条，筒
	laiLib := mj.LoadLibrary(mj.WanCard, mj.TiaoCard, mj.TongCard)
	if lp.hasGui {
		laiLib = append(laiLib, mj.Zh, mj.Zh, mj.Zh, mj.Zh)
	}

	//init
	handler := startRoundCtxHandler(setting.Nums, laiLib)
	handler.setting = setting

	//如果摸到红中，则继续摸
	var cao int
	for {
		cao = handler.table.Forward()
		if cao != mj.Zh {
			break
		}
		//todo 将牌放回原初
		//handler.table.
	}

	//从前摸张牌，当前牌为朝天，下一张为癞牌
	var lai int
	switch cao {
	case mj.W9:
		lai = mj.W1
		break
	case mj.T9:
		lai = mj.T1
	case mj.L9:
		lai = mj.L1
	default:
		lai++
	}
	return &LaiRoundCtxHandler{
		BaseRoundCtxHandler: handler,
		Cao:                 cao,
		Lai:                 lai,
	}
}

func (lp *LaiProvider) Handles() map[api.RaceType]RaceEvaluator {

	illegals := mj.Cards{lp.tileLai, lp.tileCao, lp.tileGui}

	evalMap := map[api.RaceType]RaceEvaluator{
		api.DDDRace:         &dddEvaluation{illegals: illegals},
		api.ABCRace:         &abcWithLai{tileLai: lp.tileLai, tileGui: lp.tileGui, abcEvaluation: abcEvaluation{illegals: illegals}},
		api.EEEERace:        &eeeeEvaluation{illegals: illegals},
		api.EEEEOwnRace:     &eeeeOwnEvaluation{illegals: illegals},
		api.EEEEUpgradeRace: &eeeeUpgradeEvaluation{illegals: illegals},
		api.CaoRace:         &dddEvaluation{mj.Cards{lp.tileLai, lp.tileGui}}, //忽略朝天牌
		api.LaiRace:         &fixWithLai{tile: lp.tileLai},
		api.GuiRace:         &fixWithLai{tile: lp.tileGui},
		api.WinRace:         &winWithLai{tileLai: lp.tileLai, tileGui: lp.tileGui, noWin: lp.noWin, unique: lp.unique},
	}
	//不能吃
	if lp.noABC {
		delete(evalMap, api.ABCRace)
	}
	return evalMap
}

type abcWithLai struct {
	tileLai int
	tileGui int
	abcEvaluation
}

func (eval *abcWithLai) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	ok, effects := eval.abcEvaluation.Eval(ctx, raceIdx, tiles, whoIdx, tile)
	if ok {
		//不能用含有癞子牌去吃
		for _, item := range effects {
			if item.Index(eval.tileLai) != -1 {
				return false, nil
			}
		}
		return true, effects
	}
	return false, nil
}

type fixWithLai struct {
	tile int
}

func (eval *fixWithLai) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	//只能自杠
	if raceIdx != whoIdx {
		return false, nil
	}
	//是否存在
	exist := tiles.Index(eval.tile)
	if exist != -1 {
		return false, nil
	}
	return true, []mj.Cards{{eval.tile}}
}

type winWithLai struct {
	tileLai int
	tileGui int
	noWin   bool //不能放铳
	unique  bool //唯一癞子？
	winEvaluation
}

// LaiCollect 只可能出现的癞子集
var LaiCollect = mj.Cards{
	mj.W1, mj.L1, mj.T1,
	mj.W2, mj.L2, mj.T2,
	mj.W3, mj.L3, mj.T3,
	mj.W4, mj.L4, mj.T4,
	mj.W5, mj.L5, mj.T5,
	mj.W6, mj.L6, mj.T6,
	mj.W7, mj.L7, mj.T7,
	mj.W8, mj.L8, mj.T8,
	mj.W9, mj.L9, mj.T9,
}

func (eval *winWithLai) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	// 只能自摸
	if eval.noWin && (raceIdx != whoIdx) {
		return false, nil
	}

	hands := tiles.Clone()
	//手牌中有鬼牌不允许胡
	if hands.Index(eval.tileGui) != -1 {
		return false, nil
	}

	//非自己手牌 合并目标牌后进行判定
	if raceIdx != whoIdx {
		hands = append(hands, tile)
	}

	//判断手上的癞子
	laiCount := len(hands.Indexes(eval.tileLai))
	//一个癞子才能胡
	if eval.unique && laiCount > 1 {
		return false, nil
	}

	if laiCount == 0 {
		//无癞子 按照标准牌型胡牌
		return eval.winEvaluation.Eval(ctx, raceIdx, tiles, whoIdx, tile)
	}
	//多癞子
	ok, effect := eval.multiLaiCheck(hands)
	if ok {
		//有效组合
		out := make([]mj.Cards, 0)
		out = append(out, effect.ABC...)
		out = append(out, effect.DDD...)
		out = append(out, effect.EE)
		return true, out
	}
	return false, nil
}

func (eval *winWithLai) multiLaiCheck(temp mj.Cards) (bool, *mj.WinComb) {

	//可选癞子

	combOps := &laiCombination{tiles: LaiCollect}
	//多个癞子胡牌算法
	winChecker := mj.NewWinChecker()
	//把癞子当作普通牌先检验，通过剩余的牌中的癞子再进行判断
	allCombs := winChecker.CheckAll(temp)
	for _, comb := range allCombs {

		//有效组合后，二次判断癞子不能做将
		if comb.OK {
			return comb.EE[0] != eval.tileLai, comb
		}
		//癞子作为普通牌判定后，多余的牌中可能会多出癞子
		laiIdxes := comb.Parts.Indexes(eval.tileLai)
		laiCount := len(laiIdxes)
		//无多余癞子，则表示剩下的其他牌是多余的
		if laiCount == 0 {
			return false, nil
		}
		//仅多一个，则判断是否可以组成将
		if laiCount == 1 && len(comb.Parts) == 2 {
			//update
			comb.EE = comb.Parts.Clone()
			//set empty
			comb.Parts = mj.Cards{}
			comb.OK = true
			return true, comb
		}

		//过滤癞子后的牌，再将癞子带入进行判断
		noLaiPart := comb.Parts.Remove(laiIdxes...)
		//组合
		laiCombs := combOps.product(laiCount)

		for i := 0; i < len(laiCombs); i++ {
			tempLaiComb := laiCombs[i]
			//新牌组
			nextPart := noLaiPart.Clone()
			nextPart = append(nextPart, tempLaiComb...)

			//二次判定
			winComb := winChecker.Check(nextPart)
			if winComb == nil || !winComb.OK {
				continue
			}
			//构建新的组合
			newComb := &mj.WinComb{
				OK:    true,
				ABC:   append(comb.ABC, winComb.ABC...),
				DDD:   append(comb.DDD, winComb.DDD...),
				EE:    append(comb.EE, winComb.EE...),
				Parts: make(mj.Cards, 0),
			}
			//还原癞子牌组
			return true, recoverLaiComb(newComb, tempLaiComb, eval.tileLai)
		}
	}
	return false, nil
}

func recoverLaiComb(targetComb *mj.WinComb, tempLaiComb []int, tile int) *mj.WinComb {

	abc := targetComb.ABC
	ddd := targetComb.DDD
	ee := targetComb.EE

	for i := 0; i < len(tempLaiComb); i++ {
		tempLai := tempLaiComb[i]

		hasLai := false
		for j := 0; j < len(abc); j++ {
			idx := abc[j].Index(tempLai)
			if idx != -1 {
				abc[j][idx] = tile
				hasLai = true
				break
			}
		}

		if hasLai {
			continue
		}

		for j := 0; j < len(ddd); j++ {
			idx := ddd[j].Index(tempLai)
			if idx != -1 {
				ddd[j][idx] = tile
				hasLai = true
				break
			}
		}

		if hasLai {
			continue
		}

		idx := ee.Index(tempLai)
		if idx != -1 {
			ee[idx] = tile
		}
	}
	return targetComb
}

type laiCombination struct {
	tiles []int
}

//癞子 全排列
func (comb *laiCombination) product(num int) [][]int {

	sets := make([][]int, 0)
	for i := 0; i < num; i++ {
		sets = append(sets, comb.tiles)
	}

	lens := func(i int) int { return len(sets[i]) }
	product := make([][]int, 0)
	for ix := make([]int, len(sets)); ix[0] < lens(0); comb.nextIndex(ix, lens) {
		r := make([]int, 0)
		for j, k := range ix {
			r = append(r, sets[j][k])
		}
		product = append(product, r)
	}
	return product
}

func (comb *laiCombination) nextIndex(ix []int, lens func(i int) int) {
	for j := len(ix) - 1; j >= 0; j-- {
		ix[j]++
		if j == 0 || ix[j] < lens(j) {
			return
		}
		ix[j] = 0
	}
}
