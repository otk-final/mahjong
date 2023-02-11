package engine

import (
	"mahjong/mj"
	"testing"
)

func TestTimer(t *testing.T) {

	//tb := NewTable(1, mj.LoadLibrary())
	//tb.Distribution(4)
	remains := mj.Cards{1, 2, 3, 4, 5, 6, 7}
	tailIdx := len(remains) - 1
	tail := remains[tailIdx]
	remains = remains[:tailIdx]

	t.Log(tail)
	t.Log(remains)
}
