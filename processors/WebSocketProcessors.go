package processors

import (
	utilities "PlanningPoker/Utilities"
	"PlanningPoker/models"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Upgrader configuration for converting HTTP connections to WebSocket connections
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var votesVisible = false // Add this global state

// ConnectionHandler handles the connection from the clients
func ConnectionHandler(w http.ResponseWriter, r *http.Request, broadcast chan models.Message, players *[]models.Player, mutex *sync.Mutex) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	// Create new player
	newPlayer := models.Player{
		Connection: conn,
		Name:       "",
		Points:     "",
		IsObserver: false,
	}

	// Lock for thread safety when modifying players slice
	mutex.Lock()
	*players = append(*players, newPlayer)
	mutex.Unlock()

	// Send immediate welcome message
	welcomeMsg := models.Message{
		Username: "Server",
		Message:  "Welcome to Planning Poker!",
	}
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Println("Error sending welcome message:", err)
	}

	// Don't broadcast a new player list here, just notify about the join
	broadcast <- models.Message{
		Username:   "Server",
		Message:    fmt.Sprintf("Player joined. Total players: %d", len(*players)),
		Type:       "JOIN_SESSION",
		Connection: conn,
	}

	// Message handling loop
	for {
		var msg models.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			mutex.Lock()
			*players = utilities.RemovePlayer(*players, conn)
			playerCount := len(*players)
			mutex.Unlock()

			broadcast <- models.Message{
				Username: "Server",
				Message:  fmt.Sprintf("Player left. Total players: %d", playerCount),
				Type:     "LEAVE_SESSION",
			}
			break
		}

		// Set player name when they first join
		if msg.Type == "JOIN_SESSION" {
			msg.Connection = conn
			mutex.Lock()
			for i := range *players {
				if (*players)[i].Connection == conn {
					(*players)[i].Name = msg.Username
					(*players)[i].IsObserver = msg.Message == "observer"
					break
				}
			}
			mutex.Unlock()
		}

		// Handle vote submission
		if msg.Type == "GUESS" {
			mutex.Lock()
			player := utilities.GetPlayer(*players, conn)
			player.Points = msg.Message
			utilities.SetPlayer(*players, player)
			mutex.Unlock()
		}

		log.Printf("Received Message from %s: %s (Type: %s)\n", msg.Username, msg.Message, msg.Type)
		broadcast <- msg
	}
}

// MessageHandler handles the messages from the clients
func MessageHandler(game *models.Game, players *[]models.Player, broadcast chan models.Message, mutex *sync.Mutex) {
	for {
		msg := <-broadcast

		mutex.Lock()
		if msg.Type == "JOIN_SESSION" {
			for i := range *players {
				if (*players)[i].Name == msg.Username || (*players)[i].Name == "" {
					(*players)[i].Name = msg.Username
					(*players)[i].IsObserver = msg.Message == "observer"
					break
				}
			}
		}

		// Create player list with current visibility state
		playerList := make([]map[string]interface{}, 0)
		for _, p := range *players {
			if p.Name == "" {
				continue
			}
			playerData := map[string]interface{}{
				"name":       p.Name,
				"hasVoted":   p.Points != "",
				"isObserver": p.IsObserver,
			}
			// Only include votes if they should be visible
			if votesVisible && p.Points != "" {
				playerData["vote"] = p.Points
			}
			playerList = append(playerList, playerData)
		}
		mutex.Unlock()

		playerListMsg := models.Message{
			Username: "Server",
			Type:     "PLAYER_LIST",
			Players:  playerList,
		}

		// Handle CLEAR_VOTES
		if msg.Type == "CLEAR_VOTES" {
			votesVisible = false
			mutex.Lock()
			for i := range *players {
				(*players)[i].Points = ""
			}

			// Immediately create fresh player list after clearing
			playerList = make([]map[string]interface{}, 0)
			for _, p := range *players {
				if p.Name != "" {
					playerData := map[string]interface{}{
						"name":       p.Name,
						"hasVoted":   false, // Reset hasVoted status
						"isObserver": p.IsObserver,
					}
					playerList = append(playerList, playerData)
				}
			}
			mutex.Unlock()
			playerListMsg.Players = playerList // Update the message with new player list
		}

		// Handle GUESS messages
		if msg.Type == "GUESS" {
			// Check if all players have voted
			//allVoted := true
			mutex.Lock()
			for _, p := range *players {
				if p.Name != "" && !p.IsObserver && p.Points == "" {
					//allVoted = false
					break
				}
			}

			// Create fresh player list
			playerList = make([]map[string]interface{}, 0)
			for _, p := range *players {
				if p.Name != "" {
					playerData := map[string]interface{}{
						"name":       p.Name,
						"hasVoted":   p.Points != "",
						"isObserver": p.IsObserver,
					}
					if votesVisible {
						playerData["vote"] = p.Points
					}
					playerList = append(playerList, playerData)
				}
			}
			mutex.Unlock()
			playerListMsg.Players = playerList
		}

		// Handle SHOW_VOTES
		if msg.Type == "SHOW_VOTES" {
			votesVisible = true
			mutex.Lock()
			playerList = make([]map[string]interface{}, 0)
			for _, p := range *players {
				if p.Name != "" {
					playerData := map[string]interface{}{
						"name":       p.Name,
						"hasVoted":   p.Points != "",
						"isObserver": p.IsObserver,
					}
					if votesVisible {
						playerData["vote"] = p.Points
					}
					playerList = append(playerList, playerData)
				}
			}
			mutex.Unlock()
			playerListMsg.Players = playerList
		}

		// Send updates to all players
		for _, player := range *players {
			if err := player.Connection.WriteJSON(playerListMsg); err != nil {
				log.Println("Error sending message to player:", err)
				player.Connection.Close()
				mutex.Lock()
				*players = utilities.RemovePlayer(*players, player.Connection)
				playerListMsg.Message = fmt.Sprintf("Player left. Total players: %d", len(*players))
				mutex.Unlock()
				continue
			}
		}
	}
}
