package api

import (
	"mahjong/mj"
)

type GameConfigure struct {
	Mode   string         `json:"mode"`   //模型
	Nums   int            `json:"nums"`   //玩家数
	Custom map[string]any `json:"custom"` //自定义参数
}

const (
	AvgPayment    PaymentMode = iota + 1 //AA制
	MasterPayment                        //房主支付
	WinPayment                           //庄家支付
)

type PaymentMode int
type PaymentConfigure struct {
	Mode   PaymentMode `json:"mode"`   //支付方式（AA模式，房主买单）
	Amount int         `json:"amount"` //费用
}

// JoinRoom 加入房间
type JoinRoom struct {
	RoomId string `json:"roomId"`
}

// CreateRoom 创建房间
type CreateRoom struct {
	Game    *GameConfigure    `json:"game"`
	Payment *PaymentConfigure `json:"payment"`
}

// ExitRoom 退出房间
type ExitRoom struct {
	RoomId string `json:"room_id"`
}

// Player 玩家身份信息
type Player struct {
	Idx    int    `json:"idx"`
	UId    string `json:"uid"`
	UName  string `json:"uname"`
	Alias  string `json:"alias"`
	Ip     string `json:"ip"`
	Avatar string `json:"avatar"`
}

// RoomInf 房间信息
type RoomInf struct {
	RoomId  string            `json:"roomId"`
	Own     *Player           `json:"own"`
	Players []*Player         `json:"players"`
	Begin   bool              `json:"begin"`
	Game    *GameConfigure    `json:"game"`
	Payment *PaymentConfigure `json:"payment"`
}

type GameParameter struct {
	RoomId string `json:"roomId"`
	Ploy   string `json:"ploy"`
}

type TakeParameter struct {
	RoomId    string `json:"roomId"`
	Round     int    `json:"round"`
	Direction int    `json:"direction"`
}
type TakeResult struct {
	Tile     int               `json:"tile"`
	Remained int               `json:"remained"`
	Usable   []*UsableRaceItem `json:"usable"`
}

type PutParameter struct {
	RoomId string `json:"roomId"`
	PutPayload
}
type PutResult struct {
	RoomId string   `json:"roomId"`
	Tile   int      `json:"tile"`
	Hands  mj.Cards `json:"hands"`
}

type RaceParameter struct {
	RoomId string `json:"roomId"`
	RacePayload
}

type AckParameter struct {
	RoomId string `json:"roomId"`
	Round  int    `json:"round"`
	AckPayload
}

type RacePreview struct {
	RoomId string `json:"roomId"`
	Round  int    `json:"round"`
	AckId  int    `json:"ackId"`
	Target int    `json:"target"`
	Tile   int    `json:"tile"`
}

type UsableRaceItem struct {
	RaceType RaceType   `json:"raceType"`
	Tiles    []mj.Cards `json:"tiles"`
}

type RaceEffects struct {
	Usable []*UsableRaceItem `json:"usable"`
}

type RaceResult struct {
	Hands  mj.Cards          `json:"hands"`
	Tiles  mj.Cards          `json:"tiles"`
	Tile   int               `json:"tile"`
	Usable []*UsableRaceItem `json:"usable"`
}

type GameQuery struct {
	RoomId string `json:"roomId"`
}

type GameInf struct {
	GamePayload
	RoomId string            `json:"roomId"`
	Usable []*UsableRaceItem `json:"usable"`
}

// PlayerTiles 玩家牌库
type PlayerTiles struct {
	Idx        int        `json:"idx"`
	Hands      mj.Cards   `json:"hands"`
	Races      []mj.Cards `json:"races"`
	Outs       mj.Cards   `json:"outs"`
	LastedTake int        `json:"lastedTake"`
	LastedPut  int        `json:"lastedPut"`
}

func (tile *PlayerTiles) ExplicitCopy(explicit bool) *PlayerTiles {

	tempRaces := make([]mj.Cards, 0)
	for _, comb := range tile.Races {
		tempRaces = append(tempRaces, comb.Clone())
	}
	temp := &PlayerTiles{
		Idx:        tile.Idx,
		Hands:      tile.Hands.Clone(),
		Races:      tempRaces,
		Outs:       tile.Outs.Clone(),
		LastedTake: tile.LastedTake,
		LastedPut:  tile.LastedPut,
	}
	if explicit {
		return temp
	}

	//屏蔽数据
	temp.Hands = make(mj.Cards, len(temp.Hands))
	temp.LastedTake = 0
	temp.LastedPut = 0
	return temp
}

//PlayerProfits 玩家收益
type PlayerProfits struct {
}
