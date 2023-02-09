package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/engine"
	"sort"
	"sync"
)

// GameDefine 游戏规则
type GameDefine interface {
	// InitCtx 初始化
	InitCtx(gc *api.GameConfigure, pc *api.PaymentConfigure) engine.RoundOpsCtx
	// Handles 策略集
	Handles() map[api.RaceType]RaceEvaluate
	// Renew 从上下文中恢复
	Renew(ctx *engine.RoundCtx)
	// Finish 结束
	Finish()
	// Quit 退出
	Quit()
}

// RaceEvaluate 碰，吃，杠，胡...评估
type RaceEvaluate interface {
	// Eval 可行判断
	Eval(ctx *engine.RoundCtx, raceIdx int, tiles mj.Cards, whoIdx int, tile int) (bool, []mj.Cards)
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

func BuildProvider(roundCtx *engine.RoundCtx) GameDefine {
	gc, _ := roundCtx.HandlerCtx().Configure()
	var provider = NewProvider(gc.Mode)
	provider.Renew(roundCtx)
	return provider
}

type BaseRoundCtxHandler struct {
	lock         sync.Mutex
	gc           *api.GameConfigure
	pc           *api.PaymentConfigure
	table        *engine.Table
	tiles        map[int]*api.PlayerTiles
	profits      map[int]*api.PlayerProfits
	custom       map[string]any
	recentAction engine.RecentAction //最近数据
	recentIdx    int
	recenter     map[int]*BaseRecenter
}

func (b *BaseRoundCtxHandler) Configure() (*api.GameConfigure, *api.PaymentConfigure) {
	return b.gc, b.pc
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
	own.Hands = own.Hands.Remove(tIdx)
	own.Outs = append(own.Outs, tile)

	b.recentIdx = pIdx
	b.recentAction = engine.RecentPut
	b.recenter[pIdx].action = engine.RecentPut
	b.recenter[pIdx].put = tile

}

func (b *BaseRoundCtxHandler) AddRace(pIdx int, raceType api.RaceType, tileRaces *engine.TileRaces) {
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
	b.recentAction = engine.RecentRace
	b.recenter[pIdx].action = engine.RecentRace
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
