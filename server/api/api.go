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

type Roboter struct {
	*Player
	Level int
}

// RoomInf 房间信息
type RoomInf struct {
	RoomId  string         `json:"roomId"`
	Own     *Player        `json:"own"`
	Players []*Player      `json:"players"`
	Begin   bool           `json:"begin"`
	Config  *GameConfigure `json:"config"`
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
	*PlayerTiles
	Take     int           `json:"take"`
	Remained int           `json:"remained"`
	Options  []*RaceOption `json:"options"`
}
type PutParameter struct {
	*PutPayload
	RoomId string `json:"roomId"`
}
type PutResult struct {
	*PlayerTiles
	Put int `json:"put"`
}
type RaceParameter struct {
	RoomId   string   `json:"roomId"`
	Round    int      `json:"round"`
	RaceType RaceType `json:"raceType"`
	Tiles    mj.Cards `json:"tiles"`
}
type RaceResult struct {
	*PlayerTiles
	ContinueTake int           `json:"continueTake"`
	Target       int           `json:"target"`
	TargetTile   int           `json:"targetTile"`
	Options      []*RaceOption `json:"options"`
}
type RaceOption struct {
	RaceType RaceType   `json:"raceType"`
	Tiles    []mj.Cards `json:"tiles"`
}
type RacePreview struct {
	RoomId string `json:"roomId"`
	Round  int    `json:"round"`
	AckId  int    `json:"ackId"`
	Target int    `json:"target"`
	Tile   int    `json:"tile"`
}

type RaceEffects struct {
	Options []*RaceOption `json:"options"`
}

type WinParameter struct {
	RoomId string `json:"roomId"`
	Round  int    `json:"round"`
}
type WinResult struct {
	*WinPayload
}

type AckParameter struct {
	RoomId string `json:"roomId"`
	Round  int    `json:"round"`
}

type GameQuery struct {
	RoomId string `json:"roomId"`
}

type GameInf struct {
	*GamePayload
	RoomId  string        `json:"roomId"`
	Options []*RaceOption `json:"options"`
}

type RobotParameter struct {
	RoomId string `json:"roomId"`
	Open   bool   `json:"open"`
	Level  int    `json:"level"`
}

// PlayerTiles 玩家牌库
type PlayerTiles struct {
	Idx   int        `json:"idx"`
	Hands mj.Cards   `json:"hands"`
	Races []mj.Cards `json:"races"`
	Outs  mj.Cards   `json:"outs"`
}

func (tile *PlayerTiles) ExplicitCopy(explicit bool) *PlayerTiles {

	tempRaces := make([]mj.Cards, 0)
	for _, comb := range tile.Races {
		tempRaces = append(tempRaces, comb.Clone())
	}
	temp := &PlayerTiles{
		Idx:   tile.Idx,
		Hands: tile.Hands.Clone(),
		Races: tempRaces,
		Outs:  tile.Outs.Clone(),
	}
	if explicit {
		return temp
	}

	//屏蔽手牌数据
	temp.Hands = make(mj.Cards, len(temp.Hands))
	return temp
}

//PlayerProfits 玩家收益
type PlayerProfits struct {
}
