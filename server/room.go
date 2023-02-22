package server

import (
	"fmt"
	"github.com/otk-final/thf/resp"
	"log"
	"mahjong/server/api"
	"mahjong/server/broadcast"
	"mahjong/service/engine"
	"mahjong/service/store"
	"net/http"
)

//  创建房间
func create(w http.ResponseWriter, r *http.Request, body *api.GameConfigure) *resp.Entry[*api.RoomInf] {

	//用户信息
	header := GetHeader(r)
	master := &api.Player{
		Idx:   0,
		UId:   header.UserId,
		UName: header.UserName,
		Alias: "庄家",
	}

	//save 配置信息
	roomId := store.CreateRoom(body)

	//设置庄家，虚位待坐 join
	pos, _ := engine.NewPosition(body.Nums, master)

	//save 座位信息
	store.CreatePosition(roomId, pos)

	//房间信息
	return resp.NewEntry(&api.RoomInf{
		RoomId:  roomId,
		Own:     master,
		Players: []*api.Player{},
		Config:  body,
	})
}

//  加入房间
func join(w http.ResponseWriter, r *http.Request, body *api.JoinRoom) *resp.Entry[*api.RoomInf] {

	//用户信息
	header := GetHeader(r)

	//查询座位信息
	pos, err := store.GetPosition(body.RoomId)
	if err != nil {
		return resp.NewError[*api.RoomInf](err)
	}

	//自动选座 idx = -1
	member := &api.Player{
		Idx:    -1,
		UId:    header.UserId,
		UName:  header.UserName,
		Alias:  "闲家",
		Avatar: "",
		Ip:     r.RemoteAddr,
	}

	//是否已就坐
	exist, err := pos.Index(header.UserId)
	if exist != nil {
		member = exist
	} else {
		//入座
		err = pos.Join(member)
		if err != nil {
			return resp.NewError[*api.RoomInf](err)
		}

		//update
		store.UpdatePosition(body.RoomId, pos)

		//通知新玩家加入
		broadcast.Post(body.RoomId, pos.Joined(), api.Packet(api.JoinEvent, "加入", &api.JoinPayload{NewPlayer: member}))
	}

	//判定游戏是否开始
	return resp.NewEntry(&api.RoomInf{
		RoomId:  body.RoomId,
		Own:     member,
		Begin:   pos.TurnIdx() != -1,
		Players: pos.Joined(),
		Config:  store.GetRoomConfig(body.RoomId),
	})
}

//  退出房间
func exit(w http.ResponseWriter, r *http.Request, body *api.ExitRoom) *resp.Entry[any] {
	//用户信息
	header := GetHeader(r)
	store.FreeVisitor(header.UserId)
	return resp.NewAny("exit")
}

//人机对战
func compute(w http.ResponseWriter, r *http.Request, body *api.GameConfigure) *resp.Entry[api.RoomInf] {
	//创建房间
	header := GetHeader(r)
	//生成房间号
	roomId := store.CreateRoom(body)

	//设置座位 庄家 + 机器人
	master := &api.Player{
		Idx:   0,
		UId:   header.UserId,
		UName: header.UserName,
		Alias: "庄家",
	}
	robots := make([]*api.Roboter, 0)
	for i := 0; i < body.Nums-1; i++ {
		roboter := &api.Roboter{
			Player: &api.Player{Idx: i + 1, UId: "robot1", UName: fmt.Sprintf("%d号", i+1), Alias: "闲家"},
			Level:  i + 1,
		}
		robots = append(robots, roboter)
	}
	pos, _ := engine.NewPositionRobots(body.Nums, master, robots...)
	store.CreatePosition(roomId, pos)

	return resp.NewEntry(api.RoomInf{
		RoomId:  roomId,
		Own:     master,
		Begin:   pos.TurnIdx() != -1,
		Players: pos.Joined(),
		Config:  body,
	})
}

func visitor(w http.ResponseWriter, r *http.Request, body *api.VisitorParameter) *resp.Entry[*api.Visitor] {
	log.Printf("获取游客信息：%s", r.RemoteAddr)
	vs, err := store.NewVisitor(r)
	if err != nil {
		return resp.NewError[*api.Visitor](err)
	}
	return resp.NewEntry(vs)
}
