package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"weather-server/src/repository"
)

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func IsTokenValid(receivedToken string) bool {
	count, err := repository.Redis.ExistsString("token:" + hashToken(receivedToken) + ":userid")
	if err != nil {
		log.Fatalf("Error in search: %v", err)
	}

	if count > 0 {
		return true
	}

	return false
}
