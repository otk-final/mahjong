package api

import "encoding/json"

type WebPacket struct {
	Type   int
	Packet []byte
}

func UnPacket[T any](wp WebPacket) T {
	var t T
	_ = json.Unmarshal(wp.Packet, &t)
	return t
}
