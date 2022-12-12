package server

import (
	"errors"
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/countdown"
	"mahjong/server/wrap"
	"net/http"
)

// 开始游戏
func start(w http.ResponseWriter, r *http.Request, body *api.GameRun) (*api.EmptyData, error) {

	//用户信息
	header := wrap.GetHeader(r)
	userId := header.UserId

	//校验房间号
	roomId := body.RoomId
	room := mj.Room{}

	//是否就位
	ps, ok := room.Ready()
	if !ok {
		return nil, errors.New("玩家未就位")
	}

	//1 丢骰,创建牌桌
	dice := mj.NewDice()
	tb := mj.NewTable(len(ps), dice)

	//2 根据当前配置获取牌库
	var mjLib = make([]int, 0)
	mjLib = mj.LoadLibrary(false, false)

	//3 洗牌,发牌
	mjLib = tb.Shuffle(mjLib)
	mcs := tb.Allocate(mjLib)

	//4，添加到玩家牌库
	for _, p := range ps {
		p.HandCards = mcs[p.Idx]
	}

	//存储状态
	storeTable(roomId, tb)

	//跟踪准备事件
	traceId := countdown.NewTrackId(roomId)
	//5 通知所有玩家
	for _, p := range ps {
		websocketNotify(roomId, p.Id, api.GameReadyEvent, traceId, api.GameReady{})
	}

	//开启倒计时 & 通知庄家摸牌
	countdown.New[api.GameReadyAck](traceId, len(ps), 30, func(data countdown.CallData[api.GameReadyAck]) {

		//从首摸牌
		headTake := api.TakeCard{
			RoomId:    body.RoomId,
			GameId:    "",
			Direction: 1,
		}

		master := room.TurnPlayer()
		websocketNotify(roomId, master.Id, api.TakeCardEvent, "", headTake)
	})
	return api.Empty, nil
}

//准备确认
func startAck(w http.ResponseWriter, r *http.Request, body *api.GameReadyAck) (*api.EmptyData, error) {
	countdown.Ready(body.EventId, body)
	return api.Empty, nil
}

func startLoad(w http.ResponseWriter, r *http.Request, body *api.GameReadyAck) (*api.EmptyData, error) {

	return nil, nil
}

// 缓存牌桌状态
func storeTable(roomId string, tb *mj.Table) {

}

func tableQuery(roomId string) (*mj.Table, error) {

	return nil, nil
}
