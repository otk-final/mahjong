package robot

import (
	"mahjong/server/api"
	"mahjong/service/engine"
	"time"
)

type event struct {
	roomId  string
	roboter *api.Roboter
	event   api.WebEvent
	payload any
}

var robotCh1 = make(chan *event, 0)
var robotCh2 = make(chan *event, 0)
var robotCh3 = make(chan *event, 0)

func init() {
	route := func(mind engine.NotifyHandle, event api.WebEvent, payload any) {
		switch event {
		case api.TakeEvent:
			mind.Take(payload.(*api.TakePayload))
		case api.PutEvent:
			mind.Put(payload.(*api.PutPayload))
		case api.RaceEvent:
			mind.Race(payload.(*api.RacePayload))
		case api.WinEvent:
			mind.Win(payload.(*api.WinPayload))
		case api.AckEvent:
			mind.Ack(payload.(*api.AckPayload))
		case api.TurnEvent:
			mind.Turn(payload.(*api.TurnPayload), false)
		}
	}
	//异步处理
	go func() {
		for {
			select {
			case e := <-robotCh1:
				route(&mindLevel1{roomId: e.roomId, roboter: e.roboter}, e.event, e.payload)
			case e := <-robotCh2:
				route(&mindLevel2{roomId: e.roomId, roboter: e.roboter}, e.event, e.payload)
			case e := <-robotCh3:
				route(&mindLevel3{roomId: e.roomId, roboter: e.roboter}, e.event, e.payload)
			case <-time.After(5 * time.Second):
			}
		}
	}()
}

func Post[T any](roomId string, roboter *api.Roboter, packet *api.WebPacket[T]) {
	e := &event{roomId: roomId, roboter: roboter, event: packet.Event, payload: packet.Payload}
	if roboter.Level == 1 {
		robotCh1 <- e
	} else if roboter.Level == 2 {
		robotCh2 <- e
	} else if roboter.Level == 3 {
		robotCh3 <- e
	}
}
