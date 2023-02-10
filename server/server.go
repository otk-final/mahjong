package server

import (
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"mahjong/server/wrap"
	"mahjong/server/ws"
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
	game.Path("/robot").HandlerFunc(wrap.NewWrapper(robot).Func())

	//玩家事件
	play := muxRouter.Methods("POST").PathPrefix("/play").Subrouter()
	play.Path("/take").HandlerFunc(wrap.NewWrapper(take).Func())
	play.Path("/put").HandlerFunc(wrap.NewWrapper(put).Func())
	play.Path("/race").HandlerFunc(wrap.NewWrapper(race).Func())
	play.Path("/race-pre").HandlerFunc(wrap.NewWrapper(racePre).Func())
	play.Path("/win").HandlerFunc(wrap.NewWrapper(win).Func())
	play.Path("/ignore").HandlerFunc(wrap.NewWrapper(ignore).Func())

	//websocket链接
	muxRouter.Handle("/ws/{RoomId}", ws.Route())
	return muxCors.Handler(muxRouter)
}
