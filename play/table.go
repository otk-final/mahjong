package play

import (
	"mahjong/mj"
	"math/rand"
	"sync"
	"time"
)

type Table struct {
	syncLock      sync.Mutex
	players       int      //玩家数
	dice          int      //骰子数
	fIdx          int      //向前
	bIdx          int      //向后
	remainLibrary mj.Cards //余牌
}

func NewTable(players, dice int) *Table {
	return &Table{
		syncLock:      sync.Mutex{},
		players:       players,
		dice:          dice,
		fIdx:          -1,
		bIdx:          -1,
		remainLibrary: make([]int, 0),
	}
}

// Shuffle 洗牌
func (table *Table) Shuffle(initLibrary mj.Cards) mj.Cards {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(initLibrary), func(i, j int) {
		initLibrary[i], initLibrary[j] = initLibrary[j], initLibrary[i]
	})
	return initLibrary
}

// Dispatch 发牌
func (table *Table) Dispatch(initLibrary mj.Cards) map[int][]int {

	//等比分4方
	sideCount := len(initLibrary) / 4

	//顺时针摆牌
	splitIdx := 0
	//	骰子：1，3，5对，2，6顺，4反
	switch table.dice {
	case 1, 3, 5:
		splitIdx = sideCount*2 + table.dice*2
		break
	case 2, 6:
		splitIdx = sideCount*1 + table.dice*2
		break
	case 4:
		splitIdx = sideCount*3 + table.dice*2
		break
	}
	//重新排序
	forward := initLibrary[:splitIdx]
	backward := initLibrary[splitIdx:]

	newLibrary := make([]int, 0)
	newLibrary = append(newLibrary, backward...)
	newLibrary = append(newLibrary, forward...)
	table.remainLibrary = newLibrary

	members := make(map[int][]int, 0)

	// init玩家手牌
	for i := 0; i < table.players; i++ {
		members[i] = make([]int, 0)
	}

	startIdx := 0
	//发牌 3 * 4 + 1
	for i := 0; i < 4; i++ {
		count := 4
		if i == 3 { //最后轮流一张
			count = 1
		}
		for m := 0; m < table.players; m++ {
			members[m] = append(members[m], table.remainLibrary[startIdx:startIdx+count]...)
			startIdx = startIdx + count
		}
	}

	//记录当前位置
	table.fIdx = startIdx

	return members
}

// ForwardAt 从前摸
func (table *Table) ForwardAt() int {

	defer table.syncLock.Unlock()

	table.syncLock.Lock()
	table.fIdx++

	return table.remainLibrary[0]
}

// BackwardAt 从后摸
func (table *Table) BackwardAt() int {
	defer table.syncLock.Unlock()

	table.syncLock.Lock()
	table.bIdx++

	return table.remainLibrary[len(table.remainLibrary)-1]
}

// ThiefAt 偷牌
func (table *Table) ThiefAt(idx int) {

}
