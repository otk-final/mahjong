package mj

import "sort"

type WinFilterFunc func(winComb *WinComb, temp Cards) Cards

func FilterABCAsc(winComb *WinComb, temp Cards) Cards {
	//正序过滤
	sort.Ints(temp)
	i := 0
	for i < len(temp) {
		t := temp[i]
		//顺子
		t1, t2 := temp.Index(t+1), temp.Index(t+2)
		if t1 != -1 && t2 != -1 {
			temp = temp.Remove(i, t1, t2)
			//reset
			winComb.ABC = append(winComb.ABC, Cards{t, t + 1, t + 2})
			i = 0
			continue
		}
		i++
	}
	return temp
}

func FilterABCDesc(winComb *WinComb, temp Cards) Cards {
	//倒叙过滤
	sort.Sort(sort.Reverse(sort.IntSlice(temp)))
	i := 0
	for i < len(temp) {
		t := temp[i]
		//顺子
		t1, t2 := temp.Index(t-1), temp.Index(t-2)
		if t1 != -1 && t2 != -1 {
			temp = temp.Remove(i, t1, t2)
			//reset
			winComb.ABC = append(winComb.ABC, Cards{t - 2, t - 1, t})
			i = 0
			continue
		}
		i++
	}
	return temp
}

func FilterDDD(winComb *WinComb, temp Cards) Cards {
	//刻子
	i := 0
	sort.Ints(temp)
	for i < len(temp) {
		t := temp[i]
		if len(temp) > i+2 && t == temp[i+1] && t == temp[i+2] {
			temp = temp.Remove(i, i+1, i+2)
			//reset
			winComb.DDD = append(winComb.DDD, Cards{t, t, t})
			i = 0
			continue
		}
		i++
	}
	return temp
}

type WinChecker struct {
	Filters [][]WinFilterFunc
}
type WinComb struct {
	ABC []Cards
	DDD []Cards
	EE  Cards
}

func NewWinChecker() *WinChecker {
	//判断方案 ABC*n + DDD *m + EE * 1
	return &WinChecker{Filters: [][]WinFilterFunc{
		{FilterABCAsc, FilterDDD},
		{FilterABCDesc, FilterDDD},
		{FilterDDD, FilterABCAsc},
		{FilterDDD, FilterABCDesc},
	}}
}

func (win *WinChecker) Check(data Cards) (bool, *WinComb) {

	for _, plans := range win.Filters {
		tiles := make(Cards, len(data))
		copy(tiles, data)

		//缓存结果
		out := &WinComb{
			ABC: make([]Cards, 0),
			DDD: make([]Cards, 0),
			EE:  make(Cards, 0),
		}
		for _, plan := range plans {
			tiles = plan(out, tiles)
		}

		//将牌判断
		if len(tiles) == 2 && tiles[0] == tiles[1] {
			out.EE = tiles
			return true, out
		}
	}
	return false, nil
}
