package models

import (
	"github.com/gorilla/websocket"
	"github.com/style77/stockfish-or-not/internal/engine"
	"github.com/style77/stockfish-or-not/internal/timer"
)

type Player struct {
	Conn   *websocket.Conn
	Room   *Room
	IsAI   bool
	Rank   *int    // ai only
	Engine *string // ai only
	AI     *engine.AIManager

	Timer *timer.Timer
	Color *string
}
