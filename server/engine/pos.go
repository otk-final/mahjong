package engine

import (
	"container/ring"
	"errors"
	"mahjong/server/api"
	"strings"
	"sync"
)

type Position struct {
	lock     *sync.Mutex
	seatRing *ring.Ring  //座位
	master   *api.Player //庄家
	members  map[string]*api.Player
}

func NewPosition(members int, master *api.Player) (*Position, error) {
	//虚位以待 join
	if master.Idx >= members {
		return nil, errors.New("master idx error")
	}
	return &Position{
		lock:     &sync.Mutex{},
		seatRing: ring.New(members),
		master:   master,
		members:  make(map[string]*api.Player, members),
	}, nil
}

func (pos *Position) next() int {

	//同步
	defer pos.lock.Unlock()
	pos.lock.Lock()

	//next
	pos.seatRing = pos.seatRing.Next()
	return pos.seatRing.Value.(*api.Player).Idx
}

func (pos *Position) move(who int) {
	//同步
	defer pos.lock.Unlock()
	pos.lock.Lock()

	nowIdx := pos.seatRing.Value.(*api.Player).Idx
	pos.seatRing = pos.seatRing.Move(who - nowIdx)
}

func (pos *Position) Check(who int) bool {
	return pos.seatRing.Value.(*api.Player).Idx == who
}

//从庄家开始
func (pos *Position) start() int {
	pos.seatRing = pos.seatRing.Move(pos.master.Idx)
	return pos.master.Idx
}

// Join 就坐
func (pos *Position) Join(p *api.Player) error {

	//同步
	defer pos.lock.Unlock()
	pos.lock.Lock()

	//自动选座
	if p.Idx == -1 {
		joinCount := 0
		pos.seatRing.Do(func(a any) {
			if a != nil {
				joinCount++
			}
		})
		p.Idx = joinCount
	}

	if p.Idx >= pos.seatRing.Len() {
		return errors.New("index outset")
	}

	exist := pos.seatRing.Move(p.Idx)
	if exist.Value != nil {
		return errors.New("exist player")
	}
	exist.Value = p

	pos.members[p.AcctId] = p
	return nil
}

func (pos *Position) Ready() bool {
	return false
}

func (pos *Position) Joined() []*api.Player {
	joins := make([]*api.Player, 0)
	pos.seatRing.Do(func(a any) {
		if a != nil {
			joins = append(joins, a.(*api.Player))
		}
	})
	return joins
}

func (pos *Position) IsMaster(acctId string) bool {
	return strings.EqualFold(pos.master.AcctId, acctId)
}

func (pos *Position) Index(acctId string) (*api.Player, error) {
	m, ok := pos.members[acctId]
	if !ok {
		return nil, errors.New("not found")
	}
	return m, nil
}
