package robot

import (
	"log"
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/service"
	"mahjong/service/ploy"
	"sort"
	"time"
)

//中级，优先出废牌，能胡则胡

type mindLevel2 struct {
	*minder
}

func hasOption(options []*api.RaceOption, raceType api.RaceType) (bool, []mj.Cards) {
	for _, r := range options {
		if r.RaceType == raceType {
			return true, r.Tiles
		}
	}
	return false, nil
}

func (m *mindLevel2) Put(event *api.PutPayload) {
	if event.Who == m.roboter.Idx {
		return
	}
	//判定
	options := service.DoRacePre(m.roundCtx, m.roboter.Player, &api.RacePreview{RoomId: m.roomId, Target: event.Who, Tile: event.Tile})
	m.doOptions(options)
}

func (m *mindLevel2) Turn(event *api.TurnPayload, ok bool) {
	if event.Who != m.roboter.Idx {
		return
	}

	//摸牌
	takeResult := service.DoTake(m.roundCtx, m.roboter.Player, &api.TakeParameter{RoomId: m.roomId, Direction: 1})
	log.Printf("机器人[%d] 开始摸牌 %v", m.roboter.Idx, takeResult.Take)

	//出牌
	m.doOptions(takeResult.Options)
}

func (m *mindLevel2) doOptions(options []*api.RaceOption) {
	if len(options) == 0 {
		return
	}

	//能判定则判定 胡 > 碰 > 吃
	if win, _ := hasOption(options, api.WinRace); win {
		//胡牌
		_, _ = service.DoWin(m.roundCtx, m.roboter.Player)
	} else if put, _ := hasOption(options, api.PutRace); put && len(options) == 1 {
		//出牌
		m.optimizePut(m.roboter.Idx)
	} else {

		//当前事件
		sort.Slice(options, func(i, j int) bool {
			return options[i].RaceType > options[j].RaceType
		})
		raceOps := options[0]
		log.Printf("动作：%s 推荐：%v", api.RaceNames[raceOps.RaceType], raceOps.Tiles)
		raceArg := &api.RaceParameter{RoomId: m.roomId, RaceType: raceOps.RaceType, Tiles: raceOps.Tiles[0]}

		//下个事件
		next, err := service.DoRace(m.roundCtx, m.roboter.Player, raceArg)
		if err != nil {
			log.Printf("错误")
			return
		}

		log.Printf("下一次动作：%s 推荐：%v ", m.roboter.UName, next.Options)

		//递归继续处理
		m.doOptions(next.Options)
	}
}

func (m *mindLevel2) optimizePut(ownIdx int) {

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
