package api

import "mahjong/mj"

type GameConfigure struct {
	Mode  string        `json:"mode"`  //模型
	Nums  int           `json:"nums"`  //玩家数
	Tiles []mj.CardKind `json:"tiles"` //牌库（万，筒，条，风，中，发，白）
}

const (
	AvgPayment    PaymentMode = iota + 1 //AA制
	MasterPayment                        //房主支付
	WinPayment                           //庄家支付
)

type PaymentMode int
type PaymentConfigure struct {
	Mode   PaymentMode `json:"mode"` //支付方式（AA模式，房主买单）
	Amount int         `json:"pay"`  //费用
}

// JoinRoom 加入房间
type JoinRoom struct {
	RoomId string `json:"room_id"`
}

// CreateRoom 创建房间
type CreateRoom struct {
	Game    *GameConfigure    `json:"config"`
	Payment *PaymentConfigure `json:"payment"`
}

// ExitRoom 退出房间
type ExitRoom struct {
	RoomId string `json:"room_id"`
}

// Player 玩家身份信息
type Player struct {
	Idx    int    `json:"idx"`
	AcctId string `json:"acctId"`
	Name   string `json:"name"`
	Alias  string `json:"alias"`
	Avatar string `json:"avatar"`
}

// RoomInf 房间信息
type RoomInf struct {
	RoomId  string            `json:"roomId"`
	Players map[int]*Player   `json:"players"`
	Game    *GameConfigure    `json:"game"`
	Payment *PaymentConfigure `json:"payment"`
}

type GameStart struct {
	RoomId string `json:"roomId"`
}

type TakeParameter struct {
	RoomId    string `json:"roomId"`
	Direction int    `json:"direction"`
}

type PutParameter struct {
	RoomId string `json:"roomId"`
	PutPayload
}

type RaceParameter struct {
	RoomId string `json:"roomId"`
	RacePayload
}

type AckParameter struct {
	RoomId string `json:"roomId"`
	AckPayload
}

type RacePreview struct {
	RoomId string `json:"roomId"`
	Round  int    `json:"round"`
}

type RaceEffects struct {
	Eats      [][]int       `json:"eats"`      //吃
	Pair      []int         `json:"pair"`      //碰
	OtherGang []int         `json:"otherGang"` //杠别人
	OwnGang   []int         `json:"ownGang"`   //自杠
	Win       []int         `json:"win"`       //胡
	Cao       []int         `json:"cao"`       //朝
	Tings     map[int][]int `json:"tings"`     //听
}

type RacePost struct {
	Action    string `json:"action"`    //摸牌，或出牌
	Direction int    `json:"direction"` //摸牌方向（首，尾）
}
