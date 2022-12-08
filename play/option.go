package play

// TakeDefine 摸牌
type TakeDefine interface {
	Take() int
}

// PutDefine 出牌
type PutDefine interface {
	Put([]int)
}

// RewardDefine 奖励
type RewardDefine[Ops TakeDefine | PutDefine] interface {
	// Name 名称
	Name() (string, int)
	// IsEffect 判定
	IsEffect() bool
	// Next 后置操作
	Next() Ops
}

func ExeReward[Ops TakeDefine | PutDefine](define RewardDefine[Ops]) {
	define.Next()
}
