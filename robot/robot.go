package robot

import "mahjong/server/api"

type Client struct {
	Level  int
	Player *api.Player
}
