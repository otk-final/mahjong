package server

import (
	"errors"
	"log"
	"mahjong/ploy"
	"mahjong/server/api"
	"mahjong/server/engine"
	"mahjong/server/store"
	"mahjong/server/wrap"
	"net/http"
)

const turnInterval = 30

// 开始游戏
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
	gc, pc := store.GetRoomConfig(body.RoomId)

	//每种玩法，独立逻辑处理
	provider := ploy.NewProvider(gc.Mode)

	//前置事件 初始化牌库
	roundCtxOps := provider.InitCtx(gc, pc)
	startDispatcher := &RoomDispatcher{RoomId: body.RoomId, members: pos.Joined()}
	notifyHandler := &BroadcastHandler{
		provider:   provider,
		dispatcher: startDispatcher,
	}

	//开启计时器
	exchanger := engine.NewExchanger()
	exchanger.Run(notifyHandler, pos, turnInterval)

	//注册缓存
	store.RegisterRoundCtx(body.RoomId, pos, exchanger, roundCtxOps)

	//通知牌局开始
	BroadcastFunc(startDispatcher, func(player *api.Player) *api.WebPacket[api.GamePayload] {
		//从庄家开始
		startPayload := api.GamePayload{
			TurnIdx:  0,
			Interval: turnInterval,
			Remained: roundCtxOps.Remained(),
			Players:  make([]*api.PlayerTiles, 0),
		}
		currentIdx := player.Idx
		for _, user := range startDispatcher.members {
			tiles := roundCtxOps.GetTiles(user.Idx).ExplicitCopy(currentIdx == user.Idx)
			startPayload.Players = append(startPayload.Players, tiles)
		}
		return api.Packet(api.BeginEvent, "开始", startPayload)
	})
	return api.Empty, nil
}

//查询玩家牌库
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
	roundCtxOps := roundCtx.HandlerCtx()
	joined := roundCtx.Pos().Joined()
	userTiles := make([]*api.PlayerTiles, 0)
	for _, user := range joined {
		//非自己的牌，查询是否选择明牌
		tiles := roundCtxOps.GetTiles(user.Idx).ExplicitCopy(own.Idx == user.Idx)
		userTiles = append(userTiles, tiles)
	}

	usableRaces := make([]*api.RaceOption, 0)

	// 还原牌局

	//当前回合
	turnIdx := roundCtx.Pos().TurnIdx()

	//最后一次事件 ? 可能是自己，也可能是别人（摸牌，出牌，判定）
	recentAction := roundCtx.HandlerCtx().RecentAction()
	recentIdx := roundCtx.HandlerCtx().RecentIdx()

	//本回合 已摸牌，触发判定
	ownCheck := recentIdx == own.Idx && recentAction == engine.RecentTake
	//非本回合，已出牌，触发判定
	eachCheck := recentIdx != own.Idx && recentAction == engine.RecentPut
	if ownCheck || eachCheck {

		//目标牌
		recenter := roundCtx.HandlerCtx().Recenter(recentIdx)
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
		usableRaces, err = doRacePre(roundCtx, own, raceQuery)
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

type BroadcastHandler struct {
	roomId     string
	provider   ploy.GameDefine
	dispatcher *RoomDispatcher
}

func (handler *BroadcastHandler) RoundCtx(acctId string) (*engine.RoundCtx, error) {
	return store.LoadRoundCtx(handler.roomId, acctId)
}

func (handler *BroadcastHandler) Take(event *api.TakePayload) {
	Broadcast(handler.dispatcher, api.Packet(api.TakeEvent, "摸牌", event))
}

func (handler *BroadcastHandler) Put(event *api.PutPayload) {
	Broadcast(handler.dispatcher, api.Packet(api.PutEvent, "打牌", event))
}

func (handler *BroadcastHandler) Race(event *api.RacePayload) {
	Broadcast(handler.dispatcher, api.Packet(api.RaceEvent, api.RaceNames[event.RaceType], event))
}

func (handler *BroadcastHandler) Win(event *api.RacePayload) {
	Broadcast(handler.dispatcher, api.Packet(api.WinEvent, "胡牌", event))
}

func (handler *BroadcastHandler) Ack(event *api.AckPayload) {
	log.Printf("忽略：%d %d", event.Who, event.AckId)
	Broadcast(handler.dispatcher, api.Packet(api.AckEvent, "待确认", event))
}

func (handler *BroadcastHandler) Turn(who int, interval int, ok bool) {
	log.Printf("当前回合：%d %v", who, ok)
	Broadcast(handler.dispatcher, api.Packet(api.TurnEvent, "轮转", &api.TurnPayload{Who: who, Interval: interval}))
}

func (handler *BroadcastHandler) Quit(ok bool) {
	handler.provider.Quit()
}
