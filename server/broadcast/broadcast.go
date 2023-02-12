package broadcast

import (
	"encoding/json"
	"log"
	"mahjong/robot"
	"mahjong/server/api"
	"mahjong/server/ws"
	"mahjong/service/engine"
)

type Handler struct {
	RoomId  string
	Pos     *engine.Position
	Players []*api.Player
}

//真实玩家
func (h *Handler) getPlayers() []*api.Player {
	players := make([]*api.Player, 0)
	for _, p := range h.Players {
		ok, _ := h.Pos.IsRobot(p.Idx)
		if ok {
			continue
		}
		players = append(players, p)
	}
	return players
}

//机器人
func (h *Handler) getRobots() []*api.Roboter {
	robots := make([]*api.Roboter, 0)
	for _, p := range h.Players {
		ok, r := h.Pos.IsRobot(p.Idx)
		if !ok {
			continue
		}
		robots = append(robots, r)
	}
	return robots
}

func (h *Handler) isRobot(who int) (bool, *api.Roboter) {
	return h.Pos.IsRobot(who)
}

func (h *Handler) Take(event *api.TakePayload) {
	packet := api.Packet(api.TakeEvent, "摸牌", event)
	robots := h.getRobots()
	for _, roboter := range robots {
		robot.Post(h.RoomId, roboter, packet)
	}
	Post(h.RoomId, h.getPlayers(), packet)
}

func (h *Handler) Put(event *api.PutPayload) {
	packet := api.Packet(api.PutEvent, "打牌", event)
	robots := h.getRobots()
	for _, roboter := range robots {
		robot.Post(h.RoomId, roboter, packet)
	}
	Post(h.RoomId, h.getPlayers(), packet)
}

func (h *Handler) Race(event *api.RacePayload) {
	packet := api.Packet(api.RaceEvent, api.RaceNames[event.RaceType], event)
	robots := h.getRobots()
	for _, roboter := range robots {
		robot.Post(h.RoomId, roboter, packet)
	}
	Post(h.RoomId, h.getPlayers(), packet)
}

func (h *Handler) Win(event *api.WinPayload) {
	packet := api.Packet(api.WinEvent, "胡牌", event)
	robots := h.getRobots()
	for _, roboter := range robots {
		robot.Post(h.RoomId, roboter, packet)
	}
	Post(h.RoomId, h.getPlayers(), api.Packet(api.WinEvent, "胡牌", event))
}

func (h *Handler) Ack(event *api.AckPayload) {
	log.Printf("通知：玩家[%d] pass\n", event.Who)
	Post(h.RoomId, h.getPlayers(), api.Packet(api.AckEvent, "确认", event))
}

func (h *Handler) Turn(event *api.TurnPayload, ok bool) {
	log.Printf("通知：当前回合 玩家[%d]\n", event.Who)

	packet := api.Packet(api.TurnEvent, "轮转", event)
	if ok, roboter := h.isRobot(event.Who); ok {
		robot.Post(h.RoomId, roboter, packet)
	}
	Post(h.RoomId, h.getPlayers(), packet)
}

func (h *Handler) Quit(ok bool) {

}

func Post[T any](roomId string, players []*api.Player, packet *api.WebPacket[T]) {
	//序列化 json
	msg, _ := json.Marshal(packet)
	//所有成员
	for _, member := range players {
		ws.PostMessage(roomId, member.UId, msg)
	}
}

func PostFunc[T any](roomId string, players []*api.Player, fn func(*api.Player) *api.WebPacket[T]) {
	//所有成员
	for _, member := range players {
		packet := fn(member)
		//序列化 json
		msg, _ := json.Marshal(packet)
		ws.PostMessage(roomId, member.UId, msg)
	}
}
