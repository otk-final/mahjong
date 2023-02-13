package robot

import (
	"log"
	"mahjong/server/api"
	"mahjong/service"
	"mahjong/service/engine"
	"mahjong/service/ploy"
	"mahjong/service/store"
	"sort"
	"time"
)

type task struct {
	consumer   engine.NotifyHandle
	webEvent   api.WebEvent
	webPayload any
}

var eventAfterDelay = 3 * time.Second

var robotCh1 = make(chan *task, 3)
var robotCh2 = make(chan *task, 3)
var robotCh3 = make(chan *task, 3)

func init() {
	route := func(mind engine.NotifyHandle, webEvent api.WebEvent, webPayload any) {
		switch webEvent {
		case api.TakeEvent:
			mind.Take(webPayload.(*api.TakePayload))
			break
		case api.PutEvent:
			mind.Put(webPayload.(*api.PutPayload))
			break
		case api.RaceEvent:
			mind.Race(webPayload.(*api.RacePayload))
			break
		case api.WinEvent:
			mind.Win(webPayload.(*api.WinPayload))
			break
		case api.AckEvent:
			mind.Ack(webPayload.(*api.AckPayload))
			break
		case api.TurnEvent:
			mind.Turn(webPayload.(*api.TurnPayload), false)
			break
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
	l1 := &mindLevel1{minder: dm}
	l2 := &mindLevel2{minder: dm}
	l3 := &mindLevel3{minder: dm, level1: l1, level2: l2}
	if roboter.Level == 1 {
		robotCh1 <- &task{consumer: l1, webEvent: packet.Event, webPayload: packet.Payload}
	} else if roboter.Level == 2 {
		robotCh2 <- &task{consumer: l2, webEvent: packet.Event, webPayload: packet.Payload}
	} else if roboter.Level == 3 {
		robotCh3 <- &task{consumer: l3, webEvent: packet.Event, webPayload: packet.Payload}
	} else {
		log.Printf("roboter config illegals")
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

func (m *minder) Quit(reason string) {
}

func (m *minder) randomPut(ownIdx int) {

	provider := ploy.RenewProvider(m.roundCtx)
	//获取手牌
	ops := m.roundCtx.Operating()
	hands := ops.GetTiles(ownIdx).Hands
	sort.Ints(hands)

	//随机
	targetPut, _ := randomCanPut(ownIdx, hands, provider)
	log.Printf("机器人[%d] 开始随机出牌 %v", m.roboter.Idx, targetPut)
	time.AfterFunc(eventAfterDelay, func() {
		//出牌
		put := &api.PutPayload{Who: ownIdx, Tile: targetPut}
		service.DoPut(m.roundCtx, m.roboter.Player, &api.PutParameter{PutPayload: put, RoomId: m.roomId})
	})
}
