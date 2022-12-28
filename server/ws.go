package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"mahjong/server/api"
	"net/http"
	"sync"
	"time"
)

func wsRoute() http.HandlerFunc {
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
		roomId := mux.Vars(request)["RoomId"]
		conn, err := wu.Upgrade(writer, request, writer.Header())
		if err != nil {
			return
		}
		//远程设备信息
		conn.LocalAddr()
		conn.RemoteAddr()

		//缓存
		wsc := &netChan{
			roomId: roomId,
			identity: &api.IdentityHeader{
				UserId: request.Header.Get("userId"),
				Token:  request.Header.Get("token"),
			},
			read:  make(chan []byte, 100),
			write: make(chan []byte, 100),
		}
		_, _ = netChanMap.LoadOrStore(wsc.Key(), wsc)

		//读
		go func(conn *websocket.Conn, wsc *netChan) {
			//释放
			defer func() {
				_ = conn.Close()
				wsc.Close()
				netChanMap.Delete(wsc.Key())
			}()
			for {
				tp, msg, e := conn.ReadMessage()
				if e != nil || tp != websocket.TextMessage {
					return
				}
				//TODO read
				//wsc.read <- msg
				fmt.Println(string(msg))
			}
		}(conn, wsc)

		//写
		go func(conn *websocket.Conn, wsc *netChan) {
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
	}
}

//全局缓存
var netChanMap = &sync.Map{}

type netChan struct {
	roomId   string
	identity *api.IdentityHeader
	read     chan []byte
	write    chan []byte
}

func (wsc *netChan) Close() {
	close(wsc.read)
	close(wsc.write)
}

func (wsc *netChan) Key() string {
	return fmt.Sprintf("%s#%s", wsc.roomId, wsc.identity.UserId)
}
