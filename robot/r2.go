package robot

import (
	"log"
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/service"
	"sort"
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
	if takeResult.Take == -1 {
		return
	}
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
		m.randomPut(m.roboter.Idx)
		return
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
		//递归继续处理
		m.doOptions(next.Options)
		return
	}
}
