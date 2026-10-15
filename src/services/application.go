package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"weather-server/src/repository"
)

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func IsTokenValid(receivedToken string) bool {
	count, err := repository.Redis.ExistsString("token:" + HashToken(receivedToken))
	fmt.Println(receivedToken + " " + HashToken(receivedToken))
	fmt.Println(count)

	if err != nil {
		log.Fatalf("Error in search: %v", err)
	}

	if count > 0 {
		return true
	}

	return false
}
