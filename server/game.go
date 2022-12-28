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

	//TODO 每种玩法，独立逻辑处理
	var handler ploy.GameDefine
	switch gc.Mode {
	case "laiz": //赖子
		break
	case "k5x": //卡5星
		break
	case "7d": //七对
		break
	case "sc": //四川
		break
	case "gz": //广东
		break
	}
	//前置事件 初始化牌库
	tileHandler := handler.Init(gc, pc)
	notifyHandler := &broadcastHandler{
		proxy:       handler,
		tileHandler: tileHandler,
	}

	//开启计时器
	exchanger := engine.NewExchanger(30 * time.Second)
	go exchanger.Run(notifyHandler, pos)

	//注册缓存
	store.RegisterRoundCtx(body.RoomId, pos, exchanger, tileHandler)
	//TODO 通知牌局开始

	Broadcast()

	return api.Empty, nil
}

type broadcastHandler struct {
	proxy       ploy.GameDefine
	roundCtx    store.RoundCtx
	tileHandler engine.TileHandle
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
	return handler.proxy.Finish()
}

func (handler *broadcastHandler) Ack(event *api.AckPayload) {
	Broadcast(handler.dispatcher, api.Packet(105, event))
}

func (handler *broadcastHandler) Next(who int, ok bool) {
	Broadcast(handler.dispatcher, api.Packet(106, &api.NextPayload{Who: who}))
}

func (handler *broadcastHandler) Quit() {
	handler.proxy.Quit()
}
