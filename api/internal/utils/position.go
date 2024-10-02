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
	board := chess.NewGame(chess.UseNotation(chess.UCINotation{}))

	for _, move := range moves {
		if err := board.MoveStr(move); err != nil {
			log.Printf("Error applying move %s: %v", move, err)
			continue
		}
	}

	gameEnded := board.Outcome() != chess.NoOutcome

	result := GameResult{
		Outcome:       board.Outcome(),
		OutcomeReason: board.Method().String(),
	}

	return &result, gameEnded
}
