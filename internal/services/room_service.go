package services

import (
	"context"
	"errors"

	"IFMS-BE-API/internal/model/request"
	"IFMS-BE-API/internal/model/response"
	"IFMS-BE-API/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vippergod12/IFMS-BE/sql/generated"
	"go.uber.org/zap"
)

var ErrFloorNotFound = errors.New("floor not found")
var ErrFloorMaxCapacity = errors.New("the floor has reached the maximum number of rooms allowed")

type RoomService struct {
	roomRepo  *repository.RoomRepository
	floorRepo *repository.FloorRepository
	logger    *zap.Logger
}

func NewRoomService(roomRepo *repository.RoomRepository, floorRepo *repository.FloorRepository, logger *zap.Logger) *RoomService {
	return &RoomService{
		roomRepo:  roomRepo,
		floorRepo: floorRepo,
		logger:    logger,
	}
}

func (s *RoomService) Create(ctx context.Context, req request.CreateRoomRequest, userID pgtype.UUID) (response.RoomResponse, error) {
	var floorID pgtype.UUID
	if err := floorID.Scan(req.FloorID); err != nil {
		return response.RoomResponse{}, errors.New("invalid floor id format")
	}

	floor, err := s.floorRepo.GetFloorByID(ctx, floorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.RoomResponse{}, ErrFloorNotFound
		}
		return response.RoomResponse{}, err
	}

	currentRoomCount, err := s.roomRepo.CountRoomsByFloorID(ctx, floorID)
	if err != nil {
		s.logger.Error("Failed to count rooms", zap.Error(err))
		return response.RoomResponse{}, err
	}

	if floor.MaximumRoom > 0 && int32(currentRoomCount+1) > floor.MaximumRoom {
		return response.RoomResponse{}, ErrFloorMaxCapacity
	}

	status := "available"
	if req.Status != "" {
		status = req.Status
	}

	params := db.CreateRoomParams{
		FloorID:         floorID,
		RoomName:        req.Name,
		RoomDescription: req.Description,
		RoomImage:       req.Image,
		RoomStatus:      status,
		CreatedBy:       userID,
		UpdatedBy:       userID,
	}

	room, err := s.roomRepo.CreateRoom(ctx, params)
	if err != nil {
		s.logger.Error("Failed to create room", zap.Error(err))
		return response.RoomResponse{}, err
	}

	return response.ToRoomResponse(room), nil
}

func (s *RoomService) Update(ctx context.Context, roomID pgtype.UUID, req request.UpdateRoomRequest, userID pgtype.UUID) (response.RoomResponse, error) {
	_, err := s.roomRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.RoomResponse{}, errors.New("room not found")
		}
		return response.RoomResponse{}, err
	}

	status := "available"
	if req.Status != "" {
		status = req.Status
	}

	params := db.UpdateRoomParams{
		RoomID:          roomID,
		RoomName:        req.Name,
		RoomDescription: req.Description,
		RoomImage:       req.Image,
		RoomStatus:      status,
		UpdatedBy:       userID,
	}

	room, err := s.roomRepo.UpdateRoom(ctx, params)
	if err != nil {
		s.logger.Error("Failed to update room", zap.Error(err))
		return response.RoomResponse{}, err
	}

	return response.ToRoomResponse(room), nil
}

func (s *RoomService) GetByID(ctx context.Context, roomID pgtype.UUID) (response.RoomResponse, error) {
	room, err := s.roomRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.RoomResponse{}, errors.New("room not found")
		}
		return response.RoomResponse{}, err
	}
	return response.ToRoomResponse(room), nil
}

func (s *RoomService) ListByFloorID(ctx context.Context, floorID pgtype.UUID) ([]response.RoomResponse, error) {
	rooms, err := s.roomRepo.ListRoomsByFloorID(ctx, floorID)
	if err != nil {
		return nil, err
	}
	return response.ToRoomListResponse(rooms), nil
}
