package robot

import (
	"mahjong/server/api"
	"mahjong/service"
	"mahjong/service/engine"
	"time"
)

//最初级 只摸牌，从不判定，然后随机打牌

type mindLevel1 struct {
	roundCtx *engine.RoundCtx
	roomId   string
	roboter  *api.Roboter
}

func (m *mindLevel1) Take(event *api.TakePayload) {
}

func (m *mindLevel1) Put(event *api.PutPayload) {
	if event.Who == m.roboter.Idx {
		return
	}
	//直接忽略
	ackId := m.roundCtx.Exchange().CurrentAckId()
	m.roundCtx.Exchange().PostAck(&api.AckPayload{Who: m.roboter.Idx, Round: 0, AckId: ackId})
}

func (m *mindLevel1) Race(event *api.RacePayload) {
}

func (m *mindLevel1) Win(event *api.WinPayload) {
}

func (m *mindLevel1) Ack(event *api.AckPayload) {
}

func (m *mindLevel1) Turn(event *api.TurnPayload, ok bool) {
	//摸牌
	takeResult := service.DoTake(m.roundCtx, m.roboter.Player, &api.TakeParameter{
		RoomId:    m.roomId,
		Round:     0,
		Direction: 1,
	})
	//摸什么打什么
	put := &api.PutPayload{Who: m.roboter.Idx, Round: 0, Tile: takeResult.Take}
	//延迟两秒出牌
	time.AfterFunc(3*time.Second, func() {
		service.DoPut(m.roundCtx, m.roboter.Player, &api.PutParameter{
			PutPayload: put,
			RoomId:     m.roomId,
		})
	})
}

func (m *mindLevel1) Quit(ok bool) {
}
