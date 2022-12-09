package countdown

import (
	"sync"
	"time"
)

type Type int

const (
	IsTimeOut Type = iota
	IsDone
	IsAck
)

type Timer struct {
	id       string
	fnOnce   *sync.Once
	delayCh  chan int
	delay    int
	InitTime time.Time
	AckTimes []time.Time
	DoneTime time.Time
}
type Feedback func(status Type, tm *Timer)

var countdownMap = &sync.Map{}

func New(timerId string, delta int, second int, fn Feedback) {

	tm := &Timer{
		id:       timerId,
		fnOnce:   &sync.Once{},
		delayCh:  make(chan int, delta),
		delay:    delta,
		InitTime: time.Now(),
		AckTimes: make([]time.Time, 0),
	}
	countdownMap.Store(timerId, tm)

	//异步监听
	go func(tm *Timer) {
		//释放
		defer func() {
			close(tm.delayCh)
			countdownMap.Delete(tm.id)
		}()

		inc := 0
		for inc < tm.delay {
			//堵塞执行
			select {
			case key := <-tm.delayCh:
				//手动关闭 退出
				if key == -1 {
					go tm.fnOnce.Do(func() { fn(IsDone, tm) })
					return
				}
				//正常
				inc = inc + key
				if inc < tm.delay {
					continue
				}
				//正常回执
				go tm.fnOnce.Do(func() { fn(IsAck, tm) })
			case <-time.After(time.Duration(second) * time.Second):
				//异常回执
				go tm.fnOnce.Do(func() { fn(IsTimeOut, tm) })
			}
		}
	}(tm)
}

func Ack(timerId string) {
	temp, ok := countdownMap.Load(timerId)
	if !ok {
		return
	}
	tm := temp.(*Timer)
	tm.AckTimes = append(tm.AckTimes, time.Now())
	tm.delayCh <- 1
}

func Done(timerId string) {
	temp, ok := countdownMap.Load(timerId)
	if !ok {
		return
	}
	tm := temp.(*Timer)
	tm.DoneTime = time.Now()
	tm.delayCh <- -1
}
