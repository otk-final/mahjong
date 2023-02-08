package server

import (
	"errors"
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
	if takeResult.Tile == -1 {
		return nil, errors.New("游戏结束 平局")
	}

	//判定
	usableRaces, err := doRacePre(roundCtx, own, &api.RacePreview{
		RoomId: body.RoomId,
		Round:  body.Round,
		AckId:  -1,
		Target: own.Idx,
		Tile:   takeResult.Tile,
	})
	if err != nil {
		return nil, err
	}
	takeResult.Usable = usableRaces
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

	return &api.TakeResult{Tile: takeTile, Remained: takeRemained}
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
	roundCtx.Exchange().PostPut(&body.PutPayload)

	//最新手牌
	return &api.PutResult{
		RoomId: body.RoomId,
		Tile:   body.Tile,
		Hands:  roundCtx.HandlerCtx().GetTiles(own.Idx).Hands,
	}, nil
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

	//游戏策略
	var provider = ploy.BuildProvider(roundCtx)
	eval, exist := provider.Handles()[body.RaceType]
	if !exist {
		return nil, errors.New("不支持当前操作")
	}

	//是否已经打牌了
	recentOps := roundCtx.HandlerCtx().Recenter(own.Idx)
	if recentOps != nil && recentOps.Action() == engine.RecentRace {
		return nil, errors.New("不允许重复出牌")
	}

	//判定
	if ok, _ := eval.Eval(roundCtx, own.Idx, body.Target, body.Tile); !ok {
		return nil, errors.New("不支持牌型")
	}

	//保存
	roundCtx.HandlerCtx().AddRace(own.Idx, &engine.TileRaces{Tiles: body.Tiles, TargetIdx: body.Target, Tile: body.Tile})

	//通知
	body.Who = own.Idx
	roundCtx.Exchange().PostRace(&body.RacePayload)

	//后置事件
	var usableRaces []*api.UsableRaceItem
	switch body.RaceType {
	case api.EEEERace, api.LaiRace, api.GuiRace:
		//从后往前摸牌
		takeResult := doTake(roundCtx, own, &api.TakeParameter{RoomId: body.RoomId, Round: body.Round, Direction: -1})
		if takeResult.Tile == -1 {
			return nil, errors.New("游戏结束 平局")
		}
		//判定
		usableRaces, err = doRacePre(roundCtx, own, &api.RacePreview{
			RoomId: body.RoomId,
			Round:  body.Round,
			AckId:  -1,
			Target: own.Idx,
			Tile:   takeResult.Tile,
		})
		if err != nil {
			return nil, err
		}
		break
	case api.ABCRace, api.DDDRace, api.CaoRace:
		//吃，碰 , 朝 渲染出牌入口
		usableRaces = append(usableRaces, &api.UsableRaceItem{RaceType: 0, Tiles: nil})
		break
	default:
		return nil, errors.New("事件非法")
	}

	//最新持牌
	return &api.RaceResult{
		Hands:  roundCtx.HandlerCtx().GetTiles(own.Idx).Hands,
		Tiles:  body.Tiles,
		Tile:   body.Tile,
		Usable: usableRaces,
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

	//可用判定查询
	items, err := doRacePre(roundCtx, own, body)
	if err != nil {
		return nil, err
	}
	return &api.RaceEffects{Usable: items}, nil
}

func doRacePre(roundCtx *engine.RoundCtx, own *api.Player, body *api.RacePreview) ([]*api.UsableRaceItem, error) {

	//策略
	var provider = ploy.BuildProvider(roundCtx)
	handles := provider.Handles()

	//判定可用
	items := make([]*api.UsableRaceItem, 0)
	for k, v := range handles {
		ok, usable := v.Eval(roundCtx, own.Idx, body.Target, body.Tile)
		if !ok {
			continue
		}
		items = append(items, &api.UsableRaceItem{RaceType: k, Tiles: usable})
	}

	if own.Idx == body.Target {
		//自己回合
		items = append(items, &api.UsableRaceItem{RaceType: api.PutRace})
	} else {
		//他人回合
		if len(items) > 0 { //如果有可选项，则添加忽略操作
			items = append(items, &api.UsableRaceItem{RaceType: api.PassRace})
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
		AckId: body.AckId,
	})
	return api.Empty, nil
}

//胡牌
func win(w http.ResponseWriter, r *http.Request, body api.RacePreview) (*api.NoResp, error) {
	return nil, nil
}
