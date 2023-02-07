package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"mahjong/server/api"
	"mahjong/server/wrap"
	"net/http"
)

//http api
var muxRouter = mux.NewRouter()

//跨域规则
var muxCors = cors.AllowAll()

func Apis() http.Handler {

	//房间
	room := muxRouter.Methods("POST").PathPrefix("/room").Subrouter()
	room.Path("/create").HandlerFunc(wrap.NewWrapper(create).Func())
	room.Path("/join").HandlerFunc(wrap.NewWrapper(join).Func())
	room.Path("/exit").HandlerFunc(wrap.NewWrapper(exit).Func())

	//游戏
	game := muxRouter.Methods("POST").PathPrefix("/game").Subrouter()
	game.Path("/start").HandlerFunc(wrap.NewWrapper(start).Func())
	game.Path("/load").HandlerFunc(wrap.NewWrapper(load).Func())

	//玩家事件
	play := muxRouter.Methods("POST").PathPrefix("/play").Subrouter()
	play.Path("/take").HandlerFunc(wrap.NewWrapper(take).Func())
	play.Path("/put").HandlerFunc(wrap.NewWrapper(put).Func())
	play.Path("/race").HandlerFunc(wrap.NewWrapper(race).Func())
	play.Path("/race-pre").HandlerFunc(wrap.NewWrapper(racePre).Func())
	play.Path("/win").HandlerFunc(wrap.NewWrapper(win).Func())
	play.Path("/ignore").HandlerFunc(wrap.NewWrapper(ignore).Func())

	//websocket链接
	muxRouter.Handle("/ws/{RoomId}", wsRoute())
	return muxCors.Handler(muxRouter)
}

type RoomDispatcher struct {
	RoomId  string
	members []*api.Player
}

func (rx *RoomDispatcher) GetPlayer(acctId string) (*netChan, error) {
	chKey := fmt.Sprintf("%s#%s", rx.RoomId, acctId)
	temp, ok := netChanMap.Load(chKey)
	if !ok {
		return nil, errors.New("not connected")
	}
	return temp.(*netChan), nil
}

func Broadcast[T any](dispatcher *RoomDispatcher, packet *api.WebPacket[T]) {
	//序列化 json
	msg, _ := json.Marshal(packet)
	//所有成员
	for _, member := range dispatcher.members {
		memberChan, err := dispatcher.GetPlayer(member.UId)
		if err != nil {
			continue
		}
		memberChan.write <- msg
	}
}

func BroadcastFunc[T any](dispatcher *RoomDispatcher, fn func(*api.Player) *api.WebPacket[T]) {
	//所有成员
	for _, member := range dispatcher.members {
		packet := fn(member)
		memberChan, err := dispatcher.GetPlayer(member.UId)
		if err != nil {
			continue
		}
		//序列化 json
		msg, _ := json.Marshal(packet)
		memberChan.write <- msg
	}
}
