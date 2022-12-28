package ploy

import (
	"mahjong/mj"
	"mahjong/server/api"
	"mahjong/server/engine"
)

type BaseProvider struct {
	dice  int //骰子数
	tiles mj.Cards
}

type BaseTileHandler struct {
	table   *engine.Table
	tiles   map[int]*PlayerTiles
	profits map[int]*PlayerProfit
}

// PlayerTiles 玩家牌库
type PlayerTiles struct {
	hands      []int
	races      [][]int
	outs       []int
	lastedTake int
	lastedPut  int
}

//PlayerProfit 玩家收益
type PlayerProfit struct {
}

func (b *BaseTileHandler) GetOuts(pIdx int) []int {

	return nil
}

func (b *BaseTileHandler) GetHands(pIdx int) []int {
	return nil
}

func (b *BaseTileHandler) GetRaces(pIdx int) [][]int {
	return nil
}

func (b *BaseTileHandler) AddTake(pIdx int, tile int) {

}

func (b *BaseTileHandler) AddPut(pIdx int, tile int) {

}

func (b *BaseTileHandler) AddRace(pIdx int, tiles []int, whoIdx int, tile int) {
}

func (b *BaseTileHandler) Forward(pIdx int) int {
	return b.table.Forward()
}

func (b *BaseTileHandler) Backward(pIdx int) int {
	return b.table.Backward()
}

func (bp *BaseProvider) Init(gc *api.GameConfigure, pc *api.PaymentConfigure) engine.TileHandle {
	//创建上下文处理器
	return bp.initOps(gc, pc)
}

func (bp *BaseProvider) initOps(gc *api.GameConfigure, pc *api.PaymentConfigure) *BaseTileHandler {

	//掷骰子，洗牌，发牌
	dice := engine.NewDice()
	bp.dice = dice

	//洗牌
	tiles := engine.Shuffle(bp.tiles)

	//发牌
	tb := engine.NewTable(dice, tiles)
	members := tb.Distribution(gc.Nums)

	//添加到上下文
	opsCtx := &BaseTileHandler{
		table:   tb,
		tiles:   make(map[int]*PlayerTiles, gc.Nums),
		profits: make(map[int]*PlayerProfit, gc.Nums),
	}

	//保存牌库
	for k, v := range members {
		opsCtx.tiles[k] = &PlayerTiles{
			hands:      v,
			races:      make([][]int, 0),
			outs:       make([]int, 0),
			lastedTake: 0,
			lastedPut:  0,
		}
	}

	return opsCtx
}

func (bp *BaseProvider) Finish() bool {
	return false
}

func (bp *BaseProvider) Quit() {

}

func NewBaseProvider() GameDefine {
	return &BaseProvider{
		dice:  0,
		tiles: mj.Library,
	}
}
