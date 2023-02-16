package store

import (
	"errors"
	"github.com/google/uuid"
	"log"
	"mahjong/server/api"
	"net/http"
	"strings"
	"sync"
	"time"
)

var visitorMax = 10000

type visitorMap struct {
	lock      sync.RWMutex
	visitors  map[string]*api.Visitor
	visitTime map[string]time.Time
}

var vm = &visitorMap{
	lock:      sync.RWMutex{},
	visitors:  map[string]*api.Visitor{},
	visitTime: map[string]time.Time{},
}

var liveCh = make(chan string, 1000)
var freeCh = make(chan string, 1000)

func init() {
	vt := time.NewTicker(1 * time.Minute)
	go func() {
		defer func() {
			close(liveCh)
			close(freeCh)
			vt.Stop()
		}()

		for {
			select {
			case id, _ := <-liveCh:
				setLiving(id)
			case id, _ := <-freeCh:
				setFreed(id)
			case <-vt.C:
				expired()
			}
		}
	}()
}

func setLiving(id string) {
	defer vm.lock.Unlock()
	vm.lock.Lock()

	vm.visitTime[id] = time.Now()
}

func setFreed(id string) {
	defer vm.lock.Unlock()
	vm.lock.Lock()

	log.Printf("free visitor:%s\n", id)
	delete(vm.visitTime, id)
	delete(vm.visitors, id)
}

var expireTime = time.Duration(1) * time.Minute

func expired() {
	//读锁
	defer vm.lock.RUnlock()
	vm.lock.RLock()

	now := time.Now()
	for k, v := range vm.visitTime {
		if now.Sub(v) < expireTime {
			continue
		}
		freeCh <- k
	}
}

func NewVisitor(request *http.Request) (*api.Visitor, error) {
	if len(vm.visitors) >= visitorMax {
		return nil, errors.New("游客超限，稍后访问")
	}

	defer vm.lock.Unlock()
	vm.lock.Lock()

	//save
	vs := &api.Visitor{
		Player: &api.Player{
			Idx:   -1,
			UId:   uuid.NewString(),
			UName: "游客",
			Ip:    request.RemoteAddr,
		},
		Token: uuid.NewString(),
	}
	vm.visitors[vs.UId] = vs
	vm.visitTime[vs.UId] = time.Now()

	return vs, nil
}

func IsValid(uid, token string) (bool, *api.Visitor) {
	//读锁
	defer vm.lock.RUnlock()
	vm.lock.RLock()

	visitor, ok := vm.visitors[uid]
	if ok {
		//异步写
		liveCh <- visitor.UId
		if strings.EqualFold(visitor.Token, token) {
			return true, visitor
		}
	}
	return false, nil
}

func FreeVisitor(id string) {
	freeCh <- id
}
