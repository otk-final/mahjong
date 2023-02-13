package ploy

import (
	"mahjong/mj"
	"mahjong/service/engine"
)

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

//可选癞子
var combCollect1, combCollect2, combCollect3, combCollect4 [][]int

func init() {
	var combOps = &multiLaiCombination{tiles: LaiCollect}
	combCollect1 = combOps.product(1)
	combCollect2 = combOps.product(2)
	combCollect3 = combOps.product(3)
	combCollect4 = combOps.product(4)
}

//一癞到底，手上最多只能有一个癞子，能点炮
type winWithUniqueLai struct {
	winEvaluation
	multiLaiWinChecker
}

func (eval *winWithUniqueLai) Eval(ctx *engine.RoundCtx, raceIdx int, hands mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {

	laiCount := len(hands.Indexes(eval.tileLai))
	if laiCount > 1 || hands.Index(eval.tileGui) != -1 {
		return false, nil
	}

	//按照标准胡牌判定
	ok, plans := eval.winEvaluation.Eval(ctx, raceIdx, hands, whoIdx, tile)
	if ok {
		return true, plans
	}

	//不能胡牌，则判断是否有杠
	existRaces := ctx.Operating().GetTiles(raceIdx).Races
	if eval.hasLaiRace(existRaces) {

		//合并牌
		var winIntact mj.Cards
		if raceIdx == whoIdx { //自摸
			winIntact = hands
		} else { //别人点炮
			winIntact = append(hands, tile)
		}

		ok, effect := eval.winCheck(winIntact)
		if ok {
			//有效组合
			out := make([]mj.Cards, 0)
			out = append(out, effect.ABC...)
			out = append(out, effect.DDD...)
			out = append(out, effect.EE)
			return true, out
		}
	}
	return false, nil
}

//一脚癞油（无癞），手上不能有癞子，不能点炮
type winWithNotLai struct {
	winEvaluation
	multiLaiWinChecker
}

func (eval *winWithNotLai) Eval(ctx *engine.RoundCtx, raceIdx int, hands mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {

	//只能自摸
	if raceIdx != whoIdx {
		return false, nil
	}

	//无癞子
	laiCount := len(hands.Indexes(eval.tileLai))
	if laiCount > 0 || hands.Index(eval.tileGui) != -1 {
		return false, nil
	}

	//按照标准胡牌判定
	ok, plans := eval.winEvaluation.Eval(ctx, raceIdx, hands, whoIdx, tile)
	if ok {
		return true, plans
	}
	return false, nil
}

//多个癞子，均可以胡牌 能点炮
type winWithMultipleLai struct {
	winEvaluation
	multiLaiWinChecker
}

func (eval *winWithMultipleLai) Eval(ctx *engine.RoundCtx, raceIdx int, hands mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {

	//手牌中有鬼牌不允许胡
	if hands.Index(eval.tileGui) != -1 {
		return false, nil
	}

	//标准
	laiCount := len(hands.Indexes(eval.tileLai))
	if laiCount == 0 {
		return eval.winEvaluation.Eval(ctx, raceIdx, hands, whoIdx, tile)
	}

	//合并牌
	var winIntact mj.Cards
	if raceIdx == whoIdx { //自摸
		winIntact = hands
	} else { //别人点炮
		winIntact = append(hands, tile)
	}

	ok, effect := eval.winCheck(winIntact)
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

type multiLaiWinChecker struct {
	tileLai int
	tileGui int
}

func (wc *multiLaiWinChecker) winCheck(temp mj.Cards) (bool, *mj.WinComb) {

	//多个癞子胡牌算法
	winChecker := mj.NewWinChecker()
	//把癞子当作普通牌先检验，通过剩余的牌中的癞子再进行判断
	allCombs := winChecker.CheckAll(temp)
	for _, comb := range allCombs {

		//有效组合后，二次判断癞子不能做将
		if comb.OK {
			return comb.EE[0] != wc.tileLai, comb
		}
		//癞子作为普通牌判定后，多余的牌中可能会多出癞子
		laiIdxes := comb.Parts.Indexes(wc.tileLai)
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

		var laiCombs [][]int
		//组合
		switch laiCount {
		case 1:
			laiCombs = combCollect1
			break
		case 2:
			laiCombs = combCollect2
			break
		case 3:
			laiCombs = combCollect3
			break
		case 4:
			laiCombs = combCollect4
			break
		default:
			return false, nil
		}

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
			return true, recoverLaiComb(newComb, tempLaiComb, wc.tileLai)
		}
	}
	return false, nil
}

func (wc *multiLaiWinChecker) hasLaiRace(races []mj.Cards) bool {
	//不能胡牌，则判断是否有杠
	flag := false
	for i := 0; i < len(races); i++ {
		if races[i].IsLai(wc.tileLai) {
			flag = true
			break
		}
	}
	return flag
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

type multiLaiCombination struct {
	tiles []int
}

//癞子 全排列
func (comb *multiLaiCombination) product(num int) [][]int {

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

func (comb *multiLaiCombination) nextIndex(ix []int, lens func(i int) int) {
	for j := len(ix) - 1; j >= 0; j-- {
		ix[j]++
		if j == 0 || ix[j] < lens(j) {
			return
		}
		ix[j] = 0
	}
}
