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
	defer tb.lock.Unlock()
	tb.lock.Lock()

	// init玩家手牌
	members := map[int]mj.Cards{}
	startIdx := 0
	//发牌 共13张 3 * 4 + 1
	for i := 0; i < 4; i++ {
		takeCount := 4
		if i == 3 { //最后轮流一张
			takeCount = 1
		}
		for m := 0; m < num; m++ {
			members[m] = append(members[m], tb.tiles[startIdx:startIdx+takeCount]...)
			startIdx = startIdx + takeCount
		}
	}
	//剩于牌
	tb.remains = tb.tiles[startIdx:]
	return members
}

func (tb *Table) Append(tile ...int) {
	defer tb.lock.Unlock()
	tb.lock.Lock()

	tb.remains = append(tb.remains, tile...)
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
	tb.remains = tb.remains[:tailIdx]
	return tail
}

func (tb *Table) Remains() int {
	return len(tb.remains)
}
