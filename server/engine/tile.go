package engine

import (
	"mahjong/mj"
	"math/rand"
	"sync"
	"time"
)

// TileHandle 当局
type TileHandle interface {

	// GetOuts 已出牌
	GetOuts(pIdx int) []int
	// GetHands 手上牌
	GetHands(pIdx int) []int
	// GetRaces 生效牌
	GetRaces(pIdx int) [][]int

	AddTake(pIdx int, tile int)

	AddPut(pIdx int, tile int)

	AddRace(pIdx int, tiles []int, whoIdx int, tile int)

	Forward(pIdx int) int

	Backward(pIdx int) int
}

type Table struct {
	lock  sync.Mutex
	dice  int
	tiles mj.Cards
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
func (tb *Table) Distribution(members int) map[int]mj.Cards {
	return nil
}
