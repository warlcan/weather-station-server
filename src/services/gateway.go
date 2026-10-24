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

func (s *GatewayService) StoreMeasurement(req *model.MeasureRequest) (*model.MeasureResponse, error) {
	gatewayUid, err := repository.PostgresDB.GetGatewayByHardwareId(req.GatewayHardwareId)
	if err != nil {
		return nil, err
	}

	v, err := strconv.ParseUint(req.WeatherStationHardwareId, 16, 32)
	if err != nil {
		log.Printf("measure: bad station id %q: %v", req.WeatherStationHardwareId, err)
		return nil, ErrInvalidStationId
	}
	stationHwId := int64(uint32(v))

	stationUid, err := repository.PostgresDB.GetStationUidForGateway(gatewayUid, stationHwId)
	if err != nil {
		return nil, err
	}

	m := repository.Measurement{
		Time:              req.Time,
		WeatherStationUid: stationUid,
		Temperature:       req.Temperature,
		Humidity:          req.Humidity,
		Pressure:          req.Pressure,
		VoltageLevel:      req.VoltageLevel,
	}

	if err := repository.PostgresDB.InsertMeasurement(m, req.ErrorMask); err != nil {
		return nil, err
	}

	return &model.MeasureResponse{Ok: true}, nil
}
