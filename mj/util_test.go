package mj

import (
	"log"
	"testing"
)

func FuzzName(f *testing.F) {
	f.Add()
	f.Fuzz(func() {})
}

func TestFilterABC(t *testing.T) {

	tiles := Cards{1, 2, 2, 3, 3, 4, 7, 8, 9, 34, 34}
	ok, comb := NewWinChecker().Check(tiles)
	log.Println("is win", ok)
	if ok {
		log.Println(comb)
	}
}
