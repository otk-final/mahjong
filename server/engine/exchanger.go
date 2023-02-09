package engine

import (
	"log"
	"mahjong/server/api"
	"time"
)

type Exchanger struct {
	pos     *Position
	handler NotifyHandle
	_ack    *ackQueue
	_cd     *countdown
	_exit   bool
	takeCh  chan *api.TakePayload
	putCh   chan *api.PutPayload
	raceCh  chan *api.RacePayload
	winCh   chan *api.WinPayload
	ackCh   chan *api.AckPayload
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
	Turn(who int, interval int, ok bool)
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
	defer func() {
		log.Println("结束 exchanger")
		cd.stop()
		close(exc.takeCh)
		close(exc.putCh)
		close(exc.raceCh)
		close(exc.winCh)
		close(exc.ackCh)
	}()

	//从庄家开始
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
			break
		case p := <-exc.putCh:
			//每当出一张牌，均需等待其他玩家确认或者抢占
			p.AckId = aq.newAckId()
			//出牌事件
			handler.Put(p)
			break
		case r := <-exc.raceCh:
			//抢占 碰，杠，吃，... 设置当前回合
			pos.move(r.Who)
			//开始倒计时
			cd.restart(true)
			//并清除待ack队列
			aq.reset()
			//通知
			handler.Race(r)
			break
		case _ = <-exc.winCh:
			//通知当局游戏结束
			exc._exit = true
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
				who := pos.next()
				handler.Turn(who, interval, true)
			}
			break
		case <-cd.timer.C:
			//并清除待ack队列
			aq.reset()
			//倒计
			cd.restart(false)
			//超时，玩家无任何动作
			who := pos.next()
			//非正常轮转下家
			handler.Turn(who, interval, false)
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
	exc.takeCh <- e
}
func (exc *Exchanger) PostPut(e *api.PutPayload) {
	exc.putCh <- e
}
func (exc *Exchanger) PostRace(e *api.RacePayload) {
	exc.raceCh <- e
}
func (exc *Exchanger) PostAck(e *api.AckPayload) {
	exc.ackCh <- e
}
func (exc *Exchanger) PostWin(e *api.WinPayload) {
	//通知结束 只通知一次
	if !exc._exit {
		exc.winCh <- e
	}
	//判定只会有存在一家，但胡牌可能是多家，则直接通知
	exc.handler.Win(e)
}

type countdown struct {
	interval time.Duration
	timer    *time.Timer
	nextTime time.Time
}

func newCountdown(second int) *countdown {
	interval := time.Duration(second) * time.Second
	return &countdown{
		interval: interval,
		timer:    time.NewTimer(interval),
		nextTime: time.Now().Add(interval),
	}
}

func (c *countdown) restart(reset bool) {
	if reset {
		//ok := c.timer.Reset(c.interval)
		//log.Printf("重置定时：%v", ok)
	}
	//下一次时间
	c.nextTime = time.Now().Add(c.interval)
}

func (c countdown) remaining() int {
	return int(c.nextTime.Sub(time.Now()).Seconds())
}

func (c *countdown) stop() {
	c.timer.Stop()
}
