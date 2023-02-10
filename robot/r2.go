package robot

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/service"
	"mahjong/service/engine"
)

//中级，优先出废牌，能胡则胡

type mindLevel2 struct {
	mindLevel1
	roomId   string
	roboter  *api.Roboter
	roundCtx *engine.RoundCtx
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
	options := service.DoRacePre(m.roundCtx, m.roboter.Player, &api.RacePreview{RoomId: m.roomId})
	if len(options) == 0 {
		return
	}

	//能胡则胡，不能胡则忽略
	if win, _ := hasOption(options, api.WinRace); win {
		_, _ = service.DoWin(m.roundCtx, m.roboter.Player)
	} else {
		service.DoIgnore(m.roundCtx, m.roboter.Player)
	}
}

func (m *mindLevel2) Race(event *api.RacePayload) {

}

func (m *mindLevel2) Turn(event *api.TurnPayload, ok bool) {
	//摸牌
	takeResult := service.DoTake(m.roundCtx, m.roboter.Player, &api.TakeParameter{
		RoomId:    m.roomId,
		Direction: 1,
	})
	m.doRaceAndPut(takeResult.Options, takeResult.Take)
}

func (m *mindLevel2) doRaceAndPut(options []*api.RaceOption, target int) {

}
