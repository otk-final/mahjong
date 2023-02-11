package robot

import (
	"mahjong/server/api"
)

//智能

type mindLevel3 struct {
	*minder
	level1 *mindLevel1
	level2 *mindLevel2
}

func (m *mindLevel3) Put(event *api.PutPayload) {
	m.level1.Put(event)
}

func (m *mindLevel3) Turn(event *api.TurnPayload, ok bool) {
	m.level2.Turn(event, ok)
}
