package utils

import "fmt"

func FormatTime(time float64) string {
	minutes := int(time) / 60
	seconds := int(time) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
