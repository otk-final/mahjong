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
	if pos.IsMaster(header.UserId) {
		return nil, errors.New("待庄家开启游戏")
	}

	//判定是否满座
	if !pos.Ready() {
		return nil, errors.New("待玩家就坐")
	}
	//游戏设置
	gc, pc := store.GetRoomConfig(body.RoomId)

	//每种玩法，独立逻辑处理
	provider := ploy.NewProvider(gc.Mode)

	//前置事件 初始化牌库
	tileHandler := provider.Init(gc, pc)
	notifyHandler := &broadcastHandler{
		provider:    provider,
		tileHandler: tileHandler,
	}

	//开启计时器
	exchanger := engine.NewExchanger(30 * time.Second)
	go exchanger.Run(notifyHandler, pos)

	//注册缓存
	store.RegisterRoundCtx(body.RoomId, pos, exchanger, tileHandler)

	//通知牌局开始
	Broadcast(nil, api.Packet(1, nil))

	return api.Empty, nil
}

//下一局
func next(w http.ResponseWriter, r *http.Request, body *api.GameStart) (*api.NoResp, error) {

	return nil, nil
}

type broadcastHandler struct {
	provider    ploy.GameDefine
	roundCtx    store.RoundCtx
	tileHandler engine.RoundCtxHandle
	dispatcher  *RoomDispatcher
}

func (handler *broadcastHandler) Take(event *api.TakePayload) {
	Broadcast(handler.dispatcher, api.Packet(100, event))
}

func (handler *broadcastHandler) Put(ackId int, event *api.PutPayload) {

	//广播出牌事件
	Broadcast(handler.dispatcher, api.Packet(101, event))

	//广播待确认事件
	Broadcast(handler.dispatcher, api.Packet(102, &api.AckPayload{
		Who:   event.Who,
		Round: event.Round,
		AckId: ackId,
	}))
}

func (handler *broadcastHandler) Race(event *api.RacePayload) {
	Broadcast(handler.dispatcher, api.Packet(103, event))
}

func (handler *broadcastHandler) Win(event *api.RacePayload) bool {
	Broadcast(handler.dispatcher, api.Packet(104, event))
	return handler.provider.Finish()
}

func (handler *broadcastHandler) Ack(event *api.AckPayload) {
	Broadcast(handler.dispatcher, api.Packet(105, event))
}

func (handler *broadcastHandler) Next(who int, ok bool) {
	Broadcast(handler.dispatcher, api.Packet(106, &api.NextPayload{Who: who}))
}

func (handler *broadcastHandler) Quit() {
	handler.provider.Quit()
}
