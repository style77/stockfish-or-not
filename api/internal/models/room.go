package models

import (
	"sync"
)

type Room struct {
	ID      string
	Player1 *Player
	Player2 *Player
	IsAI    bool
	Moves   []string
	Mux     sync.Mutex
	Turn    *Player

	GameEnded bool
}
