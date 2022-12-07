package scene

import "mahjong/mj"

// 癞子

type MutableCards struct {
	mj.Cards
	mutable int
}

func NewMutableCards() *MutableCards {

	return nil
}
