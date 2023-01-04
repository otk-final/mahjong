package ploy

import (
	"log"
	"mahjong/mj"
	"testing"
)

func TestLaiWin(t *testing.T) {
	el := &winLai{
		lai:           mj.W1,
		canChong:      false,
		unique:        false,
		winEvaluation: winEvaluation{},
	}
	tiles := mj.Cards{
		mj.W1, mj.W1, mj.T9, mj.L1, mj.L1,
	}
	ok, com := el.multiLaiCheck(tiles)
	if ok {
		log.Println(com)
	}
	log.Println(ok)
}
