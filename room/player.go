package room

import (
	"mahjong/mj"
	"sync"
)

type Player struct {
	Master bool
	Place  int
}

// PlayerCards 玩家牌组
type PlayerCards struct {
	HandOf  []int   //  手上的牌
	TurnOf  [][]int //  回合判定的牌
	ThrowOf []int   //  打出的牌
}

// UserEvent 用户事件
type UserEvent string

const (
	UETake UserEvent = "take" //摸牌
	UEPut  UserEvent = "put"  //出牌
	UEPass UserEvent = "pass" //跳过
)

// CardEvent 卡牌事件
type CardEvent string

const (
	CEPair CardEvent = "pair" //碰
	CEGang CardEvent = "gang" //杠
	CEList CardEvent = "list" //吃
	CETing CardEvent = "ting" //听
	CEWin  CardEvent = "win"  //胡
)

// UserEventBooster 事件增强
type UserEventBooster interface {
	// 摸牌
	Take(room *GameRoom, table *TableCards, player *Player, takeCard int)
	// 出牌
	Put(room *GameRoom, table *TableCards, player *Player, putCard int)
	// 跳过
	Pass(room *GameRoom, table *TableCards, player *Player, putCard int)
}

func (p *Player) startAck(turnLock *sync.WaitGroup) {
	defer turnLock.Done()
}

func (p *Player) setCards(cards mj.Cards) {

}
