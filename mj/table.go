package mj

import (
	"container/ring"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

type Table struct {
	syncLock      sync.Mutex
	Locator       *ring.Ring
	members       int   //玩家数
	dice          int   //骰子数
	fIdx          int   //向前
	bIdx          int   //向后
	remainLibrary Cards //余牌
}

func NewDice() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(6) + 1
}

func NewTable(members, dice int) *Table {

	//环形座位链表 虚位待坐
	locator := ring.New(members)
	for i := 0; i < members; i++ {
		locator.Value = newPlayer(i)
		locator = locator.Next()
	}

	return &Table{
		syncLock:      sync.Mutex{},
		members:       members,
		Locator:       locator,
		dice:          dice,
		fIdx:          -1,
		bIdx:          -1,
		remainLibrary: make([]int, 0),
	}
}

func (table *Table) Join(idx int, uid string) {
	if idx > table.members {
		return
	}
	p := table.Locator.Move(idx).Value.(*Player)
	p.Id = uid
}

// Shuffle 洗牌
func (table *Table) Shuffle(initLibrary Cards) Cards {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(initLibrary), func(i, j int) {
		initLibrary[i], initLibrary[j] = initLibrary[j], initLibrary[i]
	})
	return initLibrary
}

// Allocate 发牌
func (table *Table) Allocate(initLibrary Cards) map[int][]int {

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
	for i := 0; i < table.members; i++ {
		members[i] = make([]int, 0)
	}

	startIdx := 0
	//发牌 共13张 3 * 4 + 1
	for i := 0; i < 4; i++ {
		count := 4
		if i == 3 { //最后轮流一张
			count = 1
		}
		for m := 0; m < table.members; m++ {
			members[m] = append(members[m], table.remainLibrary[startIdx:startIdx+count]...)
			startIdx = startIdx + count
		}
	}

	//记录当前位置
	table.fIdx = startIdx
	return members
}

// HeadAt 从前摸
func (table *Table) HeadAt() int {

	defer table.syncLock.Unlock()

	table.syncLock.Lock()
	table.fIdx++

	head := table.remainLibrary[0]

	table.remainLibrary = table.remainLibrary[1:]
	return head
}

// TailAt 从后摸
func (table *Table) TailAt() int {
	defer table.syncLock.Unlock()

	table.syncLock.Lock()
	table.bIdx++

	tail := table.remainLibrary[len(table.remainLibrary)-1]

	table.remainLibrary = table.remainLibrary[0 : len(table.remainLibrary)-2]
	return tail
}

// ThiefAt 偷牌
func (table *Table) ThiefAt(idx int) {

}

// TurnCheck 我的回合？
func (table *Table) TurnCheck(pid string) (bool, *Player) {
	p := table.Locator.Value.(*Player)
	return strings.EqualFold(p.Id, pid), p
}

// TurnChange 回合
func (table *Table) TurnChange(pid string) *Player {
	defer table.syncLock.Unlock()
	table.syncLock.Lock()

	//获取index
	idx := -1
	table.Locator.Do(func(a any) {
		p := a.(*Player)
		if strings.EqualFold(p.Id, pid) {
			idx = p.Idx
		}
	})
	if idx <= 0 {
		return nil
	}

	//移动定位
	return table.Locator.Move(idx).Value.(*Player)
}

// TurnNext 下一回合
func (table *Table) TurnNext() *Player {
	defer table.syncLock.Unlock()
	table.syncLock.Lock()
	return table.Locator.Next().Value.(*Player)
}

// TurnPlayer 当前回合
func (table *Table) TurnPlayer() *Player {
	return table.Locator.Value.(*Player)
}

// Player 玩家手上的牌
type Player struct {
	Id          string
	Idx         int
	HandCards   Cards
	PutCards    Cards
	RewardGroup []Cards
}

func newPlayer(num int) *Player {
	return &Player{
		Idx:         num,
		HandCards:   make(Cards, 0),
		PutCards:    make(Cards, 0),
		RewardGroup: make([]Cards, 0),
	}
}

func (p *Player) AddTakeCard(mj int) {
	p.HandCards = append(p.HandCards, mj)
}

func (p *Player) AddPutCard(mj int) {
	p.PutCards = append(p.PutCards, mj)
}

func (p *Player) AddRewardCards(source []int, target int) {

	rw := make(Cards, 0)
	copy(rw, source)
	source = append(source, target)

	sort.Ints(source)
	p.RewardGroup = append(p.RewardGroup, rw)
}
