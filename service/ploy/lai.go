package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/service/engine"
)

// LaiProvider 癞子
type LaiProvider struct {
	BaseProvider
	variety string

	tileLai int
	tileCao int
	tileGui int

	hasGui bool //是否有鬼牌（红中）
	noABC  bool //不能吃
}

func newLaiProvider(variety string) GameDefine {
	bp := BaseProvider{}
	switch variety {
	case "lai-not":
		//无癞
		return &LaiProvider{
			variety:      variety,
			BaseProvider: bp,
			hasGui:       true,
			noABC:        false,
		}
	case "lai-unique":
		//一癞到底
		return &LaiProvider{
			variety:      variety,
			BaseProvider: bp,
			hasGui:       false,
			noABC:        true,
		}
	case "lai-multiple":
		//癞幌
		return &LaiProvider{
			variety:      variety,
			BaseProvider: bp,
			hasGui:       false,
			noABC:        false,
		}
	default:
		return nil
	}
}

type LaiRoundCtxHandler struct {
	*BaseRoundCtxHandler
	Cao int
	Lai int
	Gui int
}

func (lp *LaiProvider) Extras() []*mj.CardExtra {
	extras := []*mj.CardExtra{
		{Tile: lp.tileCao, Name: "朝"},
		{Tile: lp.tileLai, Name: "癞"},
	}
	if lp.hasGui {
		extras = append(extras, &mj.CardExtra{Tile: lp.tileGui, Name: "鬼"})
	}
	return extras
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
		//将牌放回末尾
		handler.table.Append(cao)
	}

	//从前摸张牌，当前牌为朝天，下一张为癞牌
	var lai int
	switch cao {
	case mj.W9:
		lai = mj.W1
		break
	case mj.T9:
		lai = mj.T1
		break
	case mj.L9:
		lai = mj.L1
		break
	default:
		lai = cao + 1
	}

	lp.tileGui = mj.Zh
	lp.tileCao = cao
	lp.tileLai = lai

	return &LaiRoundCtxHandler{
		BaseRoundCtxHandler: handler,
		Cao:                 cao,
		Lai:                 lai,
		Gui:                 mj.Zh,
	}
}

func (lp *LaiProvider) Handles() map[api.RaceType]RaceEvaluator {

	illegals := mj.Cards{lp.tileLai, lp.tileCao, lp.tileGui}

	evalMap := map[api.RaceType]RaceEvaluator{
		api.DDDRace:         &dddEvaluation{illegals: illegals},
		api.ABCRace:         &abcEvaluation{illegals: illegals},
		api.EEEERace:        &eeeeEvaluation{},
		api.EEEEOwnRace:     &eeeeOwnEvaluation{},
		api.EEEEUpgradeRace: &eeeeUpgradeEvaluation{},
		api.CaoRace:         &caoWithLai{tile: lp.tileCao},
		api.LaiRace:         &fixWithLai{tile: lp.tileLai},
	}

	//有红中
	if lp.hasGui {
		evalMap[api.GuiRace] = &fixWithLai{tile: lp.tileGui}
	}

	//不能吃
	if lp.noABC {
		delete(evalMap, api.ABCRace)
	}

	//胡牌策略
	laiChecker := multiLaiWinChecker{tileLai: lp.tileLai, tileGui: lp.tileGui}
	winChecker := winEvaluation{}
	switch lp.variety {
	case "lai-unique":
		evalMap[api.WinRace] = &winWithUniqueLai{multiLaiWinChecker: laiChecker, winEvaluation: winChecker}
		break
	case "lai-not":
		evalMap[api.WinRace] = &winWithNotLai{multiLaiWinChecker: laiChecker, winEvaluation: winChecker}
		break
	case "lai-multiple":
		evalMap[api.WinRace] = &winWithMultipleLai{multiLaiWinChecker: laiChecker, winEvaluation: winChecker}
		break
	}
	return evalMap
}

type fixWithLai struct {
	tile int
}

func (eval *fixWithLai) Valid(ctx *engine.RoundCtx, raceIdx int, racePart mj.Cards, whoIdx int, whoTile int) bool {
	return len(racePart) == 1
}

func (eval *fixWithLai) Next(ctx *engine.RoundCtx, raceIdx int, whoIdx int) RaceNext {
	return NextTake
}

func (eval *fixWithLai) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	//只能自杠
	if raceIdx != whoIdx {
		return false, nil
	}
	//是否存在
	exist := tiles.Index(eval.tile)
	if exist != -1 {
		return true, []mj.Cards{{eval.tile}}
	}
	return false, nil
}

type caoWithLai struct {
	tile int
}

func (eval *caoWithLai) Valid(ctx *engine.RoundCtx, raceIdx int, racePart mj.Cards, whoIdx int, whoTile int) bool {
	//自己回合出3张
	if raceIdx == whoIdx {
		return len(racePart) == 3
	}
	//他人回合出两张
	return len(racePart) == 2
}

func (eval *caoWithLai) Next(ctx *engine.RoundCtx, raceIdx int, whoIdx int) RaceNext {
	return NextPut
}

func (eval *caoWithLai) Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards) {
	//朝自己，或者别人
	if raceIdx == whoIdx {
		//自己回合时，判定有3张
		caos := ctx.Operating().GetTiles(raceIdx).Hands.Indexes(eval.tile)
		if len(caos) == 3 {
			return true, []mj.Cards{caos}
		}
	} else {
		//他人回合
		if eval.tile != tile {
			return false, nil
		}
		caos := ctx.Operating().GetTiles(raceIdx).Hands.Indexes(eval.tile)
		if len(caos) == 2 {
			return true, []mj.Cards{caos}
		}
	}
	return false, nil
}
