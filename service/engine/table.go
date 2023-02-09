package engine

import (
	"mahjong/mj"
	"math/rand"
	"sync"
	"time"
)

type Table struct {
	lock    sync.Mutex
	dice    int
	tiles   mj.Cards
	remains mj.Cards
}

// NewDice 掷骰子
func NewDice() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(6) + 1
}

func NewTable(dice int, tileLib mj.Cards) *Table {
	return &Table{
		lock:  sync.Mutex{},
		dice:  dice,
		tiles: tileLib,
	}
}

// Shuffle 洗牌
func Shuffle(cards mj.Cards) mj.Cards {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
	return cards
}

// Distribution 发牌
func (tb *Table) Distribution(num int) map[int]mj.Cards {

	//等比分4方
	sideCount := len(tb.tiles) / 4

	//顺时针摆牌
	splitIdx := 0
	//	骰子：1，3，5对，2，6顺，4反
	switch tb.dice {
	case 1, 3, 5:
		splitIdx = sideCount*2 + tb.dice*2
		break
	case 2, 6:
		splitIdx = sideCount*1 + tb.dice*2
		break
	case 4:
		splitIdx = sideCount*3 + tb.dice*2
		break
	}

	//重新排序
	forward := tb.tiles[:splitIdx]
	backward := tb.tiles[splitIdx:]

	newLibrary := make([]int, 0)
	newLibrary = append(newLibrary, backward...)
	newLibrary = append(newLibrary, forward...)
	tb.remains = newLibrary

	members := make(map[int]mj.Cards, 0)

	// init玩家手牌
	for i := 0; i < num; i++ {
		members[i] = make([]int, 0)
	}

	startIdx := 0
	//发牌 共13张 3 * 4 + 1
	for i := 0; i < 4; i++ {
		count := 4
		if i == 3 { //最后轮流一张
			count = 1
		}
		for m := 0; m < num; m++ {
			members[m] = append(members[m], tb.remains[startIdx:startIdx+count]...)
			startIdx = startIdx + count
		}
	}
	return members
}

func (tb *Table) Forward() int {
	defer tb.lock.Unlock()
	tb.lock.Lock()

	//empty
	if len(tb.remains) == 0 {
		return -1
	}

	head := tb.remains[0]
	tb.remains = tb.remains[1:]
	return head
}

func (tb *Table) Backward() int {
	defer tb.lock.Unlock()
	tb.lock.Lock()

	//empty
	if len(tb.remains) == 0 {
		return -1
	}

	tailIdx := len(tb.remains) - 1
	tail := tb.remains[tailIdx]
	tb.remains = tb.remains[0:tailIdx]
	return tail
}
func (tb *Table) Remains() int {
	return len(tb.remains)
}