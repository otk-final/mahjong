package mj

import (
	"log"
	"testing"
)

func TestName(t *testing.T) {
	tiles := Cards{1}
	t.Log(tiles[:len(tiles)-1])
}

func TestFilterABC(t *testing.T) {
	tiles := Cards{1, 2, 2, 3, 3, 4, 7, 8, 9, 34, 34}
	comb := NewWinChecker().Check(tiles)
	if comb != nil {
		log.Println(comb)
	}
}
