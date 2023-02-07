package server

import (
	"errors"
	"log"
	"mahjong/ploy"
	"mahjong/server/api"
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
	if !roundCtx.Position.Check(own.Idx) {
		return nil, errors.New("非当前回合")
	}

	//摸牌
	var takeTile int
	if body.Direction == -1 {
		takeTile = roundCtx.Handler.Backward(own.Idx)
	} else {
		takeTile = roundCtx.Handler.Forward(own.Idx)
	}

	//剩余牌
	takeRemained := roundCtx.Handler.Remained()
	takeResult := &api.TakeResult{Tile: takeTile, Remained: takeRemained}

	//保存摸到的牌
	roundCtx.Handler.AddTake(own.Idx, takeTile)
	//通知
	roundCtx.Exchanger.PostTake(&api.TakePayload{Who: own.Idx, Round: roundCtx.Round, Tile: takeTile, Remained: takeRemained})
	return takeResult, nil
}

//出牌
func put(w http.ResponseWriter, r *http.Request, body *api.PutParameter) (*api.NoResp, error) {
	header := wrap.GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return nil, err
	}
	//玩家信息
	own, _ := roundCtx.Player(header.UserId)
	//判定回合
	if !roundCtx.Position.Check(own.Idx) {
		return nil, errors.New("非当前回合")
	}
	//保存
	roundCtx.Handler.AddPut(own.Idx, body.Tile)
	//通知
	roundCtx.Exchanger.PostPut(&body.PutPayload)
	return api.Empty, nil
}

//吃碰杠...
func race(w http.ResponseWriter, r *http.Request, body *api.RaceParameter) (*api.RacePost, error) {

	header := wrap.GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return nil, err
	}
	//玩家信息
	own, _ := roundCtx.Player(header.UserId)

	//游戏策略 恢复状态
	gc, _ := roundCtx.Handler.WithConfig()
	var provider = ploy.NewProvider(gc.Mode)
	provider.Renew(roundCtx)
	eval, exist := provider.Handles()[body.RaceType]
	if !exist {
		return nil, errors.New("不支持当前操作")
	}
	//判定
	if ok, _ := eval.Eval(roundCtx, own.Idx, body.Who, body.Tile); !ok {
		return nil, errors.New("不支持牌型")
	}

	//通知
	roundCtx.Exchanger.PostRace(&body.RacePayload)

	//后置事件
	var nextAction *api.RacePost
	switch body.RaceType {
	case api.EEEERace, api.LaiRace, api.GuiRace:
		//杠，从后摸
		nextAction = &api.RacePost{Action: "take", Direction: -1}
		break
	case api.ABCRace:
		//吃，出牌
		nextAction = &api.RacePost{Action: "put", Direction: 0}
		break
	case api.DDDRace, api.CaoRace:
		//碰，出牌
		nextAction = &api.RacePost{Action: "put", Direction: 0}
		break
	default:
		nextAction = &api.RacePost{Action: "ignore", Direction: 0}
	}
	return nextAction, nil
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

	//判定
	gc, _ := roundCtx.Handler.WithConfig()
	var provider = ploy.NewProvider(gc.Mode)
	provider.Renew(roundCtx)
	handles := provider.Handles()

	//判定可用
	items := make([]*api.UsableRaceItem, 0)
	for k, v := range handles {
		ok, usable := v.Eval(roundCtx, own.Idx, body.Who, body.Tile)
		if !ok {
			continue
		}
		items = append(items, &api.UsableRaceItem{RaceType: k, Tiles: usable})
	}

	if own.Idx == body.Who {
		//自己回合
		items = append(items, &api.UsableRaceItem{RaceType: api.PutRace})
	} else {
		//他人回合
		if len(items) > 0 { //如果有可选项，则添加忽略操作
			items = append(items, &api.UsableRaceItem{RaceType: api.PassRace})
		} else {
			//无可选，直接回执忽略事件
			roundCtx.Exchanger.PostAck(&api.AckPayload{Who: own.Idx, Round: roundCtx.Round, AckId: body.AckId})
		}
	}
	itemNames := make([]string, 0)
	for _, i := range items {
		itemNames = append(itemNames, api.RaceNames[i.RaceType])
	}
	log.Printf("%s 判定：%v\n", header.UserId, itemNames)
	return &api.RaceEffects{Usable: items}, nil
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
	roundCtx.Exchanger.PostAck(&api.AckPayload{
		Who:   own.Idx,
		Round: roundCtx.Round,
		AckId: body.AckId,
	})
	return api.Empty, nil
}

//胡牌
func win(w http.ResponseWriter, r *http.Request, body api.RacePreview) (*api.NoResp, error) {
	return nil, nil
}
