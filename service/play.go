package service

import (
	"errors"
	"mahjong/mj"
	"mahjong/ploy"
	"mahjong/server/api"
	"mahjong/service/engine"
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
	ops.AddTake(own.Idx, takeTile)
	//剩余牌
	takeRemained := ops.Remained()
	//通知
	roundCtx.Exchange().PostTake(&api.TakePayload{Who: own.Idx, Round: body.Round, Tile: 0, Remained: takeRemained})

	return &api.TakeResult{
		PlayerTiles: ops.GetTiles(own.Idx),
		Take:        takeTile, Remained: takeRemained,
	}
}

func DoPut(roundCtx *engine.RoundCtx, own *api.Player, body *api.PutParameter) *api.PutResult {
	ops := roundCtx.Operating()
	//保存
	ops.AddPut(own.Idx, body.Tile)
	//通知
	body.Who = own.Idx
	roundCtx.Exchange().PostPut(body.PutPayload)
	//最新手牌
	return &api.PutResult{
		PlayerTiles: ops.GetTiles(own.Idx),
		Put:         body.Tile,
	}
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

	//判定
	hands := ops.GetTiles(own.Idx).Hands.Clone()
	if ok, plans := eval.Eval(roundCtx, own.Idx, hands, recentIdx, targetTile); !ok || !matchRacePlan(body.Tiles, plans) {
		return nil, errors.New("不支持牌型")
	}
	//保存
	ops.AddRace(own.Idx, body.RaceType, &engine.TileRaces{Tiles: body.Tiles.Clone(), TargetIdx: recentIdx, Tile: targetTile})

	//通知
	roundCtx.Exchange().PostRace(&api.RacePayload{
		RaceType: body.RaceType,
		Who:      own.Idx,
		Target:   recentIdx,
		Round:    body.Round,
		Tiles:    ploy.RaceTilesMerge(body.RaceType, body.Tiles, targetTile),
		Tile:     targetTile,
		Interval: api.TurnInterval,
	})

	//后置事件
	var usableRaces []*api.RaceOption
	var err error
	var continueTake int
	switch body.RaceType {
	case api.EEEERace, api.LaiRace, api.GuiRace:
		//从后往前摸牌
		takeResult := DoTake(roundCtx, own, &api.TakeParameter{RoomId: body.RoomId, Round: body.Round, Direction: -1})
		if takeResult.Take == -1 {
			return nil, errors.New("游戏结束 平局")
		}
		continueTake = takeResult.Take
		//判定
		usableRaces, err = DoRacePre(roundCtx, own, &api.RacePreview{
			RoomId: body.RoomId,
			Round:  body.Round,
			AckId:  -1,
			Target: own.Idx,
			Tile:   continueTake,
		})
		if err != nil {
			return nil, err
		}
		break
	case api.ABCRace, api.DDDRace, api.CaoRace:
		continueTake = -1
		//吃，碰 , 朝 渲染出牌入口
		usableRaces = append(usableRaces, &api.RaceOption{RaceType: api.PutRace, Tiles: nil})
		break
	default:
		return nil, errors.New("事件非法")
	}
	//最新持牌
	return &api.RaceResult{
		PlayerTiles:  ops.GetTiles(own.Idx),
		ContinueTake: continueTake,
		Options:      usableRaces,
		Target:       recentIdx,
		TargetTile:   targetTile,
	}, nil
}

func DoRacePre(roundCtx *engine.RoundCtx, own *api.Player, body *api.RacePreview) ([]*api.RaceOption, error) {

	//策略集
	var handles = ploy.RenewProvider(roundCtx).Handles()
	ops := roundCtx.Operating()
	//判定可用
	items := make([]*api.RaceOption, 0)
	hands := ops.GetTiles(own.Idx).Hands
	for k, v := range handles {
		ok, usable := v.Eval(roundCtx, own.Idx, hands.Clone(), body.Target, body.Tile)
		if !ok {
			continue
		}
		items = append(items, &api.RaceOption{RaceType: k, Tiles: usable})
	}

	if own.Idx == body.Target {
		//自己回合
		items = append(items, &api.RaceOption{RaceType: api.PutRace})
	} else {
		//他人回合
		if len(items) > 0 { //如果有可选项，则添加忽略操作
			items = append(items, &api.RaceOption{RaceType: api.PassRace})
		} else {
			//无可选，直接回执忽略事件
			roundCtx.Exchange().PostAck(&api.AckPayload{Who: own.Idx, Round: body.Round, AckId: body.AckId})
		}
	}
	return items, nil
}

func DoIgnore(roundCtx *engine.RoundCtx, own *api.Player, body *api.AckParameter) (*api.NoResp, error) {
	//通知
	roundCtx.Exchange().PostAck(&api.AckPayload{
		Who:   own.Idx,
		Round: body.Round,
		AckId: roundCtx.Exchange().CurrentAckId(),
	})
	return api.Empty, nil
}

func DoWin(roundCtx *engine.RoundCtx, own *api.Player, body *api.WinParameter) (*api.WinResult, error) {

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
		Round:  body.Round,
		Tiles:  ops.GetTiles(own.Idx),
		Target: recentIdx,
		Tile:   targetTile,
		Effect: effectType,
	}
	roundCtx.Exchange().PostWin(winPayload)
	return &api.WinResult{WinPayload: winPayload}, nil
}