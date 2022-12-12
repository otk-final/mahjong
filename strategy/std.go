package strategy

import (
	"mahjong/mj"
	"mahjong/server/api"
)

// OrderHandler 吃
type OrderHandler struct {
}

func (h OrderHandler) Named() (string, string) {
	return "order", "吃"
}

func (h OrderHandler) Func(configure *api.GameConfigure, tb *mj.Table) DeterFunc {
	return func(withPlayer *mj.Player, targetPlayer *mj.Player, withCards []int, targetCard int) bool {
		return false
	}
}

// PairHandler 碰
type PairHandler struct {
}

func (h PairHandler) Named() (string, string) {
	return "pair", "碰"
}

func (h PairHandler) Func(configure *api.GameConfigure, tb *mj.Table) DeterFunc {
	return func(withPlayer *mj.Player, targetPlayer *mj.Player, withCards []int, targetCard int) bool {
		return false
	}
}

// GangHandler 杠
type GangHandler struct {
}

func (h GangHandler) Named() (string, string) {
	return "gang", "杠"
}

func (h GangHandler) Func(configure *api.GameConfigure, tb *mj.Table) DeterFunc {
	return func(withPlayer *mj.Player, targetPlayer *mj.Player, withCards []int, targetCard int) bool {
		return false
	}
}

// HuPaiHandler 胡牌
type HuPaiHandler struct {
}

func (h HuPaiHandler) Named() (string, string) {
	return "hu", "胡牌"
}

func (h HuPaiHandler) Func(configure *api.GameConfigure, tb *mj.Table) DeterFunc {
	return func(withPlayer *mj.Player, targetPlayer *mj.Player, withCards []int, targetCard int) bool {
		return false
	}
}
