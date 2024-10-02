package game

import (
	"github.com/style77/stockfish-or-not/internal/models"
	"github.com/style77/stockfish-or-not/internal/utils"
)

func HandleGameEnd(playerTurn *models.Player, room *models.Room, reason string, result *utils.GameResult) {
	room.Mux.Lock()
	defer room.Mux.Unlock()

	room.GameEnded = true

	var aiPlayer *models.Player
	if room.IsAI {
		if room.Player1.IsAI {
			aiPlayer = room.Player1
		} else {
			aiPlayer = room.Player2
		}

		if aiPlayer.AI != nil {
			aiPlayer.AI.Close()
		}
	}

	utils.NotifyBothPlayers(room, map[string]interface{}{
		"state":   99,
		"roomID":  room.ID,
		"message": "Game ended",
		"data": map[string]interface{}{
			"result": result.Outcome.String(),
			"reason": reason,
			"isAI":   room.IsAI,
			"AIMeta": map[string]interface{}{
				"rank":   aiPlayer.Rank,
				"engine": aiPlayer.Engine,
			},
		},
	})

	if room.Player1 != nil {
		if room.Player1.Timer != nil {
			room.Player1.Timer.StopTimer()
		}
		if room.Player1.Conn != nil {
			room.Player1.Conn.Close()
		}
	}

	if room.Player2 != nil {
		if room.Player2.Timer != nil {
			room.Player2.Timer.StopTimer()
		}
		if room.Player2.Conn != nil {
			room.Player2.Conn.Close()
		}
	}

	room.Player1 = nil
	room.Player2 = nil
	room.Turn = nil
}
