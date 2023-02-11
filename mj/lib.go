package mj

import "sort"

type Card int
type Cards []int

var mjKindRange = map[Kind][]int{
	WanCard:   {W1, W9},
	TiaoCard:  {L1, L9},
	TongCard:  {T1, T9},
	WindCard:  {EAST, SOUTH, WEST, NORTH},
	OtherCard: {Zh, Fa, Ba},
}

// LoadLibrary 指定牌库
func LoadLibrary(kinds ...Kind) []int {

	//全量
	if len(kinds) == 0 {
		kinds = []Kind{WanCard, TiaoCard, TongCard, WindCard}
	}

	newLib := make([]int, 0)
	for _, choose := range kinds {
		mjRange := mjKindRange[choose]
		if choose == WindCard || choose == OtherCard {
			//风，中，发，白
			for i := 0; i < len(mjRange); i++ {
				newLib = append(newLib, mjRange[i], mjRange[i], mjRange[i], mjRange[i])
			}
		} else {
			//万，筒，条
			for i := mjRange[0]; i <= mjRange[1]; i++ {
				newLib = append(newLib, i, i, i, i)
			}
		}
	}
	return newLib
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

func (c Cards) Equal(dest Cards) bool {
	if len(c) != len(dest) {
		return false
	}
	sort.Ints(c)
	sort.Ints(dest)
	for i := 0; i < len(c); i++ {
		if c[i] != dest[i] {
			return false
		}
	}
	return true
}

func (c Cards) IsDDD() bool {
	if len(c) != 3 {
		return false
	}
	return c[0] == c[1] && c[1] == c[2]
}

func (c Cards) IsEEEE() bool {
	if len(c) != 4 {
		return false
	}
	return c[0] == c[1] && c[1] == c[2] && c[2] == c[3]
}

func (c Cards) IsABC() bool {
	if len(c) != 3 {
		return false
	}
	sort.Ints(c)
	return c[0]+1 == c[1] && c[1]+1 == c[2]
}
