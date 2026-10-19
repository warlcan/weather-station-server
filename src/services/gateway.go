package services

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"weather-server/src/model"
	"weather-server/src/repository"
)

var ErrInvalidToken = errors.New("invalid user token")
var ErrInvalidStationId = errors.New("invalid station hardware id")

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

func (s *GatewayService) UpdateStationsList(req *model.UpdateStationsListRequest) (*model.UpdateStationsListResponse, error) {
	gatewayUid, err := repository.PostgresDB.GetGatewayByHardwareId(req.GatewayHardwareId)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(req.MetheostationHardwareIds))
	for _, h := range req.MetheostationHardwareIds {
		v, err := strconv.ParseUint(h, 16, 32)
		if err != nil {
			log.Printf("stations update: bad id %q: %v", h, err)
			return nil, ErrInvalidStationId
		}
		ids = append(ids, int64(uint32(v)))
	}

	added, err := repository.PostgresDB.UpsertStationsForGateway(gatewayUid, ids)
	if err != nil {
		return nil, err
	}

	return &model.UpdateStationsListResponse{
		Added: added,
		Total: len(ids),
	}, nil
}
