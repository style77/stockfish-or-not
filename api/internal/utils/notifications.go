package utils

import (
	"log"

	"github.com/style77/stockfish-or-not/internal/models"
)

func SafelyNotifyPlayer(player *models.Player, data map[string]interface{}) error {
	if player.Conn != nil {
		err := player.Conn.WriteJSON(data)
		if err != nil {
			log.Println("Error notifying player:", err)
		}
		return err
	} else {
		// log.Println("Player has no connection.")
	}

	return nil
}

func NotifyBothPlayers(room *models.Room, message map[string]interface{}) error {
	if err := SafelyNotifyPlayer(room.Player1, message); err != nil {
		// log.Println("Error notifying player 1:", err)
	}

	if err := SafelyNotifyPlayer(room.Player2, message); err != nil {
		// log.Println("Error notifying player 2:", err)
	}

	return nil
}
