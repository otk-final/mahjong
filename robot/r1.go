package robot

import (
	"errors"
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/service"
	"mahjong/service/engine"
	"mahjong/service/ploy"
	"time"
)

//最初级，从不判定，摸什么出什么

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
	m.roundCtx.Exchange().PostAck(&api.AckPayload{Who: m.roboter.Idx, AckId: ackId})
}

func (m *mindLevel1) Race(event *api.RacePayload) {
}

func (m *mindLevel1) Win(event *api.WinPayload) {
}

func (m *mindLevel1) Ack(event *api.AckPayload) {
}

func randomPutWithHand(pIdx int, hands mj.Cards, provider ploy.GameDefine) (int, error) {
	if len(hands) == 0 {
		return 0, errors.New("hands is empty")
	}
	for _, t := range hands {
		if provider.CanPut(pIdx, t) {
			return t, nil
		}
	}
	return 0, errors.New("no can put")
}

func (m *mindLevel1) Turn(event *api.TurnPayload, ok bool) {
	provider := ploy.RenewProvider(m.roundCtx)
	//摸牌
	takeResult := service.DoTake(m.roundCtx, m.roboter.Player, &api.TakeParameter{
		RoomId:    m.roomId,
		Direction: 1,
	})
	ownIdx := m.roboter.Idx

	//不能出，则从手牌中随机选择
	targetPut := takeResult.Take
	var err error
	if !provider.CanPut(ownIdx, targetPut) {
		hands := m.roundCtx.Operating().GetTiles(ownIdx).Hands
		targetPut, err = randomPutWithHand(ownIdx, hands, provider)
	}
	if err != nil {
		return
	}

	put := &api.PutPayload{Who: m.roboter.Idx, Tile: targetPut}
	//延迟3秒出牌
	time.AfterFunc(3*time.Second, func() {
		service.DoPut(m.roundCtx, m.roboter.Player, &api.PutParameter{
			PutPayload: put,
			RoomId:     m.roomId,
		})
	})
}

func (m *mindLevel1) Quit(ok bool) {
}
