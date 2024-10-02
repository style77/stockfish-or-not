package utils

import (
	"log"

	"github.com/style77/stockfish-or-not/internal/models"
)

func ChangeTurn(room *models.Room) {
	room.Mux.Lock()
	defer room.Mux.Unlock()

	if room.Turn == room.Player1 {
		room.Turn = room.Player2
		if room.Player2 != nil {
			err := SafelyNotifyPlayer(room.Player2, map[string]interface{}{
				"message": "Your turn",
				"roomID":  room.ID,
				"state":   79,
			})

			if err != nil {
				log.Println("Error sending turn message to Player 2:", err)
			}
		}
	} else {
		room.Turn = room.Player1

		err := SafelyNotifyPlayer(room.Player1, map[string]interface{}{
			"message": "Your turn",
			"roomID":  room.ID,
			"state":   79,
		})

		if err != nil {
			log.Println("Error sending turn message to Player 1:", err)
		}

	}
}
