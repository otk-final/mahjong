package robot

import (
	"errors"
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/service"
	"mahjong/service/ploy"
	"math/rand"
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
	service.DoIgnore(m.roundCtx, m.roboter.Player)
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
	if event.Who != m.roboter.Idx {
		return
	}
	//摸牌
	takeResult := service.DoTake(m.roundCtx, m.roboter.Player, &api.TakeParameter{RoomId: m.roomId, Direction: 1})
	if takeResult.Take == -1 {
		return
	}
	//出牌
	m.randomPut(m.roboter.Idx)

}
