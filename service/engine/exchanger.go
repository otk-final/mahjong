package engine

import (
	"log"
	"mahjong/server/api"
	"sync"
	"time"
)

type Exchanger struct {
	lock       sync.Mutex
	pos        *Position
	handler    NotifyHandle
	_ack       *ackQueue
	_cd        *countdown
	_isRunning bool
	takeCh     chan *api.TakePayload
	putCh      chan *api.PutPayload
	raceCh     chan *api.RacePayload
	winCh      chan *api.WinPayload
	ackCh      chan *api.AckPayload
}

type NotifyHandle interface {
	// Take 摸牌
	Take(event *api.TakePayload)
	// Put 出牌
	Put(event *api.PutPayload)
	// Race 抢占
	Race(event *api.RacePayload)
	// Win 胡牌
	Win(event *api.WinPayload)
	// Ack 回执确认
	Ack(event *api.AckPayload)
	// Turn 轮转
	Turn(event *api.TurnPayload, ok bool)
	//Quit 退出
	Quit(ok bool)
}

// 回执队列
type ackQueue struct {
	threshold int
	incr      int
	members   int
}

func (aq *ackQueue) reset() {
	aq.threshold = 0
}

func (aq *ackQueue) newAckId() int {
	aq.incr++
	aq.threshold = aq.incr * aq.members
	return aq.incr
}

func (aq *ackQueue) incrId() int {
	return aq.incr
}

func (aq *ackQueue) ready(who int, ackId int) bool {
	aq.threshold = aq.threshold - ackId
	return aq.threshold == 0
}

func NewExchanger(handler NotifyHandle, pos *Position) *Exchanger {
	return &Exchanger{
		handler: handler,
		pos:     pos,
		takeCh:  make(chan *api.TakePayload, 1),
		putCh:   make(chan *api.PutPayload, 1),
		raceCh:  make(chan *api.RacePayload, 1),
		winCh:   make(chan *api.WinPayload, 1),
		ackCh:   make(chan *api.AckPayload, 1),
	}
}
func (exc *Exchanger) Run(interval int) {
	go exc.start(interval)
}
func (exc *Exchanger) start(interval int) {
	pos := exc.pos
	handler := exc.handler

	//计时器
	cd := newCountdown(interval)
	exc._cd = cd

	//default 就绪队列 除自己
	aq := &ackQueue{members: pos.Num() - 1}
	exc._ack = aq

	//释放
	defer exc.stop()

	//从庄家开始
	exc._isRunning = true
	exc.pos.move(exc.pos.master.Idx)

	//堵塞监听
	for {
		select {
		case t := <-exc.takeCh:
			//从摸牌开始，开始倒计时
			cd.restart(true)
			//牌库摸完了 结束当前回合
			if t.Tile == -1 {
				handler.Quit(false)
				return
			}
			handler.Take(t)
		case p := <-exc.putCh:
			//每当出一张牌，均需等待其他玩家确认或者抢占
			aq.newAckId()
			//出牌事件
			handler.Put(p)
		case r := <-exc.raceCh:
			//抢占 碰，杠，吃，... 设置当前回合
			pos.move(r.Who)
			//开始倒计时
			cd.restart(true)
			//并清除待ack队列
			aq.reset()
			//通知
			handler.Race(r)
		case _ = <-exc.winCh:
			//通知当局游戏结束
			handler.Quit(true)
			return
		case a := <-exc.ackCh:
			//过期则忽略当前事件
			if a.AckId < aq.incrId() {
				continue
			}
			handler.Ack(a)
			//就绪事件
			if aq.ready(a.Who, a.AckId) {
				//重置定时
				cd.restart(true)
				//正常轮转下家
				pre := pos.turnIdx
				who := pos.next()
				handler.Turn(&api.TurnPayload{Pre: pre, Who: who, Interval: interval}, true)
			}
		case <-cd.ticker.C:
			//超时，玩家无任何动作

			//并清除待ack队列
			aq.reset()
			//倒计
			cd.restart(false)

			pre := pos.turnIdx
			who := pos.next()
			//非正常轮转下家
			handler.Turn(&api.TurnPayload{Pre: pre, Who: who, Interval: interval}, false)
		}
	}
}

func (exc *Exchanger) CurrentAckId() int {
	return exc._ack.incrId()
}

func (exc *Exchanger) TurnTime() int {
	return exc._cd.remaining()
}

func (exc *Exchanger) PostTake(e *api.TakePayload) {
	if exc.isRunning() {
		exc.takeCh <- e
	}
}
func (exc *Exchanger) PostPut(e *api.PutPayload) {
	if exc.isRunning() {
		exc.putCh <- e
	}
}
func (exc *Exchanger) PostRace(e *api.RacePayload) {
	if exc.isRunning() {
		exc.raceCh <- e
	}
}
func (exc *Exchanger) PostAck(e *api.AckPayload) {
	if exc.isRunning() {
		exc.ackCh <- e
	}
}
func (exc *Exchanger) PostWin(e *api.WinPayload) {
	//通知结束 只通知一次
	if exc.isRunning() {
		exc.winCh <- e
	}
	//判定只会有存在一家，但胡牌可能是多家，则直接通知
	exc.handler.Win(e)
}

func (exc *Exchanger) isRunning() bool {
	defer exc.lock.Unlock()
	exc.lock.Lock()
	return exc._isRunning
}

func (exc *Exchanger) stop() {
	defer exc.lock.Unlock()
	exc.lock.Lock()

	log.Println("结束 exchanger")

	//close
	close(exc.takeCh)
	close(exc.putCh)
	close(exc.raceCh)
	close(exc.winCh)
	close(exc.ackCh)
	exc._cd.stop()
	exc._isRunning = false
}

type countdown struct {
	interval time.Duration
	ticker   *time.Ticker
	nextTime time.Time
}

func newCountdown(second int) *countdown {
	interval := time.Duration(second) * time.Second
	return &countdown{
		interval: interval,
		ticker:   time.NewTicker(interval),
		nextTime: time.Now().Add(interval),
	}
}

func (c *countdown) restart(reset bool) {
	if reset {
		c.ticker.Reset(c.interval)
	}
	//下一次时间
	c.nextTime = time.Now().Add(c.interval)
}

func (c countdown) remaining() int {
	return int(c.nextTime.Sub(time.Now()).Seconds())
}

func (c *countdown) stop() {
	c.ticker.Stop()
}
