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
		log.Println("Player has no connection.")
	}

	return nil
}
