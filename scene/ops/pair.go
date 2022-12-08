package ops

import "mahjong/play"

const RewardPairName = "碰"
const RewardPairCode = 1

// RewardPair 碰
type RewardPair struct {
}

func (r RewardPair) Name() (string, int) {
	//TODO implement me
	panic("implement me")
}

func (r RewardPair) IsEffect() bool {
	//TODO implement me
	panic("implement me")
}

func (r RewardPair) Next() TakeHead {
	return TakeHead{}
}

func aaa() {
	play.ExeReward[TakeHead](RewardPair{})
}
