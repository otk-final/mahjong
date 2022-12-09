package server

import (
	"github.com/google/uuid"
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/countdown"
	"net/http"
)

type PlayerCtrl struct{}

// 摸牌
func (p *PlayerCtrl) take(w http.ResponseWriter, r *http.Request, body api.TakeCard) (*api.Reply, error) {

	//摸牌
	tb := mj.Table{}

	tb.ForwardAt()

	//加入牌库

	//通知当前用户判定
	websocketNotify("", "", 104, "", api.Empty)

	return nil, nil
}

// 出牌
func (p *PlayerCtrl) put(w http.ResponseWriter, r *http.Request, u api.PutCard) (*api.Reply, error) {

	//特殊牌判定

	//通知玩家
	for i := 0; i < 4; i++ {
		websocketNotify("", "", 103, "", api.Empty)
	}

	//通知其他玩家判定确认
	timerId := uuid.NewString()
	countdown.New(timerId, 4-1, 30, func(status int, inf *countdown.Timer) {
		if status == -1 {
			return
		}
		//TODO 无论超时与否，均通知下家进行摸牌
	})

	return nil, nil
}

// 判定
func (p *PlayerCtrl) reward(w http.ResponseWriter, r *http.Request, u api.RewardCard) (*api.Reply, error) {

	//吃
	//碰，

	//杠，通知当前用户摸牌

	timerId := uuid.NewString()
	countdown.New("", 4-1, 30, func(status int, inf *countdown.Timer) {

	})
	return nil, nil
}

// 胡
func (p *PlayerCtrl) win(w http.ResponseWriter, r *http.Request, u api.RewardCard) (*api.Reply, error) {
	return nil, nil
}

// 跳过
func (p *PlayerCtrl) skip(w http.ResponseWriter, r *http.Request, u api.RewardCard) (*api.Reply, error) {
	return nil, nil
}
