package engine

import (
	"log"
	"sync"

	"github.com/freeeve/uci"
)

type AIManager struct {
	engine *uci.Engine
	mux    sync.Mutex
}

func NewAIManager(skillLevel int) *AIManager {
	engine, err := uci.NewEngine("../stockfish")
	if err != nil {
		log.Fatal("Error creating engine:", err)
	}

	engine.SendOption("Skill Level", skillLevel)
	return &AIManager{engine: engine}
}

func (m *AIManager) ProcessMove(position string, depth int) (string, error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	err := m.engine.SetMoves(position)
	if err != nil {
		log.Fatal("Error setting moves:", err)
		return "", err
	}

	results, err := m.engine.Go(depth, "", 0)
	if err != nil {
		log.Fatal("Error getting best move:", err)
		return "", err
	}

	bestMove := results.BestMove

	return bestMove, nil
}

func (m *AIManager) Close() {
	m.engine.Close()
}
