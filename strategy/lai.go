package strategy

import (
	"mahjong/mj"
	"mahjong/server/api"
)

// 癞子

type LaiOrderHandler struct {
	Order OrderHandler
}

func (h LaiOrderHandler) When() Turn {
	return Other
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

func (h LaiPairHandler) When() Turn {
	return Other
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

func (h LaiChaoHandler) When() Turn {
	return Both
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

// LaiFixGangHandler 癞(🀄️)
type LaiFixGangHandler struct {
	Gang GangHandler
}

func (h LaiFixGangHandler) When() Turn {
	return Mine
}

func (h LaiFixGangHandler) Named() (string, string) {
	return "hzGang", "癞杠"
}

func (h LaiFixGangHandler) Func(configure *api.GameConfigure, tb *mj.Table) DeterFunc {
	gf := h.Gang.Func(configure, tb)
	return func(withPlayer *mj.Player, targetPlayer *mj.Player, withCards []int, targetCard int) bool {
		if targetCard == 1 {
			return false
		}
		return gf(withPlayer, targetPlayer, withCards, targetCard)
	}
}

// LaiCardGangHandler 癞杠
type LaiCardGangHandler struct {
	Gang GangHandler
}

func (h LaiCardGangHandler) When() Turn {
	return Mine
}

func (h LaiCardGangHandler) Named() (string, string) {
	return "laiGang", "癞杠"
}

func (h LaiCardGangHandler) Func(configure *api.GameConfigure, tb *mj.Table) DeterFunc {
	return func(withPlayer *mj.Player, targetPlayer *mj.Player, withCards []int, targetCard int) bool {
		return false
	}
}

// LaiGangHandler 杠
type LaiGangHandler struct {
	Gang GangHandler
}

func (h LaiGangHandler) When() Turn {
	return Both
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
	HuPai HuPaiHandler
}

func (h LaiHuPaiHandler) When() Turn {
	return Both
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
