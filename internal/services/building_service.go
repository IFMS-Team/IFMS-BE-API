package services

import (
	"IFMS-BE-API/internal/model/request"
	"IFMS-BE-API/internal/model/response"
	"IFMS-BE-API/internal/repository"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vippergod12/IFMS-BE/sql/generated"
	"go.uber.org/zap"
)

type BuildingService struct {
	repo   *repository.BuildingRepository
	logger *zap.Logger
}

func NewBuildingService(repo *repository.BuildingRepository, logger *zap.Logger) *BuildingService {
	return &BuildingService{
		repo:   repo,
		logger: logger,
	}
}

func (s *BuildingService) Create(ctx context.Context, req request.CreateBuildingRequest, userID pgtype.UUID) (response.BuildingResponse, error) {
	status := int32(1)
	if req.BuildingStatus != nil {
		status = *req.BuildingStatus
	}

	maxFloor := int32(0)
	if req.MaximumFloor != nil {
		maxFloor = *req.MaximumFloor
	}

	params := db.CreateBuildingParams{
		BuildingName:        req.BuildingName,
		BuildingAddress:     req.BuildingAddress,
		BuildingDescription: req.BuildingDescription,
		BuildingImage:       req.BuildingImage,
		BuildingStatus:      status,
		MaximumFloor:        maxFloor,
		CreatedBy:           userID,
		UpdatedBy:           userID,
	}

	building, err := s.repo.CreateBuilding(ctx, params)
	if err != nil {
		s.logger.Error("Failed to create building", zap.Error(err))
		return response.BuildingResponse{}, err
	}

	return s.mapToResponse(building), nil
}

func (s *BuildingService) GetByID(ctx context.Context, buildingID pgtype.UUID) (response.BuildingResponse, error) {
	building, err := s.repo.GetBuildingByID(ctx, buildingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.BuildingResponse{}, errors.New("building not found")
		}
		return response.BuildingResponse{}, err
	}
	return s.mapToResponse(building), nil
}

func (s *BuildingService) List(ctx context.Context, limit, offset int32) ([]response.BuildingResponse, int64, error) {
	buildings, err := s.repo.ListBuildings(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountBuildings(ctx)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]response.BuildingResponse, 0, len(buildings))
	for _, b := range buildings {
		resp = append(resp, s.mapToResponse(b))
	}

	return resp, total, nil
}

func (s *BuildingService) mapToResponse(b db.Building) response.BuildingResponse {
	return response.ToBuildingResponse(response.BuildingRow{
		BuildingID:          b.BuildingID,
		BuildingName:        b.BuildingName,
		BuildingAddress:     b.BuildingAddress,
		BuildingDescription: b.BuildingDescription,
		BuildingImage:       b.BuildingImage,
		BuildingStatus:      b.BuildingStatus,
		MaximumFloor:        b.MaximumFloor,
		CreatedAt:           b.CreatedAt.Time,
		UpdatedAt:           b.UpdatedAt.Time,
		CreatedBy:           b.CreatedBy,
		UpdatedBy:           b.UpdatedBy,
	})
}
