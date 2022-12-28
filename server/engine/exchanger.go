package engine

import (
	"mahjong/server/api"
	"time"
)

type Exchanger struct {
	//超时时间
	timeout time.Duration
	pos     *Position
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

func NewExchanger(timeout time.Duration) *Exchanger {
	return &Exchanger{
		timeout: timeout,
		takeCh:  make(chan *api.TakePayload, 0),
		putCh:   make(chan *api.PutPayload, 0),
		raceCh:  make(chan *api.RacePayload, 0),
		ackCh:   make(chan *api.AckPayload, 0),
	}
}

func (exc *Exchanger) Run(handler NotifyHandle, pos *Position) {

	//计时器
	dt := time.NewTicker(exc.timeout)

	//释放
	defer func() {
		dt.Stop()
		close(exc.takeCh)
		close(exc.putCh)
		close(exc.raceCh)
		close(exc.ackCh)
	}()

	exc.pos = pos

	//从庄家开始
	masterIdx := pos.start()
	go handler.Next(masterIdx, true)

	//default 就绪队列 除自己
	aq := &ackQueue{members: pos.seatRing.Len() - 1}

	//堵塞监听
	for {
		select {
		case t := <-exc.takeCh:
			//从摸牌开始，开始倒计时
			dt.Reset(exc.timeout)
			//牌库摸完了 结束当前回合
			if t.Tile == -1 {
				//退出
				handler.Quit()
				return
			}
			//摸牌事件
			handler.Take(t)
		case p := <-exc.putCh:
			//每当出一张牌，均需等待其他玩家确认或者抢占
			ackId := aq.ackId()
			//出牌事件
			handler.Put(ackId, p)
		case r := <-exc.raceCh:
			//抢占 碰，杠，吃，胡... 设置当前回合
			pos.move(r.Who)

			//并清除待ack队列
			aq.reset()

			//重制定时器
			dt.Reset(exc.timeout)

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
		case a := <-exc.ackCh:
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
func (exc *Exchanger) GetPosition() *Position {
	return exc.pos
}
