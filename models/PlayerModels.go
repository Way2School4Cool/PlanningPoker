package models

import "github.com/gorilla/websocket"

type Player struct {
	Connection *websocket.Conn
	Name       string
	Points     string
	IsObserver bool
}
