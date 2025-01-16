package main

import (
	models "PlanningPoker/models"
	processors "PlanningPoker/processors"
	"fmt"
	"log"
	"net/http"
	"sync"
)

var (
	game = models.Game{
		Status:         models.WAITING_FOR_PLAYERS,
		Players:        make([]models.Player, 0),
		QuestionNumber: 0,
	}
	broadcast = make(chan models.Message) // Broadcast channel for sending messages
	mutex     = sync.Mutex{}
)

func main() {
	fmt.Println("Starting server...")

	// Serve static files from the html directory
	fs := http.FileServer(http.Dir("html"))
	http.Handle("/", fs)
	http.Handle("/scripts.js", fs)
	http.Handle("/styles.css", fs)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request for /ws from %s\n", r.Method, r.RemoteAddr)
		processors.ConnectionHandler(w, r, broadcast, &game.Players, &mutex)
	})

	go processors.MessageHandler(&game, &game.Players, broadcast, &mutex)

	// Start the server on port 6969
	fmt.Println("Server is running on :6969")
	if err := http.ListenAndServe("0.0.0.0:6969", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
