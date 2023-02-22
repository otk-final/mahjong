package server

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/otk-final/thf"
	"mahjong/server/api"
	"mahjong/server/ws"
	"mahjong/service/store"
	"net/http"
)

//http api
var muxRouter = mux.NewRouter()

func NewApiRouter() *mux.Router {

	//房间
	room := muxRouter.Methods("POST").PathPrefix("/room").Subrouter()
	room.Path("/create").HandlerFunc(thf.WrapIO(create).Func())
	room.Path("/join").HandlerFunc(thf.WrapIO(join).Func())
	room.Path("/exit").HandlerFunc(thf.WrapIO(exit).Func())
	room.Path("/compute").HandlerFunc(thf.WrapIO(compute).Func())
	room.Use(headerCheckIntercept)

	//游戏
	game := muxRouter.Methods("POST").PathPrefix("/game").Subrouter()
	game.Path("/start").HandlerFunc(thf.WrapIO(start).Func())
	game.Path("/load").HandlerFunc(thf.WrapIO(load).Func())
	game.Path("/robot").HandlerFunc(thf.WrapIO(robot).Func())
	game.Use(headerCheckIntercept)

	//玩家事件
	play := muxRouter.Methods("POST").PathPrefix("/play").Subrouter()
	play.Path("/take").HandlerFunc(thf.WrapIO(take).Func())
	play.Path("/put").HandlerFunc(thf.WrapIO(put).Func())
	play.Path("/race").HandlerFunc(thf.WrapIO(race).Func())
	play.Path("/race-pre").HandlerFunc(thf.WrapIO(racePre).Func())
	play.Path("/win").HandlerFunc(thf.WrapIO(win).Func())
	play.Path("/ignore").HandlerFunc(thf.WrapIO(ignore).Func())
	play.Use(headerCheckIntercept)

	vis := muxRouter.Methods("POST").PathPrefix("/room").Subrouter()
	vis.Path("/visitor").HandlerFunc(thf.WrapIO(visitor).Func())

	//websocket链接
	muxRouter.Handle("/ws/{RoomId}", ws.Route())

	return muxRouter
}

func headerCheckIntercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//header
		apiHeader := &api.IdentityHeader{
			UserId: r.Header.Get("userId"),
			Token:  r.Header.Get("token"),
		}

		//需要验证用户信息
		ok, vs := store.IsValid(apiHeader.UserId, apiHeader.Token)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("游客认证信息错误"))
			return
		}

		//set header
		apiHeader.UserName = vs.UName
		r = r.WithContext(context.WithValue(r.Context(), "header", apiHeader))

		//next
		next.ServeHTTP(w, r)
	})
}

func GetHeader(request *http.Request) *api.IdentityHeader {
	return request.Context().Value("header").(*api.IdentityHeader)
}
