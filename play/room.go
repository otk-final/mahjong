package play

import (
	"sync"
)

// Room 房间
type Room struct {
	syncLock     sync.Mutex
	MaxRound     int //最大局数
	CurrentRound int //当前第几局
	Maker        int //当前庄家
	Players      []*GamePlayer
}

type GamePlayer struct {
	Place int
}

// GameTurn 每局游戏
type GameTurn struct {
	syncLock  sync.Mutex
	Maker     int //当前庄家
	Players   []*GamePlayer
	Table     *Table
	mjLibrary []int
}

func NewRoom() *Room {
	return &Room{
		syncLock:     sync.Mutex{},
		MaxRound:     0,
		CurrentRound: 0,
		Maker:        0,
		Players:      make([]*GamePlayer, 0),
	}
}

func JoinRoom(player *GamePlayer, room *Room) int {
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
