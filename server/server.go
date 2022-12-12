package server

import (
	"github.com/gorilla/mux"
	"mahjong/server/wrap"
)

var muxRouter = mux.NewRouter()

func ApiRegister() *mux.Router {

	//房间

	//游戏
	muxRouter.Methods("POST").Path("/game/start").HandlerFunc(wrap.NewWrapper(start).Func())
	muxRouter.Methods("POST").Path("/game/startAck").HandlerFunc(wrap.NewWrapper(startAck).Func())
	muxRouter.Methods("POST").Path("/game/startLoad").HandlerFunc(wrap.NewWrapper(startLoad).Func())

	//卡牌事件
	muxRouter.Methods("POST").Path("/play/take").HandlerFunc(wrap.NewWrapper(take).Func())
	muxRouter.Methods("POST").Path("/play/put").HandlerFunc(wrap.NewWrapper(put).Func())
	muxRouter.Methods("POST").Path("/play/rewardCheck").HandlerFunc(wrap.NewWrapper(rewardCheck).Func())
	muxRouter.Methods("POST").Path("/play/rewardConfirm").HandlerFunc(wrap.NewWrapper(rewardConfirm).Func())
	muxRouter.Methods("POST").Path("/play/win").HandlerFunc(wrap.NewWrapper(win).Func())
	muxRouter.Methods("POST").Path("/play/skip").HandlerFunc(wrap.NewWrapper(skip).Func())

	//长链接
	muxRouter.Handle("/ws/{roomId}", wsRoute())
	return muxRouter
}
