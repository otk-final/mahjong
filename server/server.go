package server

import (
	"github.com/gorilla/mux"
	"mahjong/server/wrap"
)

var muxRouter = mux.NewRouter()

var playCtrl = &PlayerCtrl{}
var gameCtrl = &GameCtrl{}

func ApiRegister() *mux.Router {

	//游戏
	muxRouter.Methods("POST").Path("/game/start").HandlerFunc(wrap.NewWrapper(gameCtrl.start).Func())
	muxRouter.Methods("POST").Path("/game/startReady").HandlerFunc(wrap.NewWrapper(gameCtrl.startReady).Func())

	//卡牌事件
	muxRouter.Methods("POST").Path("/playCtrl/take").HandlerFunc(wrap.NewWrapper(playCtrl.take).Func())
	muxRouter.Methods("POST").Path("/playCtrl/put").HandlerFunc(wrap.NewWrapper(playCtrl.put).Func())
	muxRouter.Methods("POST").Path("/playCtrl/reward").HandlerFunc(wrap.NewWrapper(playCtrl.reward).Func())
	muxRouter.Methods("POST").Path("/playCtrl/win").HandlerFunc(wrap.NewWrapper(playCtrl.win).Func())
	muxRouter.Methods("POST").Path("/playCtrl/skip").HandlerFunc(wrap.NewWrapper(playCtrl.skip).Func())

	//长链接
	muxRouter.Handle("/ws/{roomId}", wsRoute())
	return muxRouter
}
