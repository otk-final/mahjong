package robot

import (
	"mahjong/server/api"
	"mahjong/service/engine"
	"mahjong/service/store"
	"time"
)

type task struct {
	consumer   engine.NotifyHandle
	webEvent   api.WebEvent
	webPayload any
}

var eventAfterDelay = 3 * time.Second

var robotCh1 = make(chan *task, 4)
var robotCh2 = make(chan *task, 4)
var robotCh3 = make(chan *task, 4)

func init() {
	route := func(mind engine.NotifyHandle, webEvent api.WebEvent, webPayload any) {
		switch webEvent {
		case api.TakeEvent:
			mind.Take(webPayload.(*api.TakePayload))
		case api.PutEvent:
			mind.Put(webPayload.(*api.PutPayload))
		case api.RaceEvent:
			mind.Race(webPayload.(*api.RacePayload))
		case api.WinEvent:
			mind.Win(webPayload.(*api.WinPayload))
		case api.AckEvent:
			mind.Ack(webPayload.(*api.AckPayload))
		case api.TurnEvent:
			mind.Turn(webPayload.(*api.TurnPayload), false)
		}
	}
	//异步处理
	go func() {
		for {
			select {
			case e := <-robotCh1:
				route(e.consumer, e.webEvent, e.webPayload)
			case e := <-robotCh2:
				route(e.consumer, e.webEvent, e.webPayload)
			case e := <-robotCh3:
				route(e.consumer, e.webEvent, e.webPayload)
			case <-time.After(5 * time.Second):
				//wait
			}
		}
	}()
}

func Post[T any](roomId string, roboter *api.Roboter, packet *api.WebPacket[T]) {

	//查询上下文
	roundCtx, err := store.LoadRoundCtx(roomId, roboter.UId)
	if err != nil {
		return
	}

	//default
	dm := &minder{roundCtx: roundCtx, roomId: roomId, roboter: roboter}
	if roboter.Level == 1 {
		robotCh1 <- &task{consumer: &mindLevel2{minder: dm}, webEvent: packet.Event, webPayload: packet.Payload}
	} else if roboter.Level == 2 {
		robotCh1 <- &task{consumer: &mindLevel2{minder: dm}, webEvent: packet.Event, webPayload: packet.Payload}
	} else if roboter.Level == 3 {
		//l1 := &mindLevel1{minder: dm}
		//l2 := &mindLevel2{minder: dm}
		robotCh1 <- &task{consumer: &mindLevel2{minder: dm}, webEvent: packet.Event, webPayload: packet.Payload}
	}
}

type minder struct {
	roundCtx *engine.RoundCtx
	roomId   string
	roboter  *api.Roboter
}

func (m *minder) Take(event *api.TakePayload) {
}

func (m *minder) Put(event *api.PutPayload) {
}

func (m *minder) Race(event *api.RacePayload) {
}

func (m *minder) Win(event *api.WinPayload) {
}

func (m *minder) Ack(event *api.AckPayload) {
}

func (m *minder) Turn(event *api.TurnPayload, ok bool) {
}

func (m *minder) Quit(ok bool) {
}
