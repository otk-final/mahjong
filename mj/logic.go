package mj

import (
	"container/ring"
	"errors"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

type Room struct {
	Id       string
	syncLock sync.Mutex
	seats    *ring.Ring
	seatsIdx int
	members  int //玩家数
	round    int
	roundIdx int
}

type Table struct {
	syncLock      sync.Mutex
	members       int   //玩家数
	dice          int   //骰子数
	fIdx          int   //向前
	bIdx          int   //向后
	remainLibrary Cards //余牌
}

func NewRoom(members int) *Room {
	return &Room{
		syncLock: sync.Mutex{},
		seats:    ring.New(members),
		members:  members,
	}
}

func NewDice() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(6) + 1
}

func NewTable(members, dice int) *Table {
	return &Table{
		members:       members,
		dice:          dice,
		fIdx:          -1,
		bIdx:          -1,
		remainLibrary: make([]int, 0),
	}
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
func (table *Table) Allocate(initLibrary Cards) map[int]Cards {

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

	members := make(map[int]Cards, 0)

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

func (room *Room) Join(p *Player) {
	defer room.syncLock.Unlock()
	room.syncLock.Lock()

	p.Idx = room.seatsIdx
	room.seats.Value = p
	room.seats = room.seats.Next()

	//座位依次就坐
	room.seatsIdx++
}

func (room *Room) Exit(pIdx int) {
	defer room.syncLock.Unlock()
	room.syncLock.Lock()
	if pIdx >= room.seats.Len() {
		return
	}
	room.seats.Unlink(pIdx)
}

func (room *Room) Ready() ([]*Player, bool) {
	ps := make([]*Player, 0)
	room.seats.Do(func(a any) {
		ps = append(ps, a.(*Player))
	})
	return ps, room.seats.Len() == room.members
}

func (room *Room) Player(pid string) (*Player, error) {
	var p *Player
	room.seats.Do(func(a any) {
		temp := a.(*Player)
		if strings.EqualFold(temp.Id, pid) {
			p = temp
			return
		}
	})
	if p == nil {
		return nil, errors.New("not found")
	}
	return p, nil
}

// TurnCheck 我的回合？
func (room *Room) TurnCheck(sIdx int) (bool, *Player) {
	p := room.seats.Value.(*Player)
	return p.Idx == sIdx, p
}

// TurnChange 回合
func (room *Room) TurnChange(sIdx int) *Player {
	defer room.syncLock.Unlock()
	room.syncLock.Lock()

	//移动定位
	return room.seats.Move(sIdx).Value.(*Player)
}

// TurnNext 下一回合
func (room *Room) TurnNext() *Player {
	defer room.syncLock.Unlock()
	room.syncLock.Lock()
	return room.seats.Next().Value.(*Player)
}

// TurnPlayer 当前回合
func (room *Room) TurnPlayer() *Player {
	return room.seats.Value.(*Player)
}

// Player 玩家手上的牌
type Player struct {
	Id          string
	Idx         int
	HandCards   Cards
	PutCards    Cards
	RewardGroup []Cards
}

func NewPlayer(pid string) *Player {
	return &Player{
		Id:          pid,
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
