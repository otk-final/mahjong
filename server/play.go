package server

import (
	"errors"
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/countdown"
	"mahjong/server/wrap"
	"mahjong/strategy"
	"net/http"
)

type PlayerCtrl struct{}

// 摸牌
func (p *PlayerCtrl) take(w http.ResponseWriter, r *http.Request, body api.TakeCard) (*api.Reply, error) {

	header := wrap.GetHeader(r)
	userId := header.UserId

	tb := mj.Table{}
	//判断当前回合是否轮到自己摸牌
	ok, player := tb.TurnCheck(userId)
	if !ok {
		return nil, errors.New("非当前回合")
	}

	//摸牌
	var takeCard int
	if body.Direction == 1 {
		takeCard = tb.HeadAt()
	} else {
		takeCard = tb.TailAt()
	}
	//加入牌库
	player.AddTakeCard(takeCard)

	return nil, nil
}

// 出牌
func (p *PlayerCtrl) put(w http.ResponseWriter, r *http.Request, body api.PutCard) (*api.Reply, error) {

	//用户信息
	header := wrap.GetHeader(r)
	userId := header.UserId

	tb := mj.Table{}
	//判断当前回合
	ok, player := tb.TurnCheck(userId)
	if !ok {
		return nil, errors.New("非当前回合")
	}

	player.AddPutCard(body.Card)

	//通知玩家判定
	for i := 0; i < 4; i++ {
		websocketNotify("", "", 103, "", body)
	}

	//倒计时 - 是否通知下家摸牌
	traceId := countdown.NewTrackId("")
	countdown.New(traceId, 4-1, api.TurnChangeTimeOut, func(data countdown.CallData[]) {
		//有人判定，当前倒计时无效
		if data.Action == countdown.IsClose {
			return
		}
		np := tb.TurnNext()
		websocketNotify("", np.Id, 300, "", nil)
	})
	return nil, nil
}

func (p *PlayerCtrl) rewardConfirm(w http.ResponseWriter, r *http.Request, body api.RewardCard) (*api.Reply, error) {

	//用户信息
	header := wrap.GetHeader(r)
	userId := header.UserId

	//房间游戏规则
	tb := &mj.Table{}

	//校验判定事件是否开启
	mine, player := tb.TurnCheck(userId)
	reg, err := strategy.Register("")
	if err != nil {
		return nil, err
	}

	// 当前回合动作
	hs := reg.WithTurn(mine)
	handler, err := hs.ActionBy(body.Action)
	if err != nil {
		return nil, err
	}

	//判定
	ok := handler.Func(nil, tb)(player, nil, body.WithCards, body.Card)
	if !ok {
		return nil, errors.New("不支持判定")
	}
	//关闭倒计时
	countdown.Close(body.EventId)

	//抢占当前回合
	tb.TurnChange(userId)
	//添加当前牌至牌库
	player.AddRewardCards(body.WithCards, body.Card)

	//通知其他用户
	for i := 0; i < 4; i++ {
		websocketNotify("", "", 13, "", api.Empty)
	}

	//响应当前用户是否摸牌还是出牌

	return nil, nil
}

// 判定
func (p *PlayerCtrl) rewardCheck(w http.ResponseWriter, r *http.Request, u api.RewardCard) (*api.Reply, error) {

	//用户信息
	header := wrap.GetHeader(r)
	userId := header.UserId
	tb := mj.Table{}

	return nil, nil
}

// 胡
func (p *PlayerCtrl) win(w http.ResponseWriter, r *http.Request, body api.RewardCard) (*api.Reply, error) {

	//结束倒计时
	countdown.Close(body.EventId)

	//积分统计

	//通知其他玩家
	for {
		websocketNotify("", "", 1, "", api.Empty)
	}
	return nil, nil
}
