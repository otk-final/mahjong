package mj

type CardKind string

var NilCard CardKind = "nil"

var WanCard CardKind = "wan"

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

var TiaoCard CardKind = "tiao"

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

var TongCard CardKind = "tong"

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

var WindCard CardKind = "feng"

//风 31 - 37
const (
	EAST  = 31
	SOUTH = 33
	WEST  = 35
	NORTH = 37
)

var OtherCard CardKind = "other"

//中发白 41 - 45
const (
	Zh = 41
	Fa = 43
	Ba = 45
)
