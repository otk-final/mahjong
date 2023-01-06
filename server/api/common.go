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
}

const (
	// WinRace 胡
	WinRace RaceType = iota + 1
	// DDDRace 碰
	DDDRace
	// ABCRace 吃
	ABCRace
	// EEEERace 杠
	EEEERace
	// LaiRace 癞
	LaiRace
	// CaoRace 朝
	CaoRace
	// GuiRace 鬼
	GuiRace
)
