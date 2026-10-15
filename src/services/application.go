package services

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"net/smtp"
	"os"
	"weather-server/src/repository"
	"weather-server/src/templates"
)

type EmailData struct {
	SenderEmail    string
	RecipientEmail string
	Code           string
}

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

func generateRandomCode() (string, error) {
	max := big.NewInt(899999)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	code := n.Int64() + 100000

	return fmt.Sprintf("%d", code), nil
}

func SendCode(email string) bool {
	randomCode, err := generateRandomCode()
	if err != nil {
		log.Printf("Error in generate random number: %v", err)
		return false
	}

	tmpl, err := templates.GetTemplate("templates/mail/sendCodeTemplate.tmpl")

	data := EmailData{
		SenderEmail:    os.Getenv("EMAIL_ADDRESS"),
		RecipientEmail: email,
		Code:           randomCode,
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		return false
	}

	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := "smtp.yandex.ru"
	smtpPort := "587"

	smtpAuth := smtp.PlainAuth("", from, password, smtpHost)
	err = smtp.SendMail(smtpHost+":"+smtpPort, smtpAuth, from, []string{email}, body.Bytes())
	if err != nil {
		log.Printf("Error in send mail: %v", err)
		return false
	}

	err = repository.Redis.WriteStringTTL("email:"+email+":code", randomCode, 5)
	if err != nil {
		log.Printf("Error in write cache: %v", err)
		return false
	}

	return true
}
