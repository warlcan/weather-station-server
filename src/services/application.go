package services

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math/big"
	"net/smtp"
	"os"
	"time"
	"weather-server/src/model"
	"weather-server/src/repository"

	"github.com/google/uuid"
)

type EmailData struct {
	SenderEmail    string
	RecipientEmail string
	Theme          string
	Code           string
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func VerifyToken(receivedToken string) bool {
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

	tmpl, err := template.ParseFiles(
		"templates/mail/headers.tmpl",
		"templates/mail/sendCodeTemplate.tmpl",
	)

	data := EmailData{
		SenderEmail:    os.Getenv("EMAIL_ADDRESS"),
		RecipientEmail: email,
		Theme:          "Код подтверждения",
		Code:           randomCode,
	}

	var body bytes.Buffer
	err = tmpl.ExecuteTemplate(&body, "headers.tmpl", data)
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
		log.Printf("Error send mail: %v", err)
		return false
	}

	err = repository.Redis.WriteStringTTL("email:"+email+":code", randomCode, 5)
	if err != nil {
		log.Printf("Error write cache: %v", err)
		return false
	}

	return true
}

func SignInOrRegister(email string, receivedCode string) (string, error) {
	dbCode, err := repository.Redis.GetString("email:" + email + ":code")
	if err != nil {
		log.Printf("Error getting code from Redis: %v", err)
		return "", err
	}

	if receivedCode != dbCode {
		return "", errors.New("invalid verification code")
	}

	uid, err := repository.PostgresDB.UpsertUser(email)
	if err != nil {
		log.Printf("Error upserting user: %v", err)
		return "", err
	}

	token := uuid.New().String()
	err = repository.Redis.WriteString("token:"+hashToken(token)+":userid", uid)
	if err != nil {
		log.Printf("Error writing token to Redis: %v", err)
		return "", err
	}

	_ = repository.Redis.DeleteString("email:" + email + ":code")

	return token, nil
}

func GetUserIdByToken(token string) (string, error) {
	return repository.Redis.GetString("token:" + hashToken(token) + ":userid")
}

func ListUserStations(userUid string) ([]model.AppWeatherStation, error) {
	rows, err := repository.PostgresDB.ListUserStations(userUid)
	if err != nil {
		return nil, err
	}

	out := make([]model.AppWeatherStation, 0, len(rows))
	for _, r := range rows {
		out = append(out, model.AppWeatherStation{
			Uid:        r.Uid,
			HardwareId: fmt.Sprintf("%08x", uint32(r.HardwareId)),
			Name:       r.Name,
			ErrorMask:  r.ErrorMask,
			GatewayUid: r.GatewayUid,
		})
	}
	return out, nil
}

func ListMeasurements(userUid string, stationUids []int64) (map[int64][]model.AppMeasurement, error) {
	stations, err := repository.PostgresDB.ListUserStations(userUid)
	if err != nil {
		return nil, err
	}

	allowed := make(map[int64]struct{}, len(stations))
	for _, s := range stations {
		allowed[s.Uid] = struct{}{}
	}

	filtered := make([]int64, 0, len(stationUids))
	for _, id := range stationUids {
		if _, ok := allowed[id]; ok {
			filtered = append(filtered, id)
		}
	}
	if len(filtered) == 0 {
		return map[int64][]model.AppMeasurement{}, nil
	}

	since := time.Now().Add(-24 * time.Hour)
	rows, err := repository.PostgresDB.ListHourlyMeasurements(filtered, since)
	if err != nil {
		return nil, err
	}

	out := make(map[int64][]model.AppMeasurement)
	for _, m := range rows {
		out[m.WeatherStationUid] = append(out[m.WeatherStationUid], model.AppMeasurement{
			Time:         m.Time,
			Temperature:  m.Temperature,
			Humidity:     m.Humidity,
			Pressure:     m.Pressure,
			VoltageLevel: m.VoltageLevel,
		})
	}
	return out, nil
}
