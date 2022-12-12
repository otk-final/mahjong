package strategy

import (
	"errors"
	"mahjong/mj"
	"mahjong/server/api"
	"strings"
)

type HandlerRegister interface {
}

type Turn int

const (
	Mine Turn = iota
	Other
	Both
)

type DeterFunc func(withPlayer *mj.Player, targetPlayer *mj.Player, withCards []int, targetCard int) bool
type DeterHandler interface {
	When() Turn
	Named() (string, string)
	Func(configure *api.GameConfigure, tb *mj.Table) DeterFunc
}

type DeterHandlers []DeterHandler

func (hs DeterHandlers) Handler(action string) (DeterHandler, error) {
	for _, h := range hs {
		n, _ := h.Named()
		if strings.EqualFold(n, action) {
			return h, nil
		}
	}
	return nil, errors.New("not found")
}

func Register(mode string) (HandlerRegister, error) {

	//普通麻将
	if strings.EqualFold(mode, "std") {

	}

	//癞子
	if strings.EqualFold(mode, "lai") {

	}

	//卡五星
	if strings.EqualFold(mode, "k5x") {

	}
	return nil, errors.New("not support")
}
