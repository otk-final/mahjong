package engine

import (
	"log"
	"mahjong/mj"
	"sync"
	"testing"
)

func TestForward(t *testing.T) {
	tb := &Table{
		lock:    sync.Mutex{},
		dice:    0,
		tiles:   nil,
		remains: mj.Cards{1, 2, 3, 4, 5, 6},
	}
	tile := 0
	for tile != -1 {
		tile = tb.Forward()
		log.Printf("take forward :%d \n", tile)
	}
}

func TestBackward(t *testing.T) {
	tb := &Table{
		lock:    sync.Mutex{},
		dice:    0,
		tiles:   nil,
		remains: mj.Cards{1, 2, 3, 4, 5, 6},
	}
	tile := 0
	for tile != -1 {
		tile = tb.Backward()
		log.Printf("take backward :%d \n", tile)
	}
}
