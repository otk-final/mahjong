package engine

import (
	"mahjong/mj"
	"mahjong/server/api"
	"sync"
)

type RoundCtx struct {
	Lock       *sync.Mutex
	round      int
	setting    *api.GameConfigure
	position   *Position
	exchanger  *Exchanger
	handlerCtx RoundOpsCtx
}

func NewRoundCtx(round int, setting *api.GameConfigure, pos *Position, exchanger *Exchanger, handlerCtx RoundOpsCtx) *RoundCtx {
	return &RoundCtx{
		Lock:       &sync.Mutex{},
		round:      round,
		setting:    setting,
		position:   pos,
		exchanger:  exchanger,
		handlerCtx: handlerCtx,
	}
}

func (ctx *RoundCtx) Configure() *api.GameConfigure {
	return ctx.setting
}

func (ctx *RoundCtx) Pos() *Position {
	return ctx.position
}
func (ctx *RoundCtx) Exchange() *Exchanger {
	return ctx.exchanger
}
func (ctx *RoundCtx) HandlerCtx() RoundOpsCtx {
	return ctx.handlerCtx
}
func (ctx *RoundCtx) Player(acctId string) (*api.Player, error) {
	return ctx.position.Index(acctId)
}

type TileRaces struct {
	Tiles     mj.Cards
	TargetIdx int
	Tile      int
}

// RoundOpsCtx 当局
type RoundOpsCtx interface {
	GetTiles(pIdx int) *api.PlayerTiles

	GetProfits(pIdx int) *api.PlayerProfits

	AddTake(pIdx int, tile int)

	AddPut(pIdx int, tile int)

	AddRace(pIdx int, raceType api.RaceType, target *TileRaces)

	Forward(pIdx int) int

	Backward(pIdx int) int

	Remained() int

	RecentAction() RecentAction

	RecentIdx() int

	Recenter(targetIdx int) RoundOpsRecent
}

type RecentAction int

const (
	RecentPut RecentAction = iota + 1
	RecentTake
	RecentRace
)

type RoundOpsRecent interface {
	Idx() int
	Put() int
	Take() int
	Race() *TileRaces
	Action() RecentAction
}
