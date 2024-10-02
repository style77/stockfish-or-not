package main

import (
	"log"
	"net/http"

	"github.com/style77/stockfish-or-not/internal"
	"github.com/style77/stockfish-or-not/internal/ws"
)

func main() {
	app := internal.CreateApp()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.HandleConnections(w, r, app)
	})

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
