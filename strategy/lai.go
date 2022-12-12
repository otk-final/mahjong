package strategy

import (
	"mahjong/mj"
	"mahjong/server/api"
)

// 癞子

type LaiOrderHandler struct {
	Order OrderHandler
}

func (h LaiOrderHandler) Named() (string, string) {
	return "order", "吃"
}

func (h LaiOrderHandler) Func(configure *api.GameConfigure, tb *mj.Table) DeterFunc {
	return func(withPlayer *mj.Player, targetPlayer *mj.Player, withCards []int, targetCard int) bool {
		return false
	}
}

// LaiPairHandler 碰
type LaiPairHandler struct {
	Pair PairHandler
}

func (h LaiPairHandler) Named() (string, string) {
	return "pair", "碰"
}

func (h LaiPairHandler) Func(configure *api.GameConfigure, tb *mj.Table) DeterFunc {
	return func(withPlayer *mj.Player, targetPlayer *mj.Player, withCards []int, targetCard int) bool {
		return false
	}
}

// LaiChaoHandler 朝天
type LaiChaoHandler struct {
	Pair PairHandler
}

func (h LaiChaoHandler) Named() (string, string) {
	return "chao", "朝天"
}

func (h LaiChaoHandler) Func(configure *api.GameConfigure, tb *mj.Table) DeterFunc {
	pf := h.Pair.Func(configure, tb)
	return func(withPlayer *mj.Player, targetPlayer *mj.Player, withCards []int, targetCard int) bool {
		if targetCard == 1 {
			return false
		}
		return pf(withPlayer, targetPlayer, withCards, targetCard)
	}
}

// LaiHZHandler 癞(🀄️)
type LaiHZHandler struct {
	Gang GangHandler
}

func (h LaiHZHandler) Named() (string, string) {
	return "hzGang", "癞杠"
}

func (h LaiHZHandler) Func(configure *api.GameConfigure, tb *mj.Table) DeterFunc {
	gf := h.Gang.Func(configure, tb)
	return func(withPlayer *mj.Player, targetPlayer *mj.Player, withCards []int, targetCard int) bool {
		if targetCard == 1 {
			return false
		}
		return gf(withPlayer, targetPlayer, withCards, targetCard)
	}
}

// LaiCardHandler 癞杠
type LaiCardHandler struct {
	Gang GangHandler
}

func (h LaiCardHandler) Named() (string, string) {
	return "laiGang", "癞杠"
}

func (h LaiCardHandler) Func(configure *api.GameConfigure, tb *mj.Table) DeterFunc {
	return func(withPlayer *mj.Player, targetPlayer *mj.Player, withCards []int, targetCard int) bool {
		return false
	}
}

// LaiGangHandler 杠
type LaiGangHandler struct {
	Gang GangHandler
}

func (h LaiGangHandler) Named() (string, string) {
	return "gang", "杠"
}

func (h LaiGangHandler) Func(configure *api.GameConfigure, tb *mj.Table) DeterFunc {
	gf := h.Gang.Func(configure, tb)
	return func(withPlayer *mj.Player, targetPlayer *mj.Player, withCards []int, targetCard int) bool {
		if targetCard == 1 {
			return false
		}
		return gf(withPlayer, targetPlayer, withCards, targetCard)
	}
}

type LaiHuPaiHandler struct {
	h HuPaiHandler
}

func (h LaiHuPaiHandler) Named() (string, string) {
	return "hu", "胡牌"
}

func (h LaiHuPaiHandler) Func(configure *api.GameConfigure, tb *mj.Table) DeterFunc {
	return func(withPlayer *mj.Player, targetPlayer *mj.Player, withCards []int, targetCard int) bool {
		//判定当前牌中是否有癞子
		return true
	}
}
