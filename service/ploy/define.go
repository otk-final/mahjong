package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/service/engine"
	"sort"
	"sync"
)

// GameDefine 游戏规则
type GameDefine interface {
	// Extras 特殊牌
	Extras() []*mj.CardExtra
	// CanPut 可以除的牌
	CanPut(pIdx int, tile int) bool
	// InitOperation 初始化
	InitOperation(setting *api.GameConfigure) engine.RoundOperation
	// Handles 策略集
	Handles() map[api.RaceType]RaceEvaluator
	// Renew 从上下文中恢复
	Renew(ctx *engine.RoundCtx) GameDefine
	// Finish 结束
	Finish()
	// Quit 退出
	Quit()
}

type RaceNext int

const (
	NextQuit RaceNext = 0
	NextTake RaceNext = 1
	NextPut  RaceNext = 2
)

// RaceEvaluator 碰，吃，杠，胡...评估
type RaceEvaluator interface {
	// Valid 验证
	Valid(ctx *engine.RoundCtx, raceIdx int, puts mj.Cards, whoIdx int, whoTile int) bool
	// Eval 可行判断
	Eval(ctx *engine.RoundCtx, raceIdx int, hands mj.Cards, whoIdx int, whoTile int) (bool, []mj.Cards)
	// Next 后置事件
	Next(ctx *engine.RoundCtx, raceIdx int, whoIdx int) RaceNext
}

func NewProvider(mode string) GameDefine {
	switch mode {
	case "std": //标准
		return newBaseProvider()
	case "lai": //赖子
		return newLaiProvider()
	case "k5x": //卡5星
		break
	case "7d": //七对
		break
	case "sc": //四川
		break
	case "gz": //广东
		break
	}
	return nil
}

func RenewProvider(ctx *engine.RoundCtx) GameDefine {
	return NewProvider(ctx.Configure().Mode).Renew(ctx)
}

type BaseRoundCtxHandler struct {
	lock         sync.Mutex
	setting      *api.GameConfigure
	table        *engine.Table
	tiles        map[int]*api.PlayerTiles
	profits      map[int]*api.PlayerProfits
	recentAction engine.RecentAction //最近数据
	recentIdx    int
	recenter     map[int]*BaseRecenter
}

func (b *BaseRoundCtxHandler) GetTiles(pIdx int) *api.PlayerTiles {
	return b.tiles[pIdx]
}

func (b *BaseRoundCtxHandler) GetProfits(pIdx int) *api.PlayerProfits {
	return b.profits[pIdx]
}

func (b *BaseRoundCtxHandler) withHands(pIdx int) mj.Cards {
	return b.tiles[pIdx].Hands
}
func (b *BaseRoundCtxHandler) withRaces(pIdx int) []mj.Cards {
	return b.tiles[pIdx].Races
}
func (b *BaseRoundCtxHandler) withOuts(pIdx int) mj.Cards {
	return b.tiles[pIdx].Outs
}

func (b *BaseRoundCtxHandler) AddTake(pIdx int, tile int) {
	defer b.lock.Unlock()
	b.lock.Lock()

	own := b.tiles[pIdx]

	own.Hands = append(own.Hands, tile)

	b.recentIdx = pIdx
	b.recentAction = engine.RecentTake
	b.recenter[pIdx].action = engine.RecentTake
	b.recenter[pIdx].take = tile
}

func (b *BaseRoundCtxHandler) AddPut(pIdx int, tile int) {
	defer b.lock.Unlock()
	b.lock.Lock()

	own := b.tiles[pIdx]

	//update hands
	tIdx := own.Hands.Index(tile)
	if tIdx != -1 {
		own.Hands = own.Hands.Remove(tIdx)
	}
	own.Outs = append(own.Outs, tile)

	b.recentIdx = pIdx
	b.recentAction = engine.RecentPut
	b.recenter[pIdx].action = engine.RecentPut
	b.recenter[pIdx].put = tile

}

func (b *BaseRoundCtxHandler) AddRace(pIdx int, raceType api.RaceType, tileRaces *engine.TileRaces) mj.Cards {
	defer b.lock.Unlock()
	b.lock.Lock()

	own := b.tiles[pIdx]
	//移除自己手上的牌
	for _, t := range tileRaces.Tiles {
		tIdx := own.Hands.Index(t)
		if tIdx == -1 {
			continue
		}
		own.Hands = own.Hands.Remove(tIdx)
	}

	//合并判定牌
	var raceIntact mj.Cards
	if pIdx == tileRaces.TargetIdx {
		//自己的牌
		raceIntact = tileRaces.Tiles
	} else {
		//别人的牌
		raceIntact = append(tileRaces.Tiles, tileRaces.Tile)
		sort.Ints(raceIntact)

		//移除目标out
		who := b.tiles[tileRaces.TargetIdx]
		lastedIdx := len(who.Outs) - 1
		who.Outs[lastedIdx] = who.Outs[lastedIdx] * -1
	}

	//如果是碰后再杠，则替换原数据
	if raceType == api.EEEEUpgradeRace {
		for i := 0; i < len(own.Races); i++ {
			raceItem := own.Races[i]
			if raceItem.IsEEEEUpgrade(raceIntact[0]) {
				own.Races[i] = append(raceItem, raceIntact[0])
				break
			}
		}
	} else {
		own.Races = append(own.Races, raceIntact)
	}

	b.recentIdx = pIdx
	b.recentAction = engine.RecentRace
	b.recenter[pIdx].action = engine.RecentRace
	b.recenter[pIdx].race = tileRaces

	return raceIntact
}

func (b *BaseRoundCtxHandler) Forward(pIdx int) int {
	return b.table.Forward()
}

func (b *BaseRoundCtxHandler) Backward(pIdx int) int {
	return b.table.Backward()
}

func (b *BaseRoundCtxHandler) Remained() int {
	return b.table.Remains()
}

func (b *BaseRoundCtxHandler) RecentAction() engine.RecentAction {
	return b.recentAction
}

func (b *BaseRoundCtxHandler) RecentIdx() int {
	return b.recentIdx
}

func (b *BaseRoundCtxHandler) Recenter(targetIdx int) engine.RoundOpsRecent {
	return b.recenter[targetIdx]
}

type BaseRecenter struct {
	idx    int
	put    int
	take   int
	race   *engine.TileRaces
	action engine.RecentAction
}

func (r *BaseRecenter) Idx() int {
	return r.idx
}

func (r *BaseRecenter) Put() int {
	return r.put
}

func (r *BaseRecenter) Take() int {
	return r.take
}

func (r *BaseRecenter) Race() *engine.TileRaces {
	return r.race
}

func (r *BaseRecenter) Action() engine.RecentAction {
	return r.action
}
