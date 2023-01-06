package server

import (
	"errors"
	"mahjong/ploy"
	"mahjong/server/api"
	"mahjong/server/engine"
	"mahjong/server/store"
	"mahjong/server/wrap"
	"net/http"
	"time"
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
	if pos.Len() != pos.Cap() {
		return nil, errors.New("待玩家就坐")
	}
	//游戏设置
	gc, pc := store.GetRoomConfig(body.RoomId)

	//每种玩法，独立逻辑处理
	provider := ploy.NewProvider(gc.Mode)

	//前置事件 初始化牌库
	roundCtxOps := provider.InitCtx(gc, pc)
	dispatcher := &RoomDispatcher{RoomId: body.RoomId, members: pos.Joined()}
	notifyHandler := &broadcastHandler{
		provider:    provider,
		roundCtxOps: roundCtxOps,
		dispatcher:  dispatcher,
	}

	//开启计时器
	exchanger := engine.NewExchanger(30 * time.Second)
	go exchanger.Run(notifyHandler, pos)

	//注册缓存
	store.RegisterRoundCtx(body.RoomId, pos, exchanger, roundCtxOps)

	//通知牌局开始
	Broadcast(dispatcher, api.Packet(api.BeginEvent, nil))

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
	//isOwn := strings.EqualFold(member.AcctId, header.UserId)

	//TODO 非自己的牌，查询是否选择明牌

	return &api.GamePlayerInf{
		RoomId:  body.RoomId,
		Idx:     body.Idx,
		Tiles:   tiles,
		Profits: profits,
	}, nil
}

type broadcastHandler struct {
	provider    ploy.GameDefine
	roundCtx    engine.RoundCtx
	roundCtxOps engine.RoundOpsCtx
	dispatcher  *RoomDispatcher
}

func (handler *broadcastHandler) Take(event *api.TakePayload) {
	Broadcast(handler.dispatcher, api.Packet(api.TakeEvent, event))
}

func (handler *broadcastHandler) Put(ackId int, event *api.PutPayload) {
	Broadcast(handler.dispatcher, api.Packet(api.PutEvent, event))
}

func (handler *broadcastHandler) Race(event *api.RacePayload) {
	Broadcast(handler.dispatcher, api.Packet(api.RaceEvent, event))
}

func (handler *broadcastHandler) Win(event *api.RacePayload) bool {
	Broadcast(handler.dispatcher, api.Packet(api.WinEvent, event))
	return handler.provider.Finish()
}

func (handler *broadcastHandler) Ack(event *api.AckPayload) {
	Broadcast(handler.dispatcher, api.Packet(api.AckEvent, event))
}

func (handler *broadcastHandler) Turn(who int, ok bool) {
	Broadcast(handler.dispatcher, api.Packet(api.TurnEvent, &api.NextPayload{Who: who}))
}

func (handler *broadcastHandler) Quit() {
	handler.provider.Quit()
}
