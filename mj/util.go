package mj

import (
	"sort"
	"strings"
)

var defaultABCMapperAsc = [][]int{
	{1, 2, 3},
	{2, 3, 4},
	{3, 4, 5},
	{4, 5, 6},
	{5, 6, 7},
	{6, 7, 8},
	{7, 8, 9},
}

var defaultABCMapperDesc = [][]int{
	{9, 8, 7},
	{8, 7, 6},
	{7, 6, 5},
	{6, 5, 4},
	{5, 4, 3},
	{4, 3, 2},
	{3, 2, 1},
}

func sameCardGroup(sameCards []int) map[int][]int {
	// 分组
	var cardGroup = make(map[int][]int, 0)
	for _, card := range sameCards {
		cs, ok := cardGroup[card]
		if !ok {
			cs = make([]int, 0)
		}
		cs = append(cs, card)
		cardGroup[card] = cs
	}
	return cardGroup
}

func NewABCMapper(kind CardKind, order string) [][]int {
	rate := 0
	if kind == WanCard { //万 1 - 9
		rate = W1 - 1
	}
	if kind == TiaoCard { //条 11 - 19
		rate = L1 - 1
	}
	if kind == TongCard { //筒 21 - 29
		rate = T1 - 1
	}

	orderMapper := defaultABCMapperAsc
	if strings.EqualFold(order, "DESC") {
		orderMapper = defaultABCMapperDesc
	}

	newMapper := make([][]int, 0)
	for _, mapper := range orderMapper {
		newMapper = append(newMapper, []int{mapper[0] + rate, mapper[1] + rate, mapper[2] + rate})
	}
	return newMapper
}

func FindABC(abcMapper [][]int, sameCards []int) ([][]int, []int) {

	// 分组
	var cardGroup = sameCardGroup(sameCards)
	var matrix = make([][]int, 0)
	// 过滤
	for _, mapper := range abcMapper {
		a, b, c := mapper[0], mapper[1], mapper[2]
		amap, aok := cardGroup[a]
		if !aok {
			amap = make([]int, 0)
		}
		bmap, bok := cardGroup[b]
		if !bok {
			bmap = make([]int, 0)
		}
		cmap, cok := cardGroup[c]
		if !cok {
			cmap = make([]int, 0)
		}

		//取最小集
		sortLens := []int{len(amap), len(bmap), len(cmap)}
		//取最小数
		sort.Ints(sortLens)
		minLen := sortLens[0]
		//符合数据
		for i := 0; i < minLen; i++ {
			matrix = append(matrix, []int{a, b, c})
		}
		cardGroup[a] = amap[minLen:]
		cardGroup[b] = bmap[minLen:]
		cardGroup[c] = cmap[minLen:]
	}

	var other = make([]int, 0)
	for _, v := range cardGroup {
		other = append(other, v...)
	}

	return matrix, other
}

func FindDDD(sameCards []int) ([][]int, []int) {
	var cardGroup = sameCardGroup(sameCards)
	var matrix = make([][]int, 0)
	var other = make([]int, 0)
	for k, v := range cardGroup {
		vLen := len(v)
		if vLen == 3 {
			//三张
			matrix = append(matrix, []int{k, k, k})
		} else if vLen == 4 {
			//四张
			matrix = append(matrix, []int{k, k, k})
			other = append(other, []int{k}...)
		} else {
			other = append(other, v...)
		}
	}
	return matrix, other
}

func FindEE(sameCards []int) ([][]int, []int) {
	var cardGroup = sameCardGroup(sameCards)
	var matrix = make([][]int, 0)
	var other = make([]int, 0)
	for k, v := range cardGroup {
		if len(v) == 2 {
			matrix = append(matrix, []int{k, k})
		} else {
			other = append(other, v...)
		}
	}
	return matrix, other
}

func FilterCards(kind CardKind, cards []int) []int {
	begin, end := 0, 0
	switch kind {
	case WanCard:
		begin, end = W1, W9
	case TiaoCard:
		begin, end = L1, L9
	case TongCard:
		begin, end = T1, T9
	case WindCard:
		begin, end = EAST, NORTH
	case OtherCard:
		begin, end = Zh, Ba
	}
	filters := make([]int, 0)
	for _, c := range cards {
		if begin <= c && c <= end {
			filters = append(filters, c)
		}
	}
	return filters
}
