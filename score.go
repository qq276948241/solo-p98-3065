package main

import (
	"fmt"
	"os"
	"strconv"
)

const highScoreFile = "highscore.txt"

func LoadHighScore() int {
	data, err := os.ReadFile(highScoreFile)
	if err != nil {
		return 0
	}
	val, err := strconv.Atoi(string(data))
	if err != nil {
		return 0
	}
	return val
}

func SaveHighScore(score int) error {
	return os.WriteFile(highScoreFile, []byte(fmt.Sprintf("%d", score)), 0644)
}

func UpdateHighScoreIfNeeded(currentScore int, highScore int) int {
	if currentScore > highScore {
		SaveHighScore(currentScore)
		return currentScore
	}
	return highScore
}
