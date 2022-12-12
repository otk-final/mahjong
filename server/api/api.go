package api

var Empty = &EmptyData{}
var TurnChangeTimeOut int = 30

type UserHeader struct {
	UserId string
	Token  string
}

type EmptyData struct {
}

type Identity struct {
	UserId   string
	Token    string
	UserName string
}

// JoinRoom 加入房间
type JoinRoom struct {
	RoomId string `json:"room_id"`
}

// CreateRoom 创建房间
type CreateRoom struct {
	Config *GameConfigure `json:"config"`
}

// ExitRoom 退出房间
type ExitRoom struct {
	RoomId string `json:"room_id"`
}

type RoomInf struct {
	RoomId  string
	Players map[int]Identity
	Config  *GameConfigure
}

type GameRun struct {
	RoomId string `json:"room_id"`
}
type GameReadyAck struct {
	EventId string `json:"event_id"`
	RoomId  string `json:"room_id"`
}

type GameInf struct {
}

type GameConfigure struct {
	Mode     string
	Players  int  `json:"players"`
	HasWind  bool `json:"has_wind"`
	HasOther bool `json:"has_other"`
}

// TakeCard 摸牌
type TakeCard struct {
	RoomId    string `json:"room_id"`
	GameId    string `json:"game_id"`
	Direction int    `json:"direction"`
}

// PutCard 出牌
type PutCard struct {
	RoomId string `json:"room_id"`
	GameId string `json:"game_id"`
	Card   int    `json:"card"`
}

// RewardCard 判定
type RewardCard struct {
	RoomId    string `json:"room_id"`
	GameId    string `json:"game_id"`
	EventId   string `json:"event_id"`
	Action    string
	WithCards []int `json:"with_cards"`
	Card      int   `json:"card"`
}

type Reply struct {
}
