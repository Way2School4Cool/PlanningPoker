package models

import (
	"github.com/gorilla/websocket"
)

// Message format
type Message struct {
	Username   string                   `json:"username"`
	Message    string                   `json:"message"`
	Type       string                   `json:"type"`
	Players    []map[string]interface{} `json:"players,omitempty"`
	Connection *websocket.Conn          `json:"-"`
}
