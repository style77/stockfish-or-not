package utils

func GetPosition(moves []string) string {
	position := ""
	for _, move := range moves {
		position += move + " "
	}

	return position
}
