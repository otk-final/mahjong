package server

import (
	"fmt"
	"github.com/google/uuid"
	"hash/crc32"
	"mahjong/server/api"
	"mahjong/server/engine"
	"mahjong/server/store"
	"mahjong/server/wrap"
	"net/http"
)

// 创建房间
func create(w http.ResponseWriter, r *http.Request, body *api.CreateRoom) (*api.RoomInf, error) {

	//用户信息
	header := wrap.GetHeader(r)
	master := &api.Player{
		Idx:    0,
		AcctId: header.UserId,
		Alias:  "庄家",
	}

	//生成房间号
	roomId := roomIdGen()

	//save 配置信息
	store.CreateRoom(roomId, body.Game, body.Payment)

	//设置庄家，虚位待坐 join
	pos, _ := engine.NewPosition(body.Game.Nums, master)
	_ = pos.Join(master)

	//save 座位信息
	store.CreatePosition(roomId, pos)

	//房间信息
	return roomInfQuery(roomId)
}

// 生成房间号
func roomIdGen() string {
	id := crc32.ChecksumIEEE([]byte(uuid.New().String()))
	return fmt.Sprintf("%d", id)
}

// 加入房间
func join(w http.ResponseWriter, r *http.Request, body *api.JoinRoom) (*api.RoomInf, error) {

	//用户信息
	header := wrap.GetHeader(r)

	//查询座位信息
	pos, err := store.GetPosition(body.RoomId)
	if err != nil {
		return nil, err
	}

	//自动选座 idx = -1
	member := &api.Player{
		Idx:    -1,
		AcctId: header.UserId,
		Name:   "",
		Alias:  "闲家",
	}

	//入座
	err = pos.Join(member)
	if err != nil {
		return nil, err
	}

	//update
	err = store.UpdatePosition(body.RoomId, pos)
	if err != nil {
		return nil, err
	}

	//通知有新玩家加入
	joins := pos.Joined()
	rx := &RoomDispatcher{RoomId: body.RoomId, members: joins}
	Broadcast(rx, api.Packet(99, &api.JoinPayload{Member: member, Round: 0}))

	//房间信息
	return roomInfQuery(body.RoomId)
}

// 退出房间
func exit(w http.ResponseWriter, r *http.Request, body *api.ExitRoom) (*api.NoResp, error) {

	return api.Empty, nil
}

func roomInfQuery(roomId string) (*api.RoomInf, error) {
	return nil, nil
}
