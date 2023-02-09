package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	engine2 "mahjong/service/engine"
	"sort"
	"sync"
)

// GameDefine 游戏规则
type GameDefine interface {
	// InitOperation 初始化
	InitOperation(setting *api.GameConfigure) engine2.RoundOperation
	// Handles 策略集
	Handles() map[api.RaceType]RaceEvaluator
	// Renew 从上下文中恢复
	Renew(ctx *engine2.RoundCtx) GameDefine
	// Finish 结束
	Finish()
	// Quit 退出
	Quit()
}

// RaceEvaluator 碰，吃，杠，胡...评估
type RaceEvaluator interface {
	// Eval 可行判断
	Eval(ctx *engine2.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards)
}

func NewProvider(mode string) GameDefine {
	switch mode {
	case "std": //标准
		return newBaseProvider()
	case "LaiCollect": //赖子
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

func RenewProvider(ctx *engine2.RoundCtx) GameDefine {
	return NewProvider(ctx.Configure().Mode).Renew(ctx)
}

type BaseRoundCtxHandler struct {
	lock         sync.Mutex
	setting      *api.GameConfigure
	table        *engine2.Table
	tiles        map[int]*api.PlayerTiles
	profits      map[int]*api.PlayerProfits
	recentAction engine2.RecentAction //最近数据
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
	b.recentAction = engine2.RecentTake
	b.recenter[pIdx].action = engine2.RecentTake
	b.recenter[pIdx].take = tile
}

func (b *BaseRoundCtxHandler) AddPut(pIdx int, tile int) {
	defer b.lock.Unlock()
	b.lock.Lock()

	own := b.tiles[pIdx]

	//update hands
	tIdx := own.Hands.Index(tile)
	own.Hands = own.Hands.Remove(tIdx)
	own.Outs = append(own.Outs, tile)

	b.recentIdx = pIdx
	b.recentAction = engine2.RecentPut
	b.recenter[pIdx].action = engine2.RecentPut
	b.recenter[pIdx].put = tile

}

func (b *BaseRoundCtxHandler) AddRace(pIdx int, raceType api.RaceType, tileRaces *engine2.TileRaces) {
	defer b.lock.Unlock()
	b.lock.Lock()

	own := b.tiles[pIdx]
	//移除手上的牌
	for _, t := range tileRaces.Tiles {
		tIdx := own.Hands.Index(t)
		own.Hands = own.Hands.Remove(tIdx)
	}

	if raceType == api.EEEEUpgradeRace {
		//杠（碰）覆盖原有的碰
		for i := 0; i < len(own.Races); i++ {
			ddd := own.Races[i].Indexes(tileRaces.Tile)
			if len(ddd) != 3 {
				continue
			}
			own.Races[i] = append(own.Races[i], tileRaces.Tile)
		}
	} else if raceType == api.ABCRace || raceType == api.DDDRace {
		//吃，碰 别人

		//合并
		races := append(tileRaces.Tiles, tileRaces.Tile)
		sort.Ints(races)
		own.Races = append(own.Races, races)

		//移交
		who := b.tiles[tileRaces.TargetIdx]
		who.Outs = who.Outs[:len(who.Outs)-1]
	} else {
		//添加
		own.Races = append(own.Races, tileRaces.Tiles)
	}

	b.recentIdx = pIdx
	b.recentAction = engine2.RecentRace
	b.recenter[pIdx].action = engine2.RecentRace
	b.recenter[pIdx].race = tileRaces
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

func (b *BaseRoundCtxHandler) RecentAction() engine2.RecentAction {
	return b.recentAction
}

func (b *BaseRoundCtxHandler) RecentIdx() int {
	return b.recentIdx
}

func (b *BaseRoundCtxHandler) Recenter(targetIdx int) engine2.RoundOpsRecent {
	return b.recenter[targetIdx]
}

type BaseRecenter struct {
	idx    int
	put    int
	take   int
	race   *engine2.TileRaces
	action engine2.RecentAction
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

func (r *BaseRecenter) Race() *engine2.TileRaces {
	return r.race
}

func (r *BaseRecenter) Action() engine2.RecentAction {
	return r.action
}
