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

func (hs DeterHandlers) Filter(turnType Turn) []DeterHandler {
	filter := make([]DeterHandler, 0)
	for _, h := range hs {
		if h.When() == turnType || h.When() == Both {
			filter = append(filter, h)
		}
	}
	return filter
}

func Register(mode string) (DeterHandlers, error) {

	stdOrder := &OrderHandler{}
	stdGang := &GangHandler{}
	stdPair := &PairHandler{}
	stdHu := &HuPaiHandler{}

	//普通麻将
	if strings.EqualFold(mode, "std") {
		return []DeterHandler{stdOrder, stdGang, stdPair}, nil
	}

	//癞子
	if strings.EqualFold(mode, "lai") {
		lc := &LaiChaoHandler{Pair: *stdPair}
		lp := &LaiPairHandler{Pair: *stdPair}
		lg := &LaiGangHandler{Gang: *stdGang}
		fg := &LaiFixGangHandler{Gang: *stdGang}
		cg := &LaiCardGangHandler{Gang: *stdGang}
		lh := &LaiHuPaiHandler{HuPai: *stdHu}
		return []DeterHandler{lc, lp, lg, fg, cg, lh}, nil
	}

	//卡五星
	if strings.EqualFold(mode, "k5x") {
		//TODO
		return nil, nil
	}
	return nil, errors.New("not support")
}
