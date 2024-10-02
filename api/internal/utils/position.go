package utils

import (
	"log"

	"github.com/notnil/chess"
)

func GetPosition(moves []string) string {
	position := ""
	for _, move := range moves {
		position += move + " "
	}

	return position
}

type GameResult struct {
	Outcome       chess.Outcome
	OutcomeReason string
}

func CheckEndGameStates(moves []string) (*GameResult, bool) {
	board := chess.NewGame()

	for _, move := range moves {
		board.MoveStr(move)
	}

	gameEnded := board.Outcome() != chess.NoOutcome

	result := GameResult{
		Outcome:       board.Outcome(),
		OutcomeReason: board.Method().String(),
	}

	log.Println("Game ended with outcome: ", result.Outcome, " Reason: ", result.OutcomeReason)
	log.Println(gameEnded)

	log.Println(board.String())

	return &result, gameEnded
}
