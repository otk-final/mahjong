package engine

import (
	"errors"
	"mahjong/server/api"
	"strings"
	"sync"
)

type Position struct {
	lock    *sync.Mutex
	turnIdx int
	num     int
	master  *api.Player //庄家
	members []*api.Player
	robots  map[int]*api.Roboter
}

func (pos *Position) RobotClosed(p *api.Player) {
	//同步
	defer pos.lock.Unlock()
	pos.lock.Lock()

	delete(pos.robots, p.Idx)
}

func (pos *Position) RobotOpen(p *api.Player, level int) {
	//同步
	defer pos.lock.Unlock()
	pos.lock.Lock()

	pos.robots[p.Idx] = &api.Roboter{Player: p, Level: level}
}

func (pos *Position) IsRobot(pIdx int) (bool, *api.Roboter) {
	level, ok := pos.robots[pIdx]
	return ok, level
}

func NewPosition(num int, master *api.Player) (*Position, error) {
	//虚位以待 join
	if master.Idx >= num {
		return nil, errors.New("master idx error")
	}
	return &Position{
		lock:    &sync.Mutex{},
		turnIdx: -1,
		num:     num,
		master:  master,
		members: []*api.Player{master},
		robots:  make(map[int]*api.Roboter, 0),
	}, nil
}

func (pos *Position) next() int {

	//同步
	defer pos.lock.Unlock()
	pos.lock.Lock()

	if pos.turnIdx == len(pos.members)-1 {
		pos.turnIdx = 0
	} else {
		//next
		pos.turnIdx++
	}
	return pos.turnIdx
}

func (pos *Position) move(who int) {
	//同步
	defer pos.lock.Unlock()
	pos.lock.Lock()

	pos.turnIdx = who
}

func (pos *Position) Check(who int) bool {
	return pos.turnIdx == who
}

func (pos *Position) TurnIdx() int {
	return pos.turnIdx
}

func (pos *Position) PreIdx() int {
	pre := pos.turnIdx - 1
	if pre < 0 {
		return pos.num - 1
	}
	return pre
}

// Join 就坐
func (pos *Position) Join(p *api.Player) error {

	//同步
	defer pos.lock.Unlock()
	pos.lock.Lock()

	joinCount := len(pos.members)
	//是否满座
	if joinCount == pos.num {
		return errors.New("full members")
	}

	//自动选座 下标从0开始
	if p.Idx == -1 {
		p.Idx = joinCount
	}

	if p.Idx > pos.num-1 {
		return errors.New("idx offset")
	}

	pos.members = append(pos.members, p)
	return nil
}

func (pos *Position) Joined() []*api.Player {
	//同步
	defer pos.lock.Unlock()
	pos.lock.Lock()

	joined := make([]*api.Player, len(pos.members))
	copy(joined, pos.members)
	return joined
}

func (pos *Position) IsMaster(acctId string) bool {
	return strings.EqualFold(pos.master.UId, acctId)
}

func (pos *Position) Index(acctId string) (*api.Player, error) {
	for _, m := range pos.members {
		if strings.EqualFold(m.UId, acctId) {
			return m, nil
		}
	}
	return nil, errors.New("not found")
}

func (pos *Position) Len() int {
	return len(pos.Joined())
}

func (pos *Position) Num() int {
	return pos.num
}
