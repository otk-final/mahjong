package service

import (
	"errors"
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/service/engine"
	"mahjong/service/ploy"
)

func DoTake(roundCtx *engine.RoundCtx, own *api.Player, body *api.TakeParameter) *api.TakeResult {
	ops := roundCtx.Operating()
	//摸牌
	var takeTile int
	if body.Direction == -1 {
		takeTile = ops.Backward(own.Idx)
	} else {
		takeTile = ops.Forward(own.Idx)
	}

	//剩余牌
	takeRemained := ops.Remained()
	options := make([]*api.RaceOption, 0)

	//游戏结束
	if takeTile == -1 {
		roundCtx.Exchange().Quit("平局")
	} else {
		//保存
		ops.AddTake(own.Idx, takeTile)
		//通知 屏蔽真实牌
		roundCtx.Exchange().PostTake(&api.TakePayload{Who: own.Idx, Tile: 0, Remained: takeRemained})
		//判定
		options = DoRacePre(roundCtx, own, &api.RacePreview{
			RoomId: body.RoomId,
			Target: own.Idx,
			Tile:   takeTile,
		})
	}

	return &api.TakeResult{PlayerTiles: ops.GetTiles(own.Idx), Take: takeTile, Remained: takeRemained, Options: options}
}

func DoPut(roundCtx *engine.RoundCtx, own *api.Player, body *api.PutParameter) *api.PutResult {
	ops := roundCtx.Operating()
	//保存
	ops.AddPut(own.Idx, body.Tile)
	//通知
	body.Who = own.Idx
	roundCtx.Exchange().PostPut(body.PutPayload)
	//最新手牌
	return &api.PutResult{PlayerTiles: ops.GetTiles(own.Idx), Put: body.Tile}
}

func matchRacePlan(target mj.Cards, plans []mj.Cards) bool {
	if len(plans) == 0 {
		return false
	}
	//满足一项即可
	for _, p := range plans {
		if p.Equal(target) {
			return true
		}
	}
	return false
}

func DoRace(roundCtx *engine.RoundCtx, own *api.Player, body *api.RaceParameter) (*api.RaceResult, error) {
	//游戏策略
	var provider = ploy.RenewProvider(roundCtx)
	eval, exist := provider.Handles()[body.RaceType]
	if !exist {
		return nil, errors.New("不支持当前操作")
	}
	ops := roundCtx.Operating()

	//目标牌
	recentIdx := ops.RecentIdx()
	recenter := ops.Recenter(recentIdx)
	targetTile := -1
	if recentIdx == own.Idx {
		targetTile = recenter.Take()
	} else {
		targetTile = recenter.Put()
	}

	//判断出牌数量是否合理
	racePart := body.Tiles
	if !eval.Valid(roundCtx, own.Idx, racePart, recentIdx, targetTile) {
		return nil, errors.New("出牌数量不符")
	}

	//判定
	hands := ops.GetTiles(own.Idx).Hands.Clone()
	if ok, plans := eval.Eval(roundCtx, own.Idx, hands, recentIdx, targetTile); !ok || !matchRacePlan(racePart, plans) {
		return nil, errors.New("不支持牌型")
	}

	//保存
	raceIntact := ops.AddRace(own.Idx, body.RaceType, &engine.TileRaces{Tiles: racePart, TargetIdx: recentIdx, Tile: targetTile})

	//通知
	roundCtx.Exchange().PostRace(&api.RacePayload{
		RaceType: body.RaceType,
		Who:      own.Idx,
		Target:   recentIdx,
		Tiles:    raceIntact,
		Tile:     targetTile,
		Interval: api.TurnInterval,
	})
	//后置事件
	next := eval.Next(roundCtx, own.Idx, recentIdx)

	//后置事件
	var options []*api.RaceOption
	var continueTake int
	switch next {
	case ploy.NextTake:
		//从后往前摸牌
		takeResult := DoTake(roundCtx, own, &api.TakeParameter{RoomId: body.RoomId, Direction: -1})
		if takeResult.Take == -1 {
			return nil, errors.New("游戏结束 平局")
		}
		continueTake = takeResult.Take
		options = takeResult.Options
		break
	case ploy.NextPut:
		//继续本回合内判定
		continueTake = -1
		options = DoRacePre(roundCtx, own, &api.RacePreview{RoomId: body.RoomId, Target: own.Idx, Tile: targetTile})
		break
	default:
		return nil, errors.New("后置事件非法")
	}

	//最新持牌
	return &api.RaceResult{
		PlayerTiles:  ops.GetTiles(own.Idx),
		ContinueTake: continueTake,
		Options:      options,
		Target:       recentIdx,
		TargetTile:   targetTile,
	}, nil
}

func DoRacePre(roundCtx *engine.RoundCtx, own *api.Player, body *api.RacePreview) []*api.RaceOption {

	//策略集
	var handles = ploy.RenewProvider(roundCtx).Handles()
	ops := roundCtx.Operating()
	//判定可用
	items := make([]*api.RaceOption, 0)
	hands := ops.GetTiles(own.Idx).Hands
	for k, v := range handles {
		ok, plans := v.Eval(roundCtx, own.Idx, hands.Clone(), body.Target, body.Tile)
		if !ok {
			continue
		}
		items = append(items, &api.RaceOption{RaceType: k, Tiles: plans})
	}

	if own.Idx == body.Target {
		//自己回合
		items = append(items, &api.RaceOption{RaceType: api.PutRace, Tiles: []mj.Cards{}})
	} else {
		//他人回合
		if len(items) > 0 {
			//如果有可选项，则添加忽略操作
			items = append(items, &api.RaceOption{RaceType: api.PassRace, Tiles: []mj.Cards{}})
		} else {
			//无可选，直接回执忽略事件
			DoIgnore(roundCtx, own)
		}
	}
	return items
}

func DoIgnore(roundCtx *engine.RoundCtx, own *api.Player) {
	//通知
	roundCtx.Exchange().PostAck(&api.AckPayload{Who: own.Idx, AckId: roundCtx.Exchange().CurrentAckId()})
}

func DoWin(roundCtx *engine.RoundCtx, own *api.Player) (*api.WinResult, error) {

	//游戏策略
	var provider = ploy.RenewProvider(roundCtx)
	winEval, exist := provider.Handles()[api.WinRace]
	if !exist {
		return nil, errors.New("不支持当前操作")
	}
	ops := roundCtx.Operating()

	//目标牌
	recentIdx := ops.RecentIdx()
	recenter := ops.Recenter(recentIdx)
	targetTile := -1
	if recentIdx == own.Idx {
		targetTile = recenter.Take()
	} else {
		targetTile = recenter.Put()
	}

	//判定
	hands := ops.GetTiles(own.Idx).Hands.Clone()
	if ok, _ := winEval.Eval(roundCtx, own.Idx, hands, recentIdx, targetTile); !ok {
		return nil, errors.New("不支持胡牌")
	}

	//TODO 根据上下文重新定义胡牌类型
	effectType := api.WinRace

	//通知
	winPayload := &api.WinPayload{
		Who:    own.Idx,
		Tiles:  ops.GetTiles(own.Idx),
		Target: recentIdx,
		Tile:   targetTile,
		Effect: effectType,
	}
	roundCtx.Exchange().PostWin(winPayload)
	return &api.WinResult{WinPayload: winPayload}, nil
}
