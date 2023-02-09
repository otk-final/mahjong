package server

import (
	"errors"
	"mahjong/mj"
	"mahjong/ploy"
	"mahjong/server/api"
	"mahjong/server/engine"
	"mahjong/server/store"
	"mahjong/server/wrap"
	"net/http"
)

// 摸牌
func take(w http.ResponseWriter, r *http.Request, body *api.TakeParameter) (*api.TakeResult, error) {
	header := wrap.GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return nil, err
	}
	//玩家信息
	own, _ := roundCtx.Player(header.UserId)
	//判定回合
	if !roundCtx.Pos().Check(own.Idx) {
		return nil, errors.New("非当前回合")
	}

	//是否已经摸牌了
	recentOps := roundCtx.HandlerCtx().Recenter(own.Idx)
	if recentOps != nil && recentOps.Action() == engine.RecentTake {
		return nil, errors.New("不允许重复摸牌")
	}

	//摸牌
	takeResult := doTake(roundCtx, own, body)
	if takeResult.Take == -1 {
		return nil, errors.New("游戏结束 平局")
	}

	//判定
	options, err := doRacePre(roundCtx, own, &api.RacePreview{
		RoomId: body.RoomId,
		Round:  body.Round,
		AckId:  -1,
		Target: own.Idx,
		Tile:   takeResult.Take,
	})
	if err != nil {
		return nil, err
	}
	takeResult.Options = options
	return takeResult, nil
}

func doTake(roundCtx *engine.RoundCtx, own *api.Player, body *api.TakeParameter) *api.TakeResult {
	//摸牌
	var takeTile int
	if body.Direction == -1 {
		takeTile = roundCtx.HandlerCtx().Backward(own.Idx)
	} else {
		takeTile = roundCtx.HandlerCtx().Forward(own.Idx)
	}
	roundCtx.HandlerCtx().AddTake(own.Idx, takeTile)
	//剩余牌
	takeRemained := roundCtx.HandlerCtx().Remained()
	//通知
	roundCtx.Exchange().PostTake(&api.TakePayload{Who: own.Idx, Round: body.Round, Tile: 0, Remained: takeRemained})

	return &api.TakeResult{
		PlayerTiles: roundCtx.HandlerCtx().GetTiles(own.Idx),
		Take:        takeTile, Remained: takeRemained,
	}
}

//出牌
func put(w http.ResponseWriter, r *http.Request, body *api.PutParameter) (*api.PutResult, error) {
	header := wrap.GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return nil, err
	}
	//玩家信息
	own, _ := roundCtx.Player(header.UserId)
	//判定回合
	if !roundCtx.Pos().Check(own.Idx) {
		return nil, errors.New("非当前回合")
	}
	//是否已经打牌了
	recentOps := roundCtx.HandlerCtx().Recenter(own.Idx)
	if recentOps != nil && recentOps.Action() == engine.RecentPut {
		return nil, errors.New("不允许重复出牌")
	}

	//保存
	roundCtx.HandlerCtx().AddPut(own.Idx, body.Tile)
	//通知
	body.Who = own.Idx
	roundCtx.Exchange().PostPut(body.PutPayload)

	//最新手牌
	return &api.PutResult{
		PlayerTiles: roundCtx.HandlerCtx().GetTiles(own.Idx),
		Put:         body.Tile,
	}, nil
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

