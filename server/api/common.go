package api

type IdentityHeader struct {
	UserId   string
	UserName string
	Token    string
}

type NoResp struct {
}

var Empty = &NoResp{}

type RaceType int

var RaceNames = map[RaceType]string{
	WinRace:  "胡",
	DDDRace:  "碰",
	ABCRace:  "吃",
	EEEERace: "杠",
	LaiRace:  "癞",
	CaoRace:  "朝",
	GuiRace:  "鬼",

	PassRace: "过",
	PutRace:  "出",
	TakeRace: "摸",
}

const (
	// WinRace 胡
	WinRace RaceType = iota + 200
	// DDDRace 碰
	DDDRace
	// ABCRace 吃
	ABCRace
	// EEEERace 杠 （别人）
	EEEERace
	// EEEEUpgradeRace 杠（碰）
	EEEEUpgradeRace
	// EEEEOwnRace 杠（自己）
	EEEEOwnRace
	// LaiRace 癞
	LaiRace
	// CaoRace 朝
	CaoRace
	// GuiRace 鬼
	GuiRace
	// PassRace 过
	PassRace RaceType = 0
	// PutRace 出
	PutRace RaceType = 1
	// TakeRace 摸
	TakeRace RaceType = 2
)

const TurnInterval = 20
