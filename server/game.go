package server

type RoomCtrl struct{}
type GameCtrl struct{}

func (game *GameCtrl) start() {
	//丢骰

	//洗牌

	//发牌

}

func (game *GameCtrl) dispatch() {
	//发牌

}

func (game *GameCtrl) load() {
	//加载同步
}

func (game *GameCtrl) ack() {
	//回执确认
}

func (game *GameCtrl) assert() {
	//判定
}
