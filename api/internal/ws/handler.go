package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/style77/stockfish-or-not/internal"
	"github.com/style77/stockfish-or-not/internal/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func HandleConnections(w http.ResponseWriter, r *http.Request, app *internal.App) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to websocket:", err)
		return
	}

	player := &models.Player{Conn: conn, IsAI: false}

	go app.FindOpponent(player)

	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading JSON:", err)
			break
		}

		log.Println("Received message:", msg)
		if move, ok := msg["move"].(string); ok {
			app.ProcessMove(player, move, msg["isFirstMove"].(bool))
		}
	}

	player.Conn.Close()
}
