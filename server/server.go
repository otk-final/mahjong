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

var player = &PlayerCtrl{}

func ApiRegister() *mux.Router {

	//卡牌事件
	muxRouter.Methods("POST").Path("/player/take").HandlerFunc(wrap.NewWrapper(player.take).Func())
	muxRouter.Methods("POST").Path("/player/put").HandlerFunc(wrap.NewWrapper(player.put).Func())
	muxRouter.Methods("POST").Path("/player/eat").HandlerFunc(wrap.NewWrapper(player.eat).Func())
	muxRouter.Methods("POST").Path("/player/pair").HandlerFunc(wrap.NewWrapper(player.pair).Func())
	muxRouter.Methods("POST").Path("/player/gang").HandlerFunc(wrap.NewWrapper(player.gang).Func())
	muxRouter.Methods("POST").Path("/player/win").HandlerFunc(wrap.NewWrapper(player.win).Func())
	muxRouter.Methods("POST").Path("/player/skip").HandlerFunc(wrap.NewWrapper(player.skip).Func())

	//长链接
	muxRouter.Handle("/ws/{roomId}", websocketUpgrader())
	return muxRouter
}

func websocketUpgrader() http.HandlerFunc {
	upgrader := websocket.Upgrader{
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
		conn, err := upgrader.Upgrade(writer, request, writer.Header())
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
			read:     make(chan api.WebPacket, 100),
			write:    make(chan api.WebPacket, 100),
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
				ty, msg, e := conn.ReadMessage()
				if e != nil {
					return
				}
				wsc.read <- api.WebPacket{Type: ty, Packet: msg}
			}

		}(conn, wsc)
		//写
		go func(conn *websocket.Conn, wsc *websocketChan) {
			for {
				select {
				case p, ok := <-wsc.write:
					if !ok {
						return
					}
					_ = conn.WriteMessage(p.Type, p.Packet)
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
	read     chan api.WebPacket
	write    chan api.WebPacket
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
			api.UnPacket(req)

			//TODO Handler
			fmt.Println(req)
		case <-time.After(5 * time.Second):
			//TODO
		}
	}

}

//回执
func websocketReply[T any](roomId string, targetId string, payload T) {
	chKey := fmt.Sprintf("%s#%s", roomId, targetId)
	temp, ok := memoryChannelMap.Load(chKey)
	if !ok {
		return
	}
	//序列化 json
	data, _ := json.Marshal(&payload)
	temp.(*websocketChan).write <- api.WebPacket{Type: websocket.TextMessage, Packet: data}
}
