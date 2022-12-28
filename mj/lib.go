package mj

// Library 牌库
var Library = Cards{
	//筒
	T1, T1, T1, T1,
	T2, T2, T2, T2,
	T3, T3, T3, T3,
	T4, T4, T4, T4,
	T5, T5, T5, T5,
	T6, T6, T6, T6,
	T7, T7, T7, T7,
	T8, T8, T8, T8,
	T9, T9, T9, T9,
	//万
	W1, W1, W1, W1,
	W2, W2, W2, W2,
	W3, W3, W3, W3,
	W4, W4, W4, W4,
	W5, W5, W5, W5,
	W6, W6, W6, W6,
	W7, W7, W7, W7,
	W8, W8, W8, W8,
	W9, W9, W9, W9,
	//条
	L1, L1, L1, L1,
	L2, L2, L2, L2,
	L3, L3, L3, L3,
	L4, L4, L4, L4,
	L5, L5, L5, L5,
	L6, L6, L6, L6,
	L7, L7, L7, L7,
	L8, L8, L8, L8,
	L9, L9, L9, L9,
	//中发白
	Zh, Zh, Zh, Zh,
	Ba, Ba, Ba, Ba,
	Fa, Fa, Fa, Fa,
	//东南西北
	EAST, EAST, EAST, EAST,
	SOUTH, SOUTH, SOUTH, SOUTH,
	WEST, WEST, WEST, WEST,
	NORTH, NORTH, NORTH, NORTH,
}

type Card int
type Cards []int

var CardRangeMap = map[CardKind][]int{
	WanCard:   {W1, W9},
	TiaoCard:  {L1, L9},
	TongCard:  {T1, T9},
	WindCard:  {EAST, NORTH},
	OtherCard: {Zh, Ba},
}

// LoadLibrary 指定牌库
func LoadLibrary(kinds ...CardKind) []int {
	newLib := make([]int, 0)

	//copy
	temp := make([]int, 0)
	copy(temp, Library)
	filter := func(kind CardKind) bool {
		for _, k := range kinds {
			if k == kind {
				return true
			}
		}
		return false
	}

	//filter
	for _, tile := range temp {
		kind := Card(tile).Kind()
		if !filter(kind) {
			continue
		}
		limit := CardRangeMap[kind]
		if limit[0] <= tile && tile <= limit[1] {
			newLib = append(newLib, tile)
		}
	}
	return newLib
}

//相邻的牌 只针对条，万，筒
func (c Card) getNeighbors() []Cards {
	if c > 29 {
		return nil
	}
	nb := make([]Cards, 0)
	//默认相邻
	nb = append(nb, []int{int(c + 1), int(c + 2)})
	nb = append(nb, []int{int(c - 2), int(c - 1)})
	nb = append(nb, []int{int(c - 1), int(c + 1)})

	// 一万，一条，一筒
	if c == T1 || c == W1 || c == L1 {
		return nb[:1]
	}
	// 九万，九条，九筒
	if c == T9 || c == W9 || c == L9 {
		return nb[1:2]
	}
	// 其他
	return nb
}

func (c Card) Kind() CardKind {
	for k, v := range CardRangeMap {
		if v[0] <= int(c) && int(c) <= v[1] {
			return k
		}
	}
	return ""
}

//相同牌
func sameCard(c Cards, mj int, match int) bool {
	count := 0
	for _, k := range c {
		if k == mj {
			count++
		}
	}
	return count >= match
}

// HasPair 碰？
func (c Cards) HasPair(mj int) bool {
	return sameCard(c, mj, 2)
}

// HasGang 杠？
func (c Cards) HasGang(mj int) bool {
	return sameCard(c, mj, 3)
}

// HasList 吃？
func (c Cards) HasList(mj int) bool {
	if mj > 29 {
		return false
	}
	// 万筒条 相邻牌
	nbs := Card(mj).getNeighbors()
	for _, nb := range nbs {
		if sameCard(c, nb[0], 1) && sameCard(c, nb[1], 1) {
			return true
		}
	}
	// 同时存在
	return false
}
