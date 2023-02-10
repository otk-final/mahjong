package robot

import (
	"mahjong/server/api"
	"mahjong/service/engine"
	"mahjong/service/store"
	"time"
)

type event struct {
	roomId     string
	roundCtx   *engine.RoundCtx
	roboter    *api.Roboter
	webEvent   api.WebEvent
	webPayload any
}

var robotCh1 = make(chan *event, 4)
var robotCh2 = make(chan *event, 4)
var robotCh3 = make(chan *event, 4)

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
				route(&mindLevel1{roomId: e.roomId, roboter: e.roboter, roundCtx: e.roundCtx}, e.webEvent, e.webPayload)
			case e := <-robotCh2:
				route(&mindLevel2{roomId: e.roomId, roboter: e.roboter, roundCtx: e.roundCtx}, e.webEvent, e.webPayload)
			case e := <-robotCh3:
				route(&mindLevel3{roomId: e.roomId, roboter: e.roboter, roundCtx: e.roundCtx}, e.webEvent, e.webPayload)
			case <-time.After(5 * time.Second):
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

	e := &event{roomId: roomId, roundCtx: roundCtx, roboter: roboter, webEvent: packet.Event, webPayload: packet.Payload}
	if roboter.Level == 1 {
		robotCh1 <- e
	} else if roboter.Level == 2 {
		robotCh2 <- e
	} else if roboter.Level == 3 {
		robotCh3 <- e
	}
}