//吃碰杠...
func race(w http.ResponseWriter, r *http.Request, body *api.RaceParameter) (*api.RaceResult, error) {

	header := wrap.GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return nil, err
	}
	//玩家信息
	own, _ := roundCtx.Player(header.UserId)

	//加锁防止并发操作
	if !roundCtx.Lock.TryLock() {
		return nil, errors.New("并发错误")
	}
	defer roundCtx.Lock.Unlock()

	//游戏策略
	var provider = ploy.RenewProvider(roundCtx)
	eval, exist := provider.Handles()[body.RaceType]
	if !exist {
		return nil, errors.New("不支持当前操作")
	}

	//目标牌
	recentIdx := roundCtx.HandlerCtx().RecentIdx()
	recenter := roundCtx.HandlerCtx().Recenter(recentIdx)
	targetTile := -1
	if recentIdx == own.Idx {
		targetTile = recenter.Take()
	} else {
		targetTile = recenter.Put()
	}

	//判定
	hands := roundCtx.HandlerCtx().GetTiles(own.Idx).Hands.Clone()
	if ok, plans := eval.Eval(roundCtx, own.Idx, hands, recentIdx, targetTile); !ok || !matchRacePlan(body.Tiles, plans) {
		return nil, errors.New("不支持牌型")
	}

	//保存
	roundCtx.HandlerCtx().AddRace(own.Idx, body.RaceType, &engine.TileRaces{Tiles: body.Tiles.Clone(), TargetIdx: recentIdx, Tile: targetTile})

	//通知
	roundCtx.Exchange().PostRace(&api.RacePayload{
		RaceType: body.RaceType,
		Who:      own.Idx,
		Target:   recentIdx,
		Round:    body.Round,
		Tiles:    ploy.RaceTilesMerge(body.RaceType, body.Tiles, targetTile),
		Tile:     targetTile,
		Interval: turnInterval,
	})

	//后置事件
	var usableRaces []*api.RaceOption
	var continueTake int
	switch body.RaceType {
	case api.EEEERace, api.LaiRace, api.GuiRace:
		//从后往前摸牌
		takeResult := doTake(roundCtx, own, &api.TakeParameter{RoomId: body.RoomId, Round: body.Round, Direction: -1})
		if takeResult.Take == -1 {
			return nil, errors.New("游戏结束 平局")
		}
		continueTake = takeResult.Take
		//判定
		usableRaces, err = doRacePre(roundCtx, own, &api.RacePreview{
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
		PlayerTiles:  roundCtx.HandlerCtx().GetTiles(own.Idx),
		ContinueTake: continueTake,
		Options:      usableRaces,
		Target:       recentIdx,
		TargetTile:   targetTile,
	}, nil
}

//吃碰杠...预览
func racePre(w http.ResponseWriter, r *http.Request, body *api.RacePreview) (*api.RaceEffects, error) {
	header := wrap.GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return nil, err
	}
	own, _ := roundCtx.Player(header.UserId)

	//目标牌
	recentIdx := roundCtx.HandlerCtx().RecentIdx()
	recenter := roundCtx.HandlerCtx().Recenter(recentIdx)
	targetTile := -1
	if recentIdx == own.Idx {
		targetTile = recenter.Take()
	} else {
		targetTile = recenter.Put()
	}

	//取内存数据
	body.Target = recentIdx
	body.Tile = targetTile
	body.AckId = roundCtx.Exchange().CurrentAckId()

	//可用判定查询
	items, err := doRacePre(roundCtx, own, body)
	if err != nil {
		return nil, err
	}
	return &api.RaceEffects{Options: items}, nil
}

func doRacePre(roundCtx *engine.RoundCtx, own *api.Player, body *api.RacePreview) ([]*api.RaceOption, error) {

	//策略集
	var handles = ploy.RenewProvider(roundCtx).Handles()

	//判定可用
	items := make([]*api.RaceOption, 0)
	hands := roundCtx.HandlerCtx().GetTiles(own.Idx).Hands
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

//过
func ignore(w http.ResponseWriter, r *http.Request, body *api.AckParameter) (*api.NoResp, error) {
	header := wrap.GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return nil, err
	}
	//玩家信息
	own, _ := roundCtx.Player(header.UserId)
	//通知
	roundCtx.Exchange().PostAck(&api.AckPayload{
		Who:   own.Idx,
		Round: body.Round,
		AckId: roundCtx.Exchange().CurrentAckId(),
	})
	return api.Empty, nil
}

//胡牌
func win(w http.ResponseWriter, r *http.Request, body api.RacePreview) (*api.NoResp, error) {
	return nil, nil
}
