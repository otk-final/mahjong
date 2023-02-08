package engine

import (
	"mahjong/server/api"
	"time"
)

type Exchanger struct {
	pos    *Position
	ack    *ackQueue
	cd     *countdown
	takeCh chan *api.TakePayload
	putCh  chan *api.PutPayload
	raceCh chan *api.RacePayload
	ackCh  chan *api.AckPayload
}

type NotifyHandle interface {
	// Take 摸牌
	Take(event *api.TakePayload)
	// Put 出牌
	Put(event *api.PutPayload)
	// Race 抢占
	Race(event *api.RacePayload)
	// Win 胡牌
	Win(event *api.RacePayload)
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

func NewExchanger() *Exchanger {
	return &Exchanger{
		takeCh: make(chan *api.TakePayload, 0),
		putCh:  make(chan *api.PutPayload, 0),
		raceCh: make(chan *api.RacePayload, 0),
		ackCh:  make(chan *api.AckPayload, 0),
	}
}
func (exc *Exchanger) Run(handler NotifyHandle, pos *Position, interval int) {
	go exc.start(handler, pos, interval)
}
func (exc *Exchanger) start(handler NotifyHandle, pos *Position, interval int) {

	//计时器
	cd := newCountdown(interval)
	exc.cd = cd

	//default 就绪队列 除自己
	aq := &ackQueue{members: pos.Num() - 1}
	exc.ack = aq

	//释放
	defer func() {
		cd.stop()
		close(exc.takeCh)
		close(exc.putCh)
		close(exc.raceCh)
		close(exc.ackCh)
	}()

	//从庄家开始
	pos.move(pos.master.Idx)

	//堵塞监听
	for {
		select {
		case t := <-exc.takeCh:
			//从摸牌开始，开始倒计时
			cd.reset()
			//牌库摸完了 结束当前回合
			if t.Tile == -1 {
				handler.Quit(false)
				return
			}
			handler.Take(t)
		case p := <-exc.putCh:
			//每当出一张牌，均需等待其他玩家确认或者抢占
			p.AckId = aq.newAckId()
			//出牌事件
			handler.Put(p)
		case r := <-exc.raceCh:
			//抢占 碰，杠，吃，... 设置当前回合
			pos.move(r.Who)
			//并清除待ack队列
			aq.reset()
			//重制定时器
			cd.reset()
			//胡牌则退出
			if r.RaceType == api.WinRace {
				handler.Win(r)
				return
			} else {
				handler.Race(r)
			}
		case a := <-exc.ackCh:
			//过期则忽略当前事件
			if a.AckId < aq.incrId() {
				continue
			}
			handler.Ack(a)
			//就绪事件
			if aq.ready(a.Who, a.AckId) {
				//正常轮转下家
				who := pos.next()
				handler.Turn(who, interval, true)
			}
		case <-cd.delay():
			//并清除待ack队列
			aq.reset()
			cd.next()
			//超时，玩家无任何动作
			who := pos.next()
			//非正常轮转下家
			handler.Turn(who, interval, false)
		}
	}
}

func (exc *Exchanger) CurrentAckId() int {
	return exc.ack.incrId()
}

func (exc *Exchanger) TurnTime() int {
	return exc.cd.remaining()
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

func (c *countdown) reset() {
	//下一次时间
	c.nextTime = time.Now().Add(c.interval)
	c.timer.Reset(c.interval)
}

func (c *countdown) next() {
	//下一次时间
	c.nextTime = time.Now().Add(c.interval)
}

func (c countdown) remaining() int {
	return int(c.nextTime.Sub(time.Now()).Seconds())
}

func (c *countdown) delay() <-chan time.Time {
	return c.timer.C
}

func (c *countdown) stop() {
	c.timer.Stop()
}
