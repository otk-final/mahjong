package server

import (
	"container/ring"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/countdown"
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
			Gamer:    4,
			HasWind:  false,
			HasOther: false,
		},
	}, nil
}

func (room *RoomCtrl) join(w http.ResponseWriter, r *http.Request, body *api.JoinRoom) {

}

// 开始游戏
func (game *GameCtrl) start(w http.ResponseWriter, r *http.Request, body *api.GameRun) (*api.EmptyData, error) {

	userId := r.Header.Get("user_id")

	//校验房间号
	roomId := body.RoomId
	room, err := roomCtrl.inf(roomId)
	if err != nil {
		return nil, err
	}
	//判断玩家是否到齐
	if room.Config.Gamer != len(room.Players) {
		return nil, errors.New("玩家未就位")
	}

	//1 丢骰
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

	//开启延迟队列
	timerId := uuid.NewString()
	countdown.New(timerId, len(room.Players), 30, func(status countdown.Type, inf *countdown.Timer) {
		//超时
		if status == 0 {
			return
		}
		//通知当前用户摸第一张牌
		websocketNotify(roomId, userId, 102, "", api.Empty)
	})

	//4 通知所有玩家
	for k, v := range mCards {
		websocketNotify(roomId, room.Players[k].UserId, 100, timerId, api.InitializeGamePayload{Cards: v})
	}
	return api.Empty, nil
}

//初始牌库
func (game *GameCtrl) status(w http.ResponseWriter, r *http.Request, body *api.GameAck) (*api.InitializeGamePayload, error) {

	//查询当前牌库
	userId := r.Header.Get("user_id")
	tb := game.tableQuery("")

	return &api.InitializeGamePayload{Cards: []int{}}, nil
}

//回执确认
func (game *GameCtrl) ack(w http.ResponseWriter, r *http.Request, body *api.GameAck) (*api.EmptyData, error) {

	countdown.Ack(body.EventId)

	return api.Empty, nil
}

// 缓存牌桌状态
func (game *GameCtrl) storeTable(roomId string, tb *mj.Table) {

}

func (game *GameCtrl) tableQuery(roomId string) *mj.Table {

	r := ring.New(4)
	r.Value =

	return nil
}

