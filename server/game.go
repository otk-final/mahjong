package server

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/countdown"
	"mahjong/server/wrap"
	"net/http"
)

type RoomCtrl struct{}
type GameCtrl struct{}

var roomCtrl = &RoomCtrl{}

func (room *RoomCtrl) inf(roomId string) (*api.RoomInf, error) {
	//模拟玩家
	ps := make(map[int]api.Identity, 0)
	for i := 0; i < 4; i++ {
		ps[i] = api.Identity{
			UserId:   fmt.Sprintf("no%d", i),
			Token:    uuid.NewString(),
			UserName: fmt.Sprintf("玩家%d", i),
		}
	}

	return &api.RoomInf{
		RoomId:  roomId,
		Players: ps,
		Config: &api.GameConfigure{
			Players:  4,
			HasWind:  false,
			HasOther: false,
		},
	}, nil
}

func (room *RoomCtrl) join(w http.ResponseWriter, r *http.Request, body *api.JoinRoom) {

}

// 开始游戏
func (game *GameCtrl) start(w http.ResponseWriter, r *http.Request, body *api.GameRun) (*api.EmptyData, error) {

	//用户信息
	header := wrap.GetHeader(r)
	userId := header.UserId

	//校验房间号
	roomId := body.RoomId
	room, err := roomCtrl.inf(roomId)
	if err != nil {
		return nil, err
	}
	//判断玩家是否到齐
	if room.Config.Players != len(room.Players) {
		return nil, errors.New("玩家未就位")
	}

	//1 丢骰,创建牌桌
	dice := mj.NewDice()
	tb := mj.NewTable(len(room.Players), dice)

	//2 根据当前配置获取牌库
	var mjLib = make([]int, 0)
	mjLib = mj.LoadLibrary(room.Config.HasWind, room.Config.HasOther)

	//3 洗牌,发牌
	mjLib = tb.Shuffle(mjLib)
	mCards := tb.Allocate(mjLib)

	//存储状态
	game.storeTable(roomId, tb)

	//跟踪准备事件
	traceId := countdown.NewTrackId(roomId)

	//4 通知所有玩家
	for k, v := range mCards {
		websocketNotify(roomId, room.Players[k].UserId, 100, traceId, api.GameReadyPacket{Cards: v})
	}

	//5 倒计时 & 通知庄家摸牌
	countdown.New[api.GameReadyAck](traceId, len(room.Players), 30, func(data countdown.CallData[api.GameReadyAck]) {
		master := tb.TurnPlayer()
		websocketNotify(roomId, master.Id, 300, "", nil)
	})
	return api.Empty, nil
}

//准备确认
func (game *GameCtrl) readyAck(w http.ResponseWriter, r *http.Request, body *api.GameReadyAck) (*api.EmptyData, error) {
	countdown.Ready(body.EventId, body)
	return api.Empty, nil
}

// 缓存牌桌状态
func (game *GameCtrl) storeTable(roomId string, tb *mj.Table) {

}

func (game *GameCtrl) tableQuery(roomId string) *mj.Table {

	return nil
}
