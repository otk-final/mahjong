package robot

import (
	"mahjong/server/api"
	"mahjong/service/engine"
)

//智能

type mindLevel3 struct {
	roomId   string
	roboter  *api.Roboter
	roundCtx *engine.RoundCtx
}

func (m *mindLevel3) Take(event *api.TakePayload) {
	//TODO implement me
	panic("implement me")
}

func (m *mindLevel3) Put(event *api.PutPayload) {
	//TODO implement me
	panic("implement me")
}

func (m *mindLevel3) Race(event *api.RacePayload) {
	//TODO implement me
	panic("implement me")
}

func (m *mindLevel3) Win(event *api.WinPayload) {
	//TODO implement me
	panic("implement me")
}

func (m *mindLevel3) Ack(event *api.AckPayload) {
	//TODO implement me
	panic("implement me")
}

func (m *mindLevel3) Turn(event *api.TurnPayload, ok bool) {
	//TODO implement me
	panic("implement me")
}

func (m *mindLevel3) Quit(ok bool) {
	//TODO implement me
	panic("implement me")
}
