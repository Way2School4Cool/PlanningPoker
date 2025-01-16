package utilities

import (
	"PlanningPoker/models"
	"errors"

	"github.com/gorilla/websocket"
)

func CheckForDuplicateUsername(clients map[*websocket.Conn]bool, msg models.Message) bool {
	for existingConn := range clients {
		var existingMsg models.Message
		if err := existingConn.ReadJSON(&existingMsg); err != nil {
			continue
		}
		if existingMsg.Username == msg.Username {
			return true
		}
	}
	return false
}

func RemovePlayer(players []models.Player, conn *websocket.Conn) []models.Player {
	for i, player := range players {
		if player.Connection == conn {
			return append(players[:i], players[i+1:]...)
		}
	}
	return players
}

func SetPlayerName(players []models.Player, name string, conn *websocket.Conn) error {
	for i, player := range players {
		if player.Connection == conn {
			player.Name = name
			players[i] = player
			return nil
		}
	}
	return errors.New("player not found")
}

func GetPlayer(players []models.Player, conn *websocket.Conn) models.Player {
	for _, player := range players {
		if player.Connection == conn {
			return player
		}
	}
	return models.Player{}
}

func SetPlayer(players []models.Player, player models.Player) {
	for i, p := range players {
		if p.Connection == player.Connection {
			players[i] = player
			return
		}
	}
}
