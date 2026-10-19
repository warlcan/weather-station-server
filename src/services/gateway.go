package services

import (
	"errors"
	"fmt"
	"log"

	"weather-server/src/model"
	"weather-server/src/repository"
)

var ErrInvalidToken = errors.New("invalid user token")

type GatewayService struct{}

func NewGatewayService() *GatewayService {
	return &GatewayService{}
}

func (s *GatewayService) InitGateway(req *model.InitGatewayRequest) (*model.InitGatewayResponse, error) {
	key := "token:" + req.UserApiToken + ":userid"
	userUid, err := repository.Redis.GetString(key)
	if err != nil {
		return nil, fmt.Errorf("redis lookup: %w", err)
	}
	if userUid == "" {
		log.Printf("gateway init: token not found in redis, key=%q", key)
		return nil, ErrInvalidToken
	}

	gatewayUid, err := repository.PostgresDB.UpsertGateway(req.GatewayHardwareId, userUid)
	if err != nil {
		return nil, err
	}

	return &model.InitGatewayResponse{GatewayUid: gatewayUid}, nil
}
