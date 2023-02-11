package robot

import (
	"errors"
	"log"
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/service"
	"mahjong/service/ploy"
	"math/rand"
	"time"
)

//最初级，从不判定，摸什么出什么

type mindLevel1 struct {
	*minder
}

func (m *mindLevel1) Put(event *api.PutPayload) {
	if event.Who == m.roboter.Idx {
		return
	}
	//直接忽略
	ackId := m.roundCtx.Exchange().CurrentAckId()
	m.roundCtx.Exchange().PostAck(&api.AckPayload{Who: m.roboter.Idx, AckId: ackId})
}

func randomCanPut(pIdx int, hands mj.Cards, provider ploy.GameDefine) (int, error) {
	cans := make([]int, 0)
	for _, t := range hands {
		if provider.CanPut(pIdx, t) {
			cans = append(cans, t)
		}
	}
	cl := len(cans)
	if cl == 0 {
		return -1, errors.New("no can put")
	}
	return cans[rand.Intn(cl)], nil
}

func (m *mindLevel1) Turn(event *api.TurnPayload, ok bool) {
	provider := ploy.RenewProvider(m.roundCtx)
	//摸牌
	takeResult := service.DoTake(m.roundCtx, m.roboter.Player, &api.TakeParameter{
		RoomId:    m.roomId,
		Direction: 1,
	})
	ownIdx := event.Who

	//不能出，则从手牌中随机选择
	targetPut := takeResult.Take
	var err error
	if !provider.CanPut(ownIdx, targetPut) {
		hands := m.roundCtx.Operating().GetTiles(ownIdx).Hands
		targetPut, err = randomCanPut(ownIdx, hands, provider)
		if err != nil {
			log.Printf("错误：%v", hands)
			return
		}
	}

	put := &api.PutPayload{Who: m.roboter.Idx, Tile: targetPut}
	//延迟3秒出牌
	time.AfterFunc(eventAfterDelay, func() {
		service.DoPut(m.roundCtx, m.roboter.Player, &api.PutParameter{
			PutPayload: put,
			RoomId:     m.roomId,
		})
	})
}
