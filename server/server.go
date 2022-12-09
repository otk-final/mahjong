package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"mahjong/server/api"
	"mahjong/server/wrap"
	"net/http"
	"sync"
	"time"
)

var muxRouter = mux.NewRouter()

var playCtrl = &PlayerCtrl{}
var gameCtrl = &GameCtrl{}

func ApiRegister() *mux.Router {

	//游戏
	muxRouter.Methods("POST").Path("/game/start").HandlerFunc(wrap.NewWrapper(gameCtrl.start).Func())
	muxRouter.Methods("POST").Path("/game/ack").HandlerFunc(wrap.NewWrapper(gameCtrl.ack).Func())

	//卡牌事件
	muxRouter.Methods("POST").Path("/playCtrl/take").HandlerFunc(wrap.NewWrapper(playCtrl.take).Func())
	muxRouter.Methods("POST").Path("/playCtrl/put").HandlerFunc(wrap.NewWrapper(playCtrl.put).Func())
	muxRouter.Methods("POST").Path("/playCtrl/reward").HandlerFunc(wrap.NewWrapper(playCtrl.reward).Func())
	muxRouter.Methods("POST").Path("/playCtrl/win").HandlerFunc(wrap.NewWrapper(playCtrl.win).Func())
	muxRouter.Methods("POST").Path("/playCtrl/skip").HandlerFunc(wrap.NewWrapper(playCtrl.skip).Func())

	//长链接
	muxRouter.Handle("/ws/{roomId}", websocketRoute())
	return muxRouter
}

func websocketRoute() http.HandlerFunc {
	wu := websocket.Upgrader{
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		Error: nil,
	}
	//创建链接
	return func(writer http.ResponseWriter, request *http.Request) {
		//非websocket请求
		if !websocket.IsWebSocketUpgrade(request) {
			return
		}
		//判断房间是否存在
		roomId := mux.Vars(request)["roomId"]
		conn, err := wu.Upgrade(writer, request, writer.Header())
		if err != nil {
			return
		}
		//远程设备信息
		conn.LocalAddr()
		conn.RemoteAddr()
		identity := &api.Identity{
			UserId:   roomId,
			Token:    request.Header.Get("token"),
			UserName: request.Header.Get("user_name"),
		}

		//缓存
		wsc := &websocketChan{
			roomId:   roomId,
			identity: identity,
			read:     make(chan []byte, 100),
			write:    make(chan []byte, 100),
		}
		_, _ = memoryChannelMap.LoadOrStore(wsc.Key(), wsc)

		//读
		go func(conn *websocket.Conn, wsc *websocketChan) {
			//释放
			defer func() {
				_ = conn.Close()
				wsc.Close()
				memoryChannelMap.Delete(wsc.Key())
			}()
			for {
				tp, msg, e := conn.ReadMessage()
				if e != nil || tp != websocket.TextMessage {
					return
				}
				wsc.read <- msg
			}
		}(conn, wsc)

		//写
		go func(conn *websocket.Conn, wsc *websocketChan) {
			for {
				select {
				case data, ok := <-wsc.write:
					if !ok {
						return
					}
					_ = conn.WriteMessage(websocket.TextMessage, data)
				case <-time.After(5 * time.Second):
					//心跳
					_ = conn.WriteMessage(websocket.PingMessage, []byte("health"))
				}
			}

		}(conn, wsc)

		//handler
		go websocketHandler(wsc)
	}
}

var memoryChannelMap = &sync.Map{}

type websocketChan struct {
	roomId   string
	identity *api.Identity
	read     chan []byte
	write    chan []byte
}

func (wsc *websocketChan) Close() {
	close(wsc.read)
	close(wsc.write)
}
func (wsc *websocketChan) Key() string {
	return fmt.Sprintf("%s#%s", wsc.roomId, wsc.identity.UserId)
}

func websocketHandler(wsc *websocketChan) {
	for {
		select {
		case req, ok := <-wsc.read:
			if !ok {
				return
			}
			//解包
			event, payload, err := api.UnPacket[map[string]interface{}](req)
			if err != nil {
				continue
			}
			fmt.Println(payload)
			//TODO 路由
			switch event {
			case 100:

			case 101:
			case 102:
			case 103:

			}
		case <-time.After(5 * time.Second):
		}
	}

}

//回执
func websocketNotify[T any](roomId string, targetId string, event int, eventId string, payload T) {
	chKey := fmt.Sprintf("%s#%s", roomId, targetId)
	temp, ok := memoryChannelMap.Load(chKey)
	if !ok {
		return
	}
	//序列化 json
	msg := &api.WebPacket[T]{Event: event, EventId: eventId, Payload: payload}
	packet, _ := json.Marshal(msg)

	temp.(*websocketChan).write <- packet
}
