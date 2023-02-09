package server

import (
	"mahjong/server/api"
	"mahjong/server/engine"
	"mahjong/server/store"
	"mahjong/server/wrap"
	"net/http"
)

// 创建房间
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
	//id := crc32.ChecksumIEEE([]byte(uuid.New().String()))
	//return fmt.Sprintf("%d", id)
	return "100"
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
		UId:    header.UserId,
		UName:  header.UserName,
		Alias:  "闲家",
		Avatar: "",
		Ip:     r.RemoteAddr,
	}
	//房间信息
	setting := store.GetRoomConfig(body.RoomId)

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
		rx := &RoomDispatcher{RoomId: body.RoomId, members: pos.Joined()}
		Broadcast(rx, api.Packet(api.JoinEvent, "加入", &api.JoinPayload{NewPlayer: member, Round: 0}))
	}

	//判定游戏是否开始
	return &api.RoomInf{
		RoomId:  body.RoomId,
		Own:     member,
		Begin:   pos.TurnIdx() != -1,
		Players: pos.Joined(),
		Config:  setting,
	}, nil
}

// 退出房间
func exit(w http.ResponseWriter, r *http.Request, body *api.ExitRoom) (*api.NoResp, error) {

	return api.Empty, nil
}
