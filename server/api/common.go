package api

type IdentityHeader struct {
	UserId string
	Token  string
}

type NoResp struct {
}

var Empty = &NoResp{}
