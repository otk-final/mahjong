package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"mahjong/server/api"
	"mahjong/server/wrap"
)

//http api
var muxRouter = mux.NewRouter()

func Apis() *mux.Router {

	//房间
	room := muxRouter.Methods("POST").PathPrefix("/room").Subrouter()
	room.Path("/create").HandlerFunc(wrap.NewWrapper(create).Func())
	room.Path("/join").HandlerFunc(wrap.NewWrapper(join).Func())
	room.Path("/exit").HandlerFunc(wrap.NewWrapper(exit).Func())

	//游戏
	game := muxRouter.Methods("POST").PathPrefix("/game").Subrouter()
	game.Path("/start").HandlerFunc(wrap.NewWrapper(start).Func())

	//玩家事件
	play := muxRouter.Methods("POST").PathPrefix("/play").Subrouter()
	play.Path("/take").HandlerFunc(wrap.NewWrapper(take).Func())
	play.Path("/put").HandlerFunc(wrap.NewWrapper(put).Func())
	play.Path("/race").HandlerFunc(wrap.NewWrapper(race).Func())
	play.Path("/race-pre").HandlerFunc(wrap.NewWrapper(racePre).Func())
	play.Path("/win").HandlerFunc(wrap.NewWrapper(win).Func())
	play.Path("/skip").HandlerFunc(wrap.NewWrapper(skip).Func())

	//websocket链接
	muxRouter.Handle("/ws/{RoomId}", wsRoute())
	return muxRouter
}

type RoomDispatcher struct {
	RoomId  string
	members []*api.Player
}

func (rx *RoomDispatcher) GetPlayer(acctId string) *netChan {
	chKey := fmt.Sprintf("%s#%s", rx.RoomId, acctId)
	temp, ok := netChanMap.Load(chKey)
	if !ok {
		return nil
	}
	return temp.(*netChan)
}

func Broadcast[T any](dispatcher *RoomDispatcher, packet *api.WebPacket[T]) {
	//序列化 json
	msg, _ := json.Marshal(packet)
	//所有成员
	for _, member := range dispatcher.members {
		memberChan := dispatcher.GetPlayer(member.AcctId)
		memberChan.write <- msg
	}
}
