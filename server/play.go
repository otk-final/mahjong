package server

import (
	"mahjong/server/api"
	"net/http"
)

type PlayerCtrl struct{}

// 摸牌
func (p *PlayerCtrl) take(w http.ResponseWriter, r *http.Request, body api.TakeCard) (*api.Reply, error) {

	//摸牌

	//加入牌库

	//通知玩家
	websocketReply(body.RoomId, body.RoomId, body)

	return nil, nil
}

// 出牌
func (p *PlayerCtrl) put(w http.ResponseWriter, r *http.Request, u api.PutCard) (*api.Reply, error) {

	//出牌

	//通知玩家

	//创建确认回执

	//通知玩家
	for {
		websocketReply(u.RoomId, "", nil)
	}

	return nil, nil
}

// 吃
func (p *PlayerCtrl) eat(w http.ResponseWriter, r *http.Request, u api.EatCard) (*api.Reply, error) {
	return nil, nil
}

// 杠
func (p *PlayerCtrl) gang(w http.ResponseWriter, r *http.Request, u api.GangCard) (*api.Reply, error) {
	return nil, nil
}

// 碰
func (p *PlayerCtrl) pair(w http.ResponseWriter, r *http.Request, u api.PairCard) (*api.Reply, error) {
	return nil, nil
}

// 胡
func (p *PlayerCtrl) win(w http.ResponseWriter, r *http.Request, u api.PairCard) (*api.Reply, error) {
	return nil, nil
}

// 跳过
func (p *PlayerCtrl) skip(w http.ResponseWriter, r *http.Request, u api.PairCard) (*api.Reply, error) {
	return nil, nil
}
