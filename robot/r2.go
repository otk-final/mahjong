package robot

import (
	"mahjong/server/api"
)

type mindLevel2 struct {
	roomId  string
	roboter *api.Roboter
}

func (m *mindLevel2) Take(event *api.TakePayload) {
	//TODO implement me
	panic("implement me")
}

func (m *mindLevel2) Put(event *api.PutPayload) {
	//TODO implement me
	panic("implement me")
}

func (m *mindLevel2) Race(event *api.RacePayload) {
	//TODO implement me
	panic("implement me")
}

func (m *mindLevel2) Win(event *api.WinPayload) {
	//TODO implement me
	panic("implement me")
}

func (m *mindLevel2) Ack(event *api.AckPayload) {
	//TODO implement me
	panic("implement me")
}

func (m *mindLevel2) Turn(event *api.TurnPayload, ok bool) {
	//TODO implement me
	panic("implement me")
}

func (m *mindLevel2) Quit(ok bool) {
	//TODO implement me
	panic("implement me")
}
