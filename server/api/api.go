package api

var Empty = &EmptyData{}

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
type GameAck struct {
	Event   int    `json:"event"`
	EventId string `json:"event_id"`
	RoomId  string `json:"room_id"`
}

type GameInf struct {
}

type GameConfigure struct {
	Gamer    int  `json:"gamer"`
	HasWind  bool `json:"has_wind"`
	HasOther bool `json:"has_other"`
}

// TakeCard 摸牌
type TakeCard struct {
	RoomId string `json:"room_id"`
	GameId string `json:"game_id"`
}

// PutCard 出牌
type PutCard struct {
	RoomId string `json:"room_id"`
	GameId string `json:"game_id"`
	Card   int    `json:"card"`
}

// RewardCard 判定
type RewardCard struct {
	RoomId     string `json:"room_id"`
	GameId     string `json:"game_id"`
	Event      string `json:"event"`
	PlayCards  []int  `json:"play_cards"`
	RewardCard int    `json:"reward_card"`
}

type Reply struct {
}
