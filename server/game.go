package server

import (
	"errors"
	"github.com/otk-final/thf/resp"
	"mahjong/server/api"
	"mahjong/server/broadcast"
	"mahjong/service"
	"mahjong/service/engine"
	"mahjong/service/ploy"
	"mahjong/service/store"
	"net/http"
)

//  开始游戏
func start(w http.ResponseWriter, r *http.Request, body *api.GameParameter) *resp.Entry[any] {

	//用户信息
	header := GetHeader(r)

	//当前就坐信息
	pos, err := store.GetPosition(body.RoomId)
	if err != nil {
		return resp.NewError[any](err)
	}

	//已开始
	if pos.TurnIdx() != -1 {
		return resp.NewError[any](errors.New("游戏已开始"))
	}

	//校验用户信息，是否为庄家
	if !pos.IsMaster(header.UserId) {
		return resp.NewError[any](errors.New("等待庄家开启游戏"))
	}

	//判定是否满座
	if pos.Len() != pos.Num() {
		return resp.NewError[any](errors.New("等待玩家就坐"))
	}
	//游戏设置
	setting := store.GetRoomConfig(body.RoomId)

	//每种玩法，独立逻辑处理
	provider := ploy.NewProvider(setting.Mode)

	//前置事件 初始化牌库
	roundCtxOps := provider.InitOperation(setting)
	joined := pos.Joined()
	//广播通知
	notifyHandler := &broadcast.Handler{
		RoomId:  body.RoomId,
		Pos:     pos,
		Players: joined,
	}

	//开启计时器
	exchanger := engine.NewExchanger(notifyHandler, pos)
	exchanger.Run(api.TurnInterval)

	//注册缓存
	store.CreateRoundCtx(body.RoomId, setting, pos, exchanger, roundCtxOps)

	//通知牌局开始
	broadcast.PostFunc(body.RoomId, joined, func(player *api.Player) *api.WebPacket[api.GamePayload] {
		//从庄家开始
		startPayload := api.GamePayload{
			TurnIdx:  0,
			Interval: api.TurnInterval,
			Remained: roundCtxOps.Remained(),
			Players:  make([]*api.PlayerTiles, 0),
			Extras:   provider.Extras(),
		}
		currentIdx := player.Idx
		for _, user := range joined {
			pt := roundCtxOps.GetTiles(user.Idx)
			startPayload.Players = append(startPayload.Players, pt.Visibility(currentIdx == user.Idx))
		}
		return api.Packet(api.BeginEvent, "开始", startPayload)
	})
	return resp.NewAny("started")
}

//  查询玩家牌库
func load(w http.ResponseWriter, r *http.Request, body *api.GameParameter) *resp.Entry[*api.GameInf] {
	//用户信息
	header := GetHeader(r)

	//查询上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return resp.NewError[*api.GameInf](err)
	}
	own, _ := roundCtx.Player(header.UserId)

	//查询牌库
	roundCtxOps := roundCtx.Operating()
	joined := roundCtx.Pos().Joined()
	userTiles := make([]*api.PlayerTiles, 0)
	for _, user := range joined {
		pt := roundCtxOps.GetTiles(user.Idx)
		userTiles = append(userTiles, pt.Visibility(own.Idx == user.Idx))
	}

	options := make([]*api.RaceOption, 0)

	// 还原牌局
	ops := roundCtx.Operating()

	//当前回合
	turnIdx := roundCtx.Pos().TurnIdx()

	//最后一次事件 ? 可能是自己，也可能是别人（摸牌，出牌，判定）
	recentAction := ops.RecentAction()
	recentIdx := ops.RecentIdx()

	//本回合 已摸牌，触发判定
	ownCheck := recentIdx == own.Idx && recentAction == engine.RecentTake
	//非本回合，已出牌，触发判定
	eachCheck := recentIdx != own.Idx && recentAction == engine.RecentPut
	if ownCheck || eachCheck {
		//目标牌
		recenter := ops.Recenter(recentIdx)
		raceTile := -1
		if ownCheck {
			raceTile = recenter.Take()
		} else {
			raceTile = recenter.Put()
		}
		//判定
		raceQuery := &api.RacePreview{
			RoomId: body.RoomId,
			Target: recentIdx,
			Tile:   raceTile,
		}
		options = service.DoRacePre(roundCtx, own, raceQuery)
	}

	//每种玩法，独立逻辑处理
	provider := ploy.RenewProvider(roundCtx)
	inf := &api.GameInf{
		GamePayload: &api.GamePayload{
			TurnIdx:   turnIdx,
			Interval:  roundCtx.Exchange().TurnTime(),
			RecentIdx: recentIdx,
			Remained:  roundCtxOps.Remained(),
			Players:   userTiles,
			Extras:    provider.Extras(),
		},
		Options: options,
		RoomId:  body.RoomId,
	}
	return resp.NewEntry(inf)
}

//  挂机
func robot(w http.ResponseWriter, r *http.Request, body *api.RobotParameter) *resp.Entry[any] {
	//用户信息
	header := GetHeader(r)
	//当前就坐信息
	pos, err := store.GetPosition(body.RoomId)
	if err != nil {
		return resp.NewError[any](err)
	}
	//是否加入房间
	joinPlayer, err := pos.Index(header.UserId)
	if err != nil {
		return resp.NewError[any](err)
	}

	if body.Open {
		pos.RobotOpen(joinPlayer, body.Level)
	} else {
		pos.RobotClosed(joinPlayer)
	}
	return resp.NewEntry[any]("robot")
}
