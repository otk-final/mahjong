package ploy

import (
	"log"
	"mahjong/mj"
	"testing"
)

func TestLaiWin(t *testing.T) {
	el := &winLai{
		lai:           mj.W2,
		canChong:      false,
		unique:        false,
		winEvaluation: winEvaluation{},
	}
	tiles := mj.Cards{
		mj.W1, mj.W3, mj.W2, mj.W2, mj.L1,
	}
	ok, com := el.multiLaiCheck(tiles)
	if ok {
		log.Println(com)
	}
	log.Println(ok)
}
