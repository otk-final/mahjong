package server

import (
	"mahjong/server/api"
	"mahjong/server/store"
	"mahjong/server/wrap"
	"net/http"
)

// 摸牌
func take(w http.ResponseWriter, r *http.Request, body *api.TakeParameter) (*api.NoResp, error) {

	header := wrap.GetHeader(r)

	//摸牌
	if body.Direction == -1 {
		//从后摸

	} else {
		//从前摸

	}
	takeTile := 1

	//计时器
	cd, err := store.GetCountdown(body.RoomId)
	if err != nil {
		return nil, err
	}

	//通知处理
	err = cd.ToTake(&api.TakePayload{Who: 0, Round: 9, Take: takeTile})
	if err != nil {
		return nil, err
	}

	return api.Empty, nil
}

//出牌
func put(w http.ResponseWriter, r *http.Request, body *api.PutParameter) (*api.NoResp, error) {
	return nil, nil
}

//吃碰杠...
func race(w http.ResponseWriter, r *http.Request, body *api.RaceParameter) (*api.RacePost, error) {

	//处理

	//通知
	cd, err := store.GetCountdown(body.RoomId)
	if err != nil {
		return nil, err
	}
	cd.ToRace(nil)

	//后置事件
	var nextAction *api.RacePost
	switch body.RaceType {
	case api.GangRace: //杠，从后摸
		nextAction = &api.RacePost{Action: "take", Direction: -1}
		break
	case api.EatRace: //吃，出牌
		nextAction = &api.RacePost{Action: "put", Direction: 0}
		break
	case api.PairRace: //碰，出牌
		nextAction = &api.RacePost{Action: "put", Direction: 0}
		break
	default:
		nextAction = &api.RacePost{Action: "skip", Direction: 0}
	}
	return nextAction, nil
}

//吃碰杠...预览
func racePre(w http.ResponseWriter, r *http.Request, body *api.RacePreview) (*api.RaceEffects, error) {
	return nil, nil
}

//过
func skip(w http.ResponseWriter, r *http.Request, body *api.AckParameter) (*api.NoResp, error) {
	return nil, nil
}

//胡牌
func win(w http.ResponseWriter, r *http.Request, body api.RacePreview) (*api.NoResp, error) {
	return nil, nil
}
