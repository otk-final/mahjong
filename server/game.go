package server

import (
	"errors"
	"mahjong/server/api"
	"mahjong/server/broadcast"
	"mahjong/server/wrap"
	"mahjong/service"
	"mahjong/service/engine"
	"mahjong/service/ploy"
	"mahjong/service/store"
	"net/http"
)

//  开始游戏
func start(w http.ResponseWriter, r *http.Request, body *api.GameParameter) (*api.NoResp, error) {

	//用户信息
	header := wrap.GetHeader(r)

	//当前就坐信息
	pos, err := store.GetPosition(body.RoomId)
	if err != nil {
		return nil, err
	}

	//已开始
	if pos.TurnIdx() != -1 {
		return api.Empty, nil
	}

	//校验用户信息，是否为庄家
	if !pos.IsMaster(header.UserId) {
		return nil, errors.New("待庄家开启游戏")
	}

	//判定是否满座
	if pos.Len() != pos.Num() {
		return nil, errors.New("待玩家就坐")
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
		}
		currentIdx := player.Idx
		for _, user := range joined {
			tiles := roundCtxOps.GetTiles(user.Idx).ExplicitCopy(currentIdx == user.Idx)
			startPayload.Players = append(startPayload.Players, tiles)
		}
		return api.Packet(api.BeginEvent, "开始", startPayload)
	})
	return api.Empty, nil
}

//  查询玩家牌库
func load(w http.ResponseWriter, r *http.Request, body *api.GameParameter) (*api.GameInf, error) {
	//用户信息
	header := wrap.GetHeader(r)

	//查询上下文
	roundCtx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return nil, err
	}
	own, _ := roundCtx.Player(header.UserId)

	//查询牌库
	roundCtxOps := roundCtx.Operating()
	joined := roundCtx.Pos().Joined()
	userTiles := make([]*api.PlayerTiles, 0)
	for _, user := range joined {
		//非自己的牌，查询是否选择明牌
		tiles := roundCtxOps.GetTiles(user.Idx).ExplicitCopy(own.Idx == user.Idx)
		userTiles = append(userTiles, tiles)
	}

	usableRaces := make([]*api.RaceOption, 0)

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
			Round:  0,
			AckId:  roundCtx.Exchange().CurrentAckId(),
			Target: recentIdx,
			Tile:   raceTile,
		}
		usableRaces = service.DoRacePre(roundCtx, own, raceQuery)
	} else {
		//如果是本回合，兜底显示出牌入口
		if turnIdx == own.Idx {
			usableRaces = append(usableRaces, &api.RaceOption{RaceType: api.PassRace})
		}
	}

	return &api.GameInf{
		GamePayload: &api.GamePayload{
			TurnIdx:   turnIdx,
			Interval:  roundCtx.Exchange().TurnTime(),
			RecentIdx: recentIdx,
			Remained:  roundCtxOps.Remained(),
			Players:   userTiles,
		},
		Options: usableRaces,
		RoomId:  body.RoomId,
	}, nil
}

//  挂机
func robot(w http.ResponseWriter, r *http.Request, body *api.RobotParameter) (*api.NoResp, error) {
	//用户信息
	header := wrap.GetHeader(r)
	//当前就坐信息
	pos, err := store.GetPosition(body.RoomId)
	if err != nil {
		return nil, err
	}
	//是否加入房间
	joinPlayer, err := pos.Index(header.UserId)
	if err != nil {
		return nil, err
	}

	if body.Open {
		pos.RobotOpen(joinPlayer, body.Level)
	} else {
		pos.RobotClosed(joinPlayer)
	}
	return api.Empty, nil
}
