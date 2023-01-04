package ploy

import (
	"log"
	"mahjong/mj"
	"testing"
)

func TestWin(t *testing.T) {

	tiles := mj.Cards{
		mj.W1, mj.W2, mj.W1, mj.W2, mj.W3, mj.W3, mj.W3, mj.W4, mj.W5, mj.T1, mj.T1,
	}

	winChecker := mj.NewWinChecker()
	coms := winChecker.CheckAll(tiles)

	for _, c := range coms {
		log.Println(c)
	}

}
