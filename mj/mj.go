package mj

type Kind string

var WanCard Kind = "wan"

//万 1 - 9
const (
	W1 = iota + 1
	W2
	W3
	W4
	W5
	W6
	W7
	W8
	W9
)

var TiaoCard Kind = "tiao"

//条 11 - 19
const (
	L1 = iota + 11
	L2
	L3
	L4
	L5
	L6
	L7
	L8
	L9
)

var TongCard Kind = "tong"

//筒 21 - 29
const (
	T1 = iota + 21
	T2
	T3
	T4
	T5
	T6
	T7
	T8
	T9
)

var WindCard Kind = "feng"

//风 31 - 37
const (
	EAST  = 31
	SOUTH = 33
	WEST  = 35
	NORTH = 37
)

var OtherCard Kind = "other"

//中发白 41 - 45
const (
	Zh = 41
	Fa = 43
	Ba = 45
)
