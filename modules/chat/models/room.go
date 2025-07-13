package models

import (
	"encoding/gob"
	"slices"

	commonGen "github.com/bozoteam/roshan/adapter/grpc/gen/common"
	userGen "github.com/bozoteam/roshan/adapter/grpc/gen/user"

	"github.com/bozoteam/roshan/helpers"
	ws_hub "github.com/bozoteam/roshan/modules/websocket/hub"
	"github.com/bozoteam/roshan/modules/websocket/ws_client"
)

func init() {
	gob.Register((*Room)(nil)) // Register pointer type
	gob.Register((*ws_hub.RoomI)(nil))
	gob.Register((*ws_hub.ClientI)(nil))
	gob.Register((*ws_client.Client)(nil))
}

type Team = string
type UUID = string

var _ ws_hub.RoomI = (*Room)(nil) // Correct - pointer implements interface

// Room implements the RoomI interface for chat rooms
type Room struct {
	ID        string
	Name      string
	CreatorID string

	Clients     map[UUID]ws_hub.ClientTeam   // Maps client ID to ClientI interface
	ClientTeams map[Team][]ws_hub.ClientTeam // Maps team to clients

	Teams []string

	Kind string

	someoneEntered bool
}

func NewRoom(name string, creatorId string, teams []string, kind string) *Room {
	return &Room{
		ID:          helpers.GenUUID(),
		Name:        name,
		CreatorID:   creatorId,
		Clients:     make(map[string]ws_hub.ClientTeam),
		ClientTeams: make(map[Team][]ws_hub.ClientTeam),
		Teams:       teams,
		Kind:        kind,

		someoneEntered: false,
	}
}

func (r *Room) GetClientsFromTeam(team string) []ws_hub.ClientTeam {
	return r.ClientTeams[team]
}

func (r *Room) GetTeamMapping() map[Team][]ws_hub.ClientTeam {
	return r.ClientTeams
}

func (r *Room) GetID() string {
	return r.ID
}

func (r *Room) RegisterClient(client ws_hub.ClientI, team string) {
	r.Clients[client.GetID()] = ws_hub.ClientTeam{
		ClientI: client,
		Team:    team,
	}
}

func (r *Room) UnregisterClient(clientId string) {
	team := r.Clients[clientId].Team
	delete(r.Clients, clientId)
	if teamClients, exists := r.ClientTeams[team]; exists {
		for i, client := range teamClients {
			if client.GetID() == clientId {
				r.ClientTeams[team] = slices.Delete(teamClients, i, i+1)
				break
			}
		}
		if len(r.ClientTeams[team]) == 0 {
			delete(r.ClientTeams, team)
		}
	}
}

func (r *Room) GetClients() map[string]ws_hub.ClientTeam {
	return r.Clients
}

func (r *Room) SetSomeoneEntered(entered bool) {
	r.someoneEntered = entered
}

func (r *Room) GetSomeoneEntered() bool {
	return r.someoneEntered
}

func (r *Room) Clone() ws_hub.RoomI {
	return helpers.Clone(r)
}

func (r *Room) UserIsInRoom(userId string) bool {
	_, exists := r.Clients[userId]
	return exists
}

func (r *Room) ToGRPCRoom() *commonGen.Room {
	teamUserMap := make(map[string]*commonGen.UserList)

	for team, clients := range r.ClientTeams {
		userList := &commonGen.UserList{
			Users: make([]*userGen.User, len(clients)),
		}
		for i, client := range clients {
			userList.Users[i] = &userGen.User{
				Id:    client.GetID(),
				Name:  client.GetUser().Name,
				Email: client.GetUser().Email,
			}
		}
		teamUserMap[team] = userList
	}

	kind := commonGen.RoomKind_ROOM_KIND_UNSPECIFIED

	switch r.Kind {
	case "chat":
		kind = commonGen.RoomKind_ROOM_KIND_CHAT
	case "game":
		kind = commonGen.RoomKind_ROOM_KIND_GAME
	default:
		kind = commonGen.RoomKind_ROOM_KIND_UNSPECIFIED

	}

	return &commonGen.Room{
		Id:           r.ID,
		CreatorId:    r.CreatorID,
		Name:         r.Name,
		AllowedTeams: r.Teams,
		TeamUserMap:  teamUserMap,
		Kind:         kind,
	}
}
