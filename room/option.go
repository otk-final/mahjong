package room

// AssertConfig 判定规则
type AssertConfig struct {
	CanPair bool //能碰？
	CanList bool //能吃？
	CanGang bool //能杠？
	CanTing bool //能听？
	CanWin  bool //能自摸？
}

// PlayConfig 玩法
type PlayConfig struct {
	MinPlayer   int   //最小玩家数
	CardLibrary []int //牌库
}
