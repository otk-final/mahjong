package engine

import (
	"errors"
	"mahjong/server/api"
	"time"
)

type Countdown struct {
	//超时时间
	timeout time.Duration
	locator *Position
	takeCh  chan *api.TakePayload
	putCh   chan *api.PutPayload
	raceCh  chan *api.RacePayload
	ackCh   chan *api.AckPayload
}

type NotifyHandle interface {
	// Take 摸牌
	Take(event *api.TakePayload)
	// Put 出牌
	Put(ackId int, event *api.PutPayload)
	// Race 抢占
	Race(event *api.RacePayload)
	// Win 胡牌
	Win(event *api.RacePayload) bool
	// Ack 回执确认
	Ack(event *api.AckPayload)
	// Next 轮转
	Next(who int, ok bool)
	//Quit 退出
	Quit()
}

// 回执队列
type ackQueue struct {
	ackInit  int
	batchIdx int
	members  int
}

func (aq *ackQueue) reset() {
	aq.ackInit = 0
}

func (aq *ackQueue) ackId() int {
	aq.batchIdx++
	aq.ackInit = aq.batchIdx * aq.members
	return aq.batchIdx
}

func (aq *ackQueue) ready(who int, ackId int) bool {
	aq.ackInit = aq.ackInit - ackId
	return aq.ackInit == 0
}

func NewCountdown(timeout time.Duration) *Countdown {
	return &Countdown{
		timeout: timeout,
		takeCh:  make(chan *api.TakePayload, 0),
		putCh:   make(chan *api.PutPayload, 0),
		raceCh:  make(chan *api.RacePayload, 0),
		ackCh:   make(chan *api.AckPayload, 0),
	}
}

func (cd *Countdown) Run(handler NotifyHandle, pos *Position) {

	//计时器
	dt := time.NewTicker(cd.timeout)

	//释放
	defer func() {
		dt.Stop()
		close(cd.takeCh)
		close(cd.putCh)
		close(cd.raceCh)
		close(cd.ackCh)
	}()

	cd.locator = pos

	//从庄家开始
	masterIdx := pos.start()
	go handler.Next(masterIdx, true)

	//default 就绪队列 除自己
	aq := &ackQueue{members: pos.seatRing.Len() - 1}

	//堵塞监听
	for {
		select {
		case t := <-cd.takeCh:
			//从摸牌开始，开始倒计时
			dt.Reset(cd.timeout)
			//牌库摸完了 结束当前回合
			if t.Take == -1 {
				//退出
				handler.Quit()
				return
			}
			//摸牌事件
			handler.Take(t)
		case p := <-cd.putCh:
			//每当出一张牌，均需等待其他玩家确认或者抢占
			ackId := aq.ackId()
			//出牌事件
			handler.Put(ackId, p)
		case r := <-cd.raceCh:
			//抢占 碰，杠，吃，胡... 设置当前回合
			pos.move(r.Who)

			//并清除待ack队列
			aq.reset()

			//重制定时器
			dt.Reset(cd.timeout)

			//事件通知
			if r.RaceType == api.WinRace { //根据业务规则判断胡牌后是否继续
				goon := handler.Win(r)
				if goon {
					//退出
					handler.Quit()
					return
				}
			} else {
				handler.Race(r)
			}
		case a := <-cd.ackCh:
			handler.Ack(a)
			//就绪事件
			if aq.ready(a.Who, a.AckId) {
				//正常轮转下家
				who := pos.next()
				handler.Next(who, true)
			}
		case <-dt.C:
			//并清除待ack队列
			aq.reset()
			//超时，玩家无任何动作
			who := pos.next()
			//非正常轮转下家
			handler.Next(who, false)
		}
	}
}

func (cd *Countdown) ToTake(e *api.TakePayload) error {
	if !cd.locator.Check(e.Who) {
		return errors.New("not current round")
	}
	cd.takeCh <- e
	return nil
}
func (cd *Countdown) ToPut(e *api.PutPayload) error {
	if !cd.locator.Check(e.Who) {
		return errors.New("not current round")
	}
	cd.putCh <- e
	return nil
}
func (cd *Countdown) ToRace(e *api.RacePayload) {
	cd.raceCh <- e
}
func (cd *Countdown) ToAck(e *api.AckPayload) {
	cd.ackCh <- e
}
