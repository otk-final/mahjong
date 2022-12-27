package store

// RoundDefine 当局状态
type RoundDefine interface {
	// LaiTile 赖子牌
	LaiTile() int
	// OutTiles 已出牌
	OutTiles(pIdx int) []int
	// HandTiles 手上牌
	HandTiles(pIdx int) []int
	// RaceTiles 生效牌
	RaceTiles(pIdx int) [][]int
}
