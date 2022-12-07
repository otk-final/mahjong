package room

import (
	"sync"
)

// GameRoom 房间
type GameRoom struct {
	syncLock     sync.Mutex
	MaxRound     int //最大局数
	CurrentRound int //当前第几局
	Maker        int //当前庄家
	Players      []*Player
}

// GameTurn 每局游戏
type GameTurn struct {
	syncLock  sync.Mutex
	Maker     int //当前庄家
	Players   []*Player
	Table     *TableCards
	mjLibrary []int
}

func NewRoom() *GameRoom {
	return &GameRoom{
		syncLock:     sync.Mutex{},
		MaxRound:     0,
		CurrentRound: 0,
		Maker:        0,
		Players:      make([]*Player, 0),
	}
}

func JoinRoom(player *Player, room *GameRoom) int {
	defer room.syncLock.Unlock()
	room.syncLock.Lock()

	// 第一个进入房间的坐庄
	count := len(room.Players)
	if count == 0 {
		room.Maker = 0
	} else {
		count++
	}
	player.Place = count
	room.Players = append(room.Players, player)
	return count
}

func (room *GameRoom) NewGameTurn() *GameTurn {
	return nil
}

func (turn *GameTurn) findPlayer(idx int) *Player {
	return nil
}

// Run 开始一局游戏
func (turn *GameTurn) Run() {

	table := turn.Table
	// 堵塞 等待就绪
	wait := &sync.WaitGroup{}
	wait.Add(len(turn.Players))
	for _, p := range turn.Players {
		go p.startAck(wait)
	}
	wait.Done()

	//洗牌
	initCards := table.Shuffle(turn.mjLibrary)
	//发牌
	memberCards := table.Dispatch(initCards)
	//更新玩家牌库
	for k, cards := range memberCards {
		turn.findPlayer(k).setCards(cards)
	}
	for {

	}
}
