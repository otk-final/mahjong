package server

import (
	"fmt"
	"github.com/google/uuid"
	"hash/crc32"
	"log"
	"mahjong/server/api"
	"mahjong/server/broadcast"
	"mahjong/server/wrap"
	"mahjong/service/engine"
	"mahjong/service/store"
	"net/http"
)

//  创建房间
func create(w http.ResponseWriter, r *http.Request, body *api.GameConfigure) (*api.RoomInf, error) {

	//用户信息
	header := wrap.GetHeader(r)
	master := &api.Player{
		Idx:   0,
		UId:   header.UserId,
		UName: header.UserName,
		Alias: "庄家",
	}

	//生成房间号
	roomId := roomIdGen()

	//save 配置信息
	store.CreateRoom(roomId, body)

	//设置庄家，虚位待坐 join
	pos, _ := engine.NewPosition(body.Nums, master)

	//save 座位信息
	store.CreatePosition(roomId, pos)

	//房间信息
	return &api.RoomInf{
		RoomId:  roomId,
		Own:     master,
		Players: []*api.Player{},
		Config:  body,
	}, nil
}

// 生成房间号
func roomIdGen() string {
	id := crc32.ChecksumIEEE([]byte(uuid.New().String()))
	return fmt.Sprintf("%d", id)
}

//  加入房间
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
			return nil, err
		}

		//update
		store.UpdatePosition(body.RoomId, pos)

		//通知新玩家加入
		broadcast.Post(body.RoomId, pos.Joined(), api.Packet(api.JoinEvent, "加入", &api.JoinPayload{NewPlayer: member}))
	}

	//判定游戏是否开始
	return &api.RoomInf{
		RoomId:  body.RoomId,
		Own:     member,
		Begin:   pos.TurnIdx() != -1,
		Players: pos.Joined(),
		Config:  store.GetRoomConfig(body.RoomId),
	}, nil
}

//  退出房间
func exit(w http.ResponseWriter, r *http.Request, body *api.ExitRoom) (*api.NoResp, error) {

	//用户信息
	header := wrap.GetHeader(r)

	store.FreeVisitor(header.UserId)
	return api.Empty, nil
}

//人机对战
func compute(w http.ResponseWriter, r *http.Request, body *api.GameConfigure) (*api.RoomInf, error) {
	//创建房间
	header := wrap.GetHeader(r)
	//生成房间号
	roomId := roomIdGen()
	store.CreateRoom(roomId, body)

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

	return &api.RoomInf{
		RoomId:  roomId,
		Own:     master,
		Begin:   pos.TurnIdx() != -1,
		Players: pos.Joined(),
		Config:  body,
	}, nil
}

func visitor(w http.ResponseWriter, r *http.Request, body *api.VisitorParameter) (*api.Visitor, error) {
	log.Printf("获取游客信息：%s", r.RemoteAddr)
	return store.NewVisitor(r)
}
