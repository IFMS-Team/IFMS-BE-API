package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"IFMS-BE-API/internal/model/request"
	"IFMS-BE-API/internal/model/response"
	"IFMS-BE-API/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vippergod12/IFMS-BE/sql/generated"
	"go.uber.org/zap"
)

type FloorService struct {
	floorRepo    *repository.FloorRepository
	buildingRepo *repository.BuildingRepository
	logger       *zap.Logger
}

func NewFloorService(floorRepo *repository.FloorRepository, buildingRepo *repository.BuildingRepository, logger *zap.Logger) *FloorService {
	return &FloorService{
		floorRepo:    floorRepo,
		buildingRepo: buildingRepo,
		logger:       logger,
	}
}

var ErrBuildingMaxCapacity = errors.New("the building has reached the maximum number of floors allowed")
var ErrBuildingNotFound = errors.New("building not found")

func (s *FloorService) Create(ctx context.Context, req request.CreateFloorRequest, userID pgtype.UUID) (response.FloorResponse, error) {
	var buildingID pgtype.UUID
	if err := buildingID.Scan(req.BuildingID); err != nil {
		return response.FloorResponse{}, errors.New("invalid building id format")
	}

	// 2. Building existence
	bldg, err := s.buildingRepo.GetBuildingByID(ctx, buildingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.FloorResponse{}, ErrBuildingNotFound
		}
		return response.FloorResponse{}, err
	}

	// 3. Maximum floor constraint
	currentFloorCount, err := s.floorRepo.CountFloorsByBuildingID(ctx, buildingID)
	if err != nil {
		s.logger.Error("Failed to count floors", zap.Error(err))
		return response.FloorResponse{}, err
	}

	if int32(currentFloorCount+1) > bldg.MaximumFloor {
		return response.FloorResponse{}, ErrBuildingMaxCapacity
	}

	// 4. Data to persist
	params := db.CreateFloorParams{
		BuildingID:       buildingID,
		FloorName:        req.Name,
		FloorDescription: "",
		FloorImage:       "",
		FloorStatus:      int32(1),
		MaximumRoom:      int32(0),
		CreatedBy:        userID,
		UpdatedBy:        userID,
	}

	createdFloor, err := s.floorRepo.CreateFloor(ctx, params)
	if err != nil {
		s.logger.Error("Failed to create floor", zap.Error(err))
		return response.FloorResponse{}, err
	}

	// 5. Audit log
	newDataMap := map[string]interface{}{
		"buildingId": req.BuildingID,
		"name":       req.Name,
		"createdBy":  userID.Bytes,
		"updatedBy":  userID.Bytes,
		"createdAt":  time.Now(),
		"updatedAt":  time.Now(),
	}
	newDataBytes, _ := json.Marshal(newDataMap)

	auditParams := db.CreateAuditLogParams{
		UserID:     userID,
		Username:   "system", // Ideally extract username from context 
		Path:       "POST /api/v1/floors",
		Action:     "FLOOR_CREATED",
		TableName:  "floors",
		RecordID:   response.ToFloorResponseFromDB(createdFloor).ID,
		OldData:    nil,
		NewData:    newDataBytes,
		Request:    nil,
		Response:   nil,
		StatusCode: 201,
		LatencyMs:  0,
		IpAddress:  "",
		UserAgent:  "",
	}
	_ = s.floorRepo.CreateAuditLog(ctx, auditParams)

	return response.ToFloorResponseFromDB(createdFloor), nil
}
