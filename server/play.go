package server

import (
	"errors"
	"github.com/otk-final/thf/resp"
	"mahjong/server/api"
	"mahjong/service"
	"mahjong/service/engine"
	"mahjong/service/ploy"
	"mahjong/service/store"
	"net/http"
)

//  摸牌
func take(w http.ResponseWriter, r *http.Request, body *api.TakeParameter) *resp.Entry[*api.TakeResult] {
	header := GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return resp.NewError[*api.TakeResult](err)
	}
	//玩家信息
	own, _ := roundCtx.Player(header.UserId)
	//判定回合
	if !roundCtx.Pos().Check(own.Idx) {
		return resp.NewError[*api.TakeResult](errors.New("非当前回合"))
	}
	ops := roundCtx.Operating()

	//是否已经摸牌了
	recentOps := ops.Recenter(own.Idx)
	if recentOps != nil && recentOps.Action() == engine.RecentTake {
		return resp.NewError[*api.TakeResult](errors.New("不允许重复摸牌"))
	}

	//摸牌
	takeResult := service.DoTake(roundCtx, own, body)
	if takeResult.Take == -1 {
		return resp.NewError[*api.TakeResult](errors.New("游戏结束 平局"))
	}

	//判定
	takeResult.Options = service.DoRacePre(roundCtx, own, &api.RacePreview{
		RoomId: body.RoomId,
		Target: own.Idx,
		Tile:   takeResult.Take,
	})
	return resp.NewEntry(takeResult)
}

// 出牌
func put(w http.ResponseWriter, r *http.Request, body *api.PutParameter) *resp.Entry[*api.PutResult] {
	header := GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return resp.NewError[*api.PutResult](err)
	}
	//玩家信息
	own, _ := roundCtx.Player(header.UserId)
	//判定回合
	if !roundCtx.Pos().Check(own.Idx) {
		return resp.NewError[*api.PutResult](errors.New("非当前回合"))
	}
	ops := roundCtx.Operating()
	//是否已经打牌了
	recentOps := ops.Recenter(own.Idx)
	if recentOps != nil && recentOps.Action() == engine.RecentPut {
		return resp.NewError[*api.PutResult](errors.New("不允许重复出牌"))
	}

	if !ploy.RenewProvider(roundCtx).CanPut(own.Idx, body.Tile) {
		return resp.NewError[*api.PutResult](errors.New("不允许单独出牌"))
	}

	putResult := service.DoPut(roundCtx, own, body)
	return resp.NewEntry(putResult)
}

//  吃碰杠...
func race(w http.ResponseWriter, r *http.Request, body *api.RaceParameter) *resp.Entry[*api.RaceResult] {

	header := GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return resp.NewError[*api.RaceResult](err)
	}
	//玩家信息
	own, _ := roundCtx.Player(header.UserId)

	//加锁防止并发操作 互斥
	if !roundCtx.Lock.TryLock() {
		return resp.NewError[*api.RaceResult](errors.New("并发错误"))
	}
	defer roundCtx.Lock.Unlock()

	raceResult, err := service.DoRace(roundCtx, own, body)
	if err != nil {
		return resp.NewError[*api.RaceResult](err)
	}
	return resp.NewEntry(raceResult)
}

//  吃碰杠...预览
func racePre(w http.ResponseWriter, r *http.Request, body *api.RacePreview) *resp.Entry[*api.RaceEffects] {
	header := GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return resp.NewError[*api.RaceEffects](err)
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
	return resp.NewEntry(&api.RaceEffects{Options: items})
}

//  过
func ignore(w http.ResponseWriter, r *http.Request, body *api.AckParameter) *resp.Entry[any] {
	header := GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return resp.NewError[any](err)
	}
	own, _ := roundCtx.Player(header.UserId)
	//忽略
	service.DoIgnore(roundCtx, own)
	return resp.NewEntry[any]("ignore")
}

//  胡牌
func win(w http.ResponseWriter, r *http.Request, body *api.WinParameter) *resp.Entry[*api.WinResult] {

	header := GetHeader(r)
	//上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return resp.NewError[*api.WinResult](err)
	}
	//玩家信息
	own, _ := roundCtx.Player(header.UserId)

	//加锁防止并发操作 等待
	defer roundCtx.Lock.Unlock()
	roundCtx.Lock.Lock()

	winResult, err := service.DoWin(roundCtx, own)
	if err != nil {
		return resp.NewError[*api.WinResult](err)
	}
	return resp.NewEntry(winResult)
}
