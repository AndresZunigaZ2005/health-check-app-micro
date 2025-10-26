package utils

import (
	"log"
)

func InitLogger() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func LogInfo(msg string) {
	log.Println("[INFO]", msg)
}

func LogError(msg string) {
	log.Println("[ERROR]", msg)
}
