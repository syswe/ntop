package main

import (
	"fmt"
	"strings"
	"time"
)

func renderBar(percentage float64, width int) string {
	if percentage < 0 {
		percentage = 0
	} else if percentage > 100 {
		percentage = 100
	}

	filled := int((percentage / 100) * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return bar
}

func showLoadingBar(done chan bool) {
	for {
		select {
		case <-done:
			return
		default:
			fmt.Print(".")
			time.Sleep(500 * time.Millisecond) // Adjust the speed as needed
		}
	}
}

func calculatePercentage(current, limit float64) float64 {
	if limit == 0 {
		return 0 // Avoid division by zero; handle as needed
	}
	return (current / limit) * 100
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
