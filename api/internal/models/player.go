package models

import (
	"github.com/gorilla/websocket"
	"github.com/style77/stockfish-or-not/internal/engine"
)

type Player struct {
	Conn   *websocket.Conn
	Room   *Room
	IsAI   bool
	Rank   *int    // ai only
	Engine *string // ai only
	AI     *engine.AIManager
}
