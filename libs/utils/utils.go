package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"
)

func GenerateRandomString(length int) string {
	buffer := make([]byte, length/2)
	_, err := rand.Read(buffer)
	if err != nil {
		fmt.Println("Cannot generate a file name.")
		return "untitled.jpg"
	}
	return hex.EncodeToString(buffer)
}

func CreateFolder(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func FormatExecutionTime(duration time.Duration) string {
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60

	return fmt.Sprintf("%d m %d s", minutes, seconds)
}
