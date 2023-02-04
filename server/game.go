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

// 开始游戏
func start(w http.ResponseWriter, r *http.Request, body *api.GameStart) (*api.NoResp, error) {

	//用户信息
	header := wrap.GetHeader(r)

	//当前就坐信息
	pos, err := store.GetPosition(body.RoomId)
	if err != nil {
		return nil, err
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
	exchanger := engine.NewExchanger(30)
	go exchanger.Run(notifyHandler, pos)

	//注册缓存
	store.RegisterRoundCtx(body.RoomId, pos, exchanger, roundCtxOps)

	//通知牌局开始
	BroadcastFunc(startDispatcher, func(player *api.Player) *api.WebPacket[api.BeginPayload] {

		//从庄家开始
		startPayload := api.BeginPayload{
			TurnIdx:  0,
			Remained: roundCtxOps.Remained(),
			Tiles:    make([]*api.PlayerTiles, 0),
		}
		currentIdx := player.Idx
		for _, user := range startDispatcher.members {
			tiles := roundCtxOps.LoadTiles(user.Idx).Copy(currentIdx == user.Idx)
			startPayload.Tiles = append(startPayload.Tiles, tiles)
		}
		return api.Packet(api.BeginEvent, "开始", startPayload)
	})
	return api.Empty, nil
}

//查询玩家牌库
func load(w http.ResponseWriter, r *http.Request, body *api.GamePlayerQuery) (*api.GamePlayerInf, error) {
	//用户信息
	header := wrap.GetHeader(r)

	//查询上下文
	ctx, err := store.LoadRoundCtx(body.RoomId, header.UserId)
	if err != nil {
		return nil, err
	}

	//查询基础牌信息
	tiles := ctx.Handler.LoadTiles(body.Idx)
	profits := ctx.Handler.LoadProfits(body.Idx)

	//查询玩家信息
	//member, _ := ctx.Position.Index(header.UserId)
	//isOwn := strings.EqualFold(member.UId, header.UserId)

	//TODO 非自己的牌，查询是否选择明牌

	return &api.GamePlayerInf{
		RoomId:  body.RoomId,
		Idx:     body.Idx,
		Tiles:   tiles,
		Profits: profits,
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
	log.Printf("广播：take\n")
	Broadcast(handler.dispatcher, api.Packet(api.TakeEvent, "摸牌", event))
}

func (handler *BroadcastHandler) Put(ackId int, event *api.PutPayload) {
	log.Printf("广播：put\n")
	Broadcast(handler.dispatcher, api.Packet(api.PutEvent, "打牌", event))
}

func (handler *BroadcastHandler) Race(event *api.RacePayload) {
	log.Printf("广播：race %d %s\n", event.RaceType, api.RaceNames[event.RaceType])
	Broadcast(handler.dispatcher, api.Packet(api.RaceEvent, api.RaceNames[event.RaceType], event))
}

func (handler *BroadcastHandler) Win(event *api.WinPayload) bool {
	log.Printf("广播：win\n")
	Broadcast(handler.dispatcher, api.Packet(api.WinEvent, "胡牌", event))
	return handler.provider.Finish()
}

func (handler *BroadcastHandler) Ack(event *api.AckPayload) {
	log.Printf("广播：ack\n")
	Broadcast(handler.dispatcher, api.Packet(api.AckEvent, "待确认", event))
}

func (handler *BroadcastHandler) Turn(who int, duration int, ok bool) {
	log.Printf("广播：turn %d\n", who)
	Broadcast(handler.dispatcher, api.Packet(api.TurnEvent, "轮转", &api.TurnPayload{Who: who, Duration: duration}))
}

func (handler *BroadcastHandler) Quit(ok bool) {
	handler.provider.Quit()
}
