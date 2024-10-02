package engine

import (
	"math/rand"
)

var StockfishSkillElo = map[int]int{
	0:  800,
	1:  850,
	2:  900,
	3:  1000,
	4:  1100,
	5:  1200,
	6:  1300,
	7:  1400,
	8:  1500,
	9:  1600,
	10: 1700,
	11: 1800,
	12: 1900,
	13: 2000,
	14: 2100,
	15: 2200,
	16: 2300,
	17: 2400,
	18: 2500,
	19: 2600,
	20: 3000, // 3000+ Elo for highest level
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func getSkillLevel(elo int) int {
	closestLevel := 0
	closestDifference := 10000

	for level, rating := range StockfishSkillElo {
		difference := abs(rating - elo)
		if difference < closestDifference {
			closestDifference = difference
			closestLevel = level
		}
	}
	return closestLevel
}

func DeterminateAI() (*AIManager, int) {
	selectedRank := rand.Intn(2000) + 100 // Random Elo between 100 and 2100
	skillLevel := getSkillLevel(selectedRank)

	manager := NewAIManager(skillLevel)

	return manager, StockfishSkillElo[skillLevel]
}
