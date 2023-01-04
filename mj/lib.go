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
	temp := make([]int, len(Library))
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

func (c Card) Kind() CardKind {
	for k, v := range CardRangeMap {
		if v[0] <= int(c) && int(c) <= v[1] {
			return k
		}
	}
	return NilCard
}

func (c Cards) Remove(index ...int) Cards {
	temp := c.Clone()
	//占位
	for _, idx := range index {
		temp[idx] = -1
	}

	remain := make(Cards, 0)
	for i := 0; i < len(temp); i++ {
		t := temp[i]
		if t == -1 {
			continue
		}
		remain = append(remain, t)
	}
	return remain
}

func (c Cards) Index(mj int) int {
	for i := 0; i < len(c); i++ {
		if c[i] == mj {
			return i
		}
	}
	return -1
}

func (c Cards) Indexes(mj int) []int {
	is := make([]int, 0)
	for i := 0; i < len(c); i++ {
		if c[i] == mj {
			is = append(is, i)
		}
	}
	return is
}

func (c Cards) Replace(target []int, dest int) (int, []int) {

	targetCards := Cards(target)
	newCards := c.Clone()

	exist := 0
	//替换
	for i := 0; i < len(newCards); i++ {
		tile := newCards[i]
		if targetCards.Index(tile) != -1 {
			newCards[i] = dest
			exist++
		}
	}
	return exist, newCards
}

func (c Cards) Clone() Cards {
	dest := make(Cards, len(c))
	copy(dest, c)
	return dest
}
