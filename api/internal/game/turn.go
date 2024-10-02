package game

import (
	"log"

	"github.com/style77/stockfish-or-not/internal/models"
	"github.com/style77/stockfish-or-not/internal/utils"
)

func ChangeTurn(room *models.Room) {
	var nextTurnPlayer *models.Player
	var lastTurnPlayer *models.Player

	if room.Turn == room.Player1 {
		nextTurnPlayer = room.Player2
		lastTurnPlayer = room.Player1
	} else {
		nextTurnPlayer = room.Player1
		lastTurnPlayer = room.Player2
	}

	room.Mux.Lock()
	room.Turn = nextTurnPlayer
	room.Mux.Unlock()

	err := utils.SafelyNotifyPlayer(nextTurnPlayer, map[string]interface{}{
		"message": "Your turn",
		"roomID":  room.ID,
		"state":   79,
	})

	if err != nil {
		log.Println("Error sending turn message to Player 2:", err)
	}

	// if both of timers are stopped then start the next turn player's timer
	if !lastTurnPlayer.Timer.IsStarted && !nextTurnPlayer.Timer.IsStarted {
		nextTurnPlayer.Timer.StartTimer()
		return
	} else if !nextTurnPlayer.Timer.IsStarted {
		nextTurnPlayer.Timer.StartTimer()
	}

	lastTurnPlayer.Timer.PauseTimer()
	nextTurnPlayer.Timer.ResumeTimer()
}
