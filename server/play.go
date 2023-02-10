package server

import (
	"errors"
	"mahjong/server/api"
	"mahjong/server/wrap"
	"mahjong/service"
	"mahjong/service/engine"
	"mahjong/service/ploy"
	"mahjong/service/store"
	"net/http"
)

//  摸牌
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
	ops := roundCtx.Operating()

	//是否已经摸牌了
	recentOps := ops.Recenter(own.Idx)
	if recentOps != nil && recentOps.Action() == engine.RecentTake {
		return nil, errors.New("不允许重复摸牌")
	}

	//摸牌
	takeResult := service.DoTake(roundCtx, own, body)
	if takeResult.Take == -1 {
		return nil, errors.New("游戏结束 平局")
	}

	//判定
	takeResult.Options = service.DoRacePre(roundCtx, own, &api.RacePreview{
		RoomId: body.RoomId,
		Target: own.Idx,
		Tile:   takeResult.Take,
	})
	return takeResult, nil
}

// 出牌
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
	ops := roundCtx.Operating()
	//是否已经打牌了
	recentOps := ops.Recenter(own.Idx)
	if recentOps != nil && recentOps.Action() == engine.RecentPut {
		return nil, errors.New("不允许重复出牌")
	}

	if !ploy.RenewProvider(roundCtx).CanPut(own.Idx, body.Tile) {
		return nil, errors.New("不允许单独出牌")
	}

	return service.DoPut(roundCtx, own, body), nil
}

//  吃碰杠...
func race(w http.ResponseWriter, r *http.Request, body *api.RaceParameter) (*api.RaceResult, error) {

	header := wrap.GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return nil, err
	}
	//玩家信息
	own, _ := roundCtx.Player(header.UserId)

	//加锁防止并发操作 互斥
	if !roundCtx.Lock.TryLock() {
		return nil, errors.New("并发错误")
	}
	defer roundCtx.Lock.Unlock()

	return service.DoRace(roundCtx, own, body)
}

//  吃碰杠...预览
func racePre(w http.ResponseWriter, r *http.Request, body *api.RacePreview) (*api.RaceEffects, error) {
	header := wrap.GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return nil, err
	}
	own, _ := roundCtx.Player(header.UserId)
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

	//取内存数据
	body.Target = recentIdx
	body.Tile = targetTile

	//可用判定查询
	items := service.DoRacePre(roundCtx, own, body)
	return &api.RaceEffects{Options: items}, nil
}

//  过
func ignore(w http.ResponseWriter, r *http.Request, body *api.AckParameter) (*api.NoResp, error) {
	header := wrap.GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return nil, err
	}
	own, _ := roundCtx.Player(header.UserId)
	//忽略
	service.DoIgnore(roundCtx, own)
	return api.Empty, nil
}

//  胡牌
func win(w http.ResponseWriter, r *http.Request, body *api.WinParameter) (*api.WinResult, error) {

	header := wrap.GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return nil, err
	}
	//玩家信息
	own, _ := roundCtx.Player(header.UserId)

	//加锁防止并发操作 等待
	defer roundCtx.Lock.Unlock()
	roundCtx.Lock.Lock()

	return service.DoWin(roundCtx, own)
}
