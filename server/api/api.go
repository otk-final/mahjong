package api

type Identity struct {
	UserId   string
	Token    string
	UserName string
}

// 加入房间
type JoinRoom struct {
}

// 创建房间
type CreateRoom struct {
}

// 退出房间
type ExitRoom struct {
}

type GameRun struct {
	RoomId string
}

type GameConfigure struct {
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

type PlayerAction struct {
	RoomId    string `json:"room_id"`
	GameId    string `json:"game_id"`
	Player    int    `json:"player"`
	Card      int    `json:"card"`
	WithCards []int  `json:"with_cards"`
}

// PairCard 碰
type PairCard struct {
	PlayerAction
}

// GangCard 杠
type GangCard struct {
	PlayerAction
}

// EatCard 吃
type EatCard struct {
	PlayerAction
}

type Reply struct {
	AWait int
	Event string
}
