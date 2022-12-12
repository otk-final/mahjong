package server

import (
	"errors"
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/wrap"
	"net/http"
	"sync"
)

// 加入房间
func join(w http.ResponseWriter, r *http.Request, body *api.JoinRoom) (*api.EmptyData, error) {

	//用户信息
	header := wrap.GetHeader(r)
	userId := header.UserId

	//查询房间信息
	room, err := roomQuery(body.RoomId)
	if err != nil {
		return nil, err
	}

	//是否满员
	_, ok := room.Ready()
	if ok {
		return nil, errors.New("房间已满")
	}

	//加入房间
	room.Join(mj.NewPlayer(userId))
	return api.Empty, nil
}

// 退出房间
func exit(w http.ResponseWriter, r *http.Request, body *api.ExitRoom) (*api.EmptyData, error) {

	//用户信息
	header := wrap.GetHeader(r)
	userId := header.UserId

	//查询房间信息
	room, err := roomQuery(body.RoomId)
	if err != nil {
		return nil, err
	}
	//查询玩家，并退出
	p, err := room.Player(userId)
	if err != nil {
		return nil, err
	}
	room.Exit(p.Idx)

	return nil, nil
}

var roomManager = &sync.Map{}

func storeRoom(roomId string, room *mj.Room) {
	roomManager.Store(roomId, room)
}

func roomQuery(roomId string) (*mj.Room, error) {
	temp, ok := roomManager.Load(roomId)
	if !ok {
		return nil, errors.New("room not found")
	}
	return temp.(*mj.Room), nil
}
