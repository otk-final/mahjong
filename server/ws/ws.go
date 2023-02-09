package ws

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"mahjong/server/api"
	"net/http"
	"net/url"
	"sync"
	"time"
)

func Route() http.HandlerFunc {

	//创建链接
	return func(writer http.ResponseWriter, request *http.Request) {
		//获取认证信息
		subProtocolsHeaders := websocket.Subprotocols(request)
		wu := websocket.Upgrader{
			HandshakeTimeout: 10 * time.Second,
			ReadBufferSize:   1024 * 2,
			WriteBufferSize:  1024 * 2,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
				log.Println(reason)
			},
			Subprotocols: subProtocolsHeaders,
		}

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
		//获取认证信息
		var identity *api.IdentityHeader
		if len(subProtocolsHeaders) > 0 {
			un, _ := url.QueryUnescape(subProtocolsHeaders[1])
			identity = &api.IdentityHeader{
				UserId:   subProtocolsHeaders[0],
				UserName: un,
				Token:    subProtocolsHeaders[2],
			}
		} else {
			identity = &api.IdentityHeader{
				UserId:   request.Header.Get("userId"),
				UserName: request.Header.Get("userName"),
				Token:    request.Header.Get("token"),
			}
		}

		//缓存
		wsc := &netChan{
			roomId:   roomId,
			identity: identity,
			read:     make(chan []byte, 100),
			write:    make(chan []byte, 100),
		}
		_, _ = netChanMap.LoadOrStore(wsc.Key(), wsc)

		//远程设备信息
		log.Printf("玩家：%s 连接[%s]成功 Ip:%s\n", wsc.identity.UserName, roomId, conn.RemoteAddr().String())

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

func PostMessage(roomId, acctId string, data []byte) {
	chKey := fmt.Sprintf("%s#%s", roomId, acctId)
	temp, ok := netChanMap.Load(chKey)
	if !ok {
		return
	}
	temp.(*netChan).write <- data
}
