package scene

// HandCards 手上的牌
type HandCards []int

// TurnCards 回合判定的牌
type TurnCards [][]int

// ThrowCards 打出的牌
type ThrowCards []int

// 标准 nAAA+mABC+DD

type StanderPloy struct {
}

func (s *StanderPloy) IsWin(mj int) bool {
	//TODO implement me
	panic("implement me")
}

func (s *StanderPloy) AtWin() []int {
	//TODO implement me
	panic("implement me")
}
