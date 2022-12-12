package countdown

import (
	"github.com/google/uuid"
	"sync"
	"time"
)

type timer[T any] struct {
	traceId    string
	funcOnce   *sync.Once
	readyCh    chan *T
	doneCh     chan struct{}
	startTime  time.Time
	expireTime time.Time
}

type CallData[T any] struct {
	Action   ActionType
	Payloads []*T
}
type Callable[R any] func(data CallData[R])

type ActionType int

const (
	IsReady ActionType = iota
	IsClose
	IsExpired
)

var countdownMap = &sync.Map{}

func NewTrackId(roomId string) string {
	return uuid.NewString()
}

func listener[R any](tm *timer[R], delta int, fn Callable[R]) {

	//倒计时
	tk := time.NewTicker(1 * time.Second)

	//释放
	defer func() {
		close(tm.readyCh)
		close(tm.doneCh)
		tk.Stop()
		countdownMap.Delete(tm.traceId)
	}()

	tempData := make([]*R, 0)
	current := 0
	for {
		select {
		case p := <-tm.readyCh:
			tempData = append(tempData, p)
			//累计
			current++
			if current < delta {
				//重置剩余时间
				residue := tm.expireTime.Sub(time.Now())
				if residue <= 0 {
					residue = time.Second * 1
				}
				tk.Reset(residue)
				break
			}
			//通知业务方
			data := CallData[R]{
				Action:   IsReady,
				Payloads: tempData,
			}
			go tm.funcOnce.Do(func() { fn(data) })
		case <-tm.doneCh:
			//强制结束
			go tm.funcOnce.Do(func() { fn(CallData[R]{Action: IsClose, Payloads: tempData}) })
			return
		case <-tk.C:
			//自动超时
			go tm.funcOnce.Do(func() { fn(CallData[R]{Action: IsExpired, Payloads: tempData}) })
			return
		}
	}
}

func New[R any](traceId string, delta int, second int, call Callable[R]) {
	//计时器
	now := time.Now()
	tm := &timer[R]{
		traceId:    traceId,
		funcOnce:   &sync.Once{},
		readyCh:    make(chan *R, 0),
		doneCh:     make(chan struct{}, 0),
		startTime:  now,
		expireTime: now.Add(time.Duration(second) * time.Second),
	}
	countdownMap.Store(traceId, tm)

	//异步监听
	go listener(tm, delta, call)
}

func Ready[R any](traceId string, data *R) {
	temp, ok := countdownMap.Load(traceId)
	if !ok {
		return
	}
	tm := temp.(*timer[R])
	tm.readyCh <- data
}

func Close[R](traceId string) {
	temp, ok := countdownMap.Load(traceId)
	if !ok {
		return
	}
	tm := temp.(*timer[R])
	tm.doneCh <- struct{}{}
}
