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
	room.Path("/create").HandlerFunc(wrap.NewWrapper(create, true).Func())
	room.Path("/join").HandlerFunc(wrap.NewWrapper(join, true).Func())
	room.Path("/exit").HandlerFunc(wrap.NewWrapper(exit, true).Func())
	room.Path("/compute").HandlerFunc(wrap.NewWrapper(compute, true).Func())
	room.Path("/visitor").HandlerFunc(wrap.NewWrapper(visitor, false).Func())

	//游戏
	game := muxRouter.Methods("POST").PathPrefix("/game").Subrouter()
	game.Path("/start").HandlerFunc(wrap.NewWrapper(start, true).Func())
	game.Path("/load").HandlerFunc(wrap.NewWrapper(load, true).Func())
	game.Path("/robot").HandlerFunc(wrap.NewWrapper(robot, true).Func())

	//玩家事件
	play := muxRouter.Methods("POST").PathPrefix("/play").Subrouter()
	play.Path("/take").HandlerFunc(wrap.NewWrapper(take, true).Func())
	play.Path("/put").HandlerFunc(wrap.NewWrapper(put, true).Func())
	play.Path("/race").HandlerFunc(wrap.NewWrapper(race, true).Func())
	play.Path("/race-pre").HandlerFunc(wrap.NewWrapper(racePre, true).Func())
	play.Path("/win").HandlerFunc(wrap.NewWrapper(win, true).Func())
	play.Path("/ignore").HandlerFunc(wrap.NewWrapper(ignore, true).Func())

	//websocket链接
	muxRouter.Handle("/ws/{RoomId}", ws.Route())
	return muxCors.Handler(muxRouter)
}
