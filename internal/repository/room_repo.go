package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/vippergod12/IFMS-BE/sql/generated"
)

type RoomRepository struct {
	db      *pgxpool.Pool
	queries *db.Queries
}

func NewRoomRepository(pool *pgxpool.Pool, queries *db.Queries) *RoomRepository {
	return &RoomRepository{
		db:      pool,
		queries: queries,
	}
}

func (r *RoomRepository) CreateRoom(ctx context.Context, params db.CreateRoomParams) (db.Room, error) {
	return r.queries.CreateRoom(ctx, params)
}

func (r *RoomRepository) GetRoomByID(ctx context.Context, roomID pgtype.UUID) (db.Room, error) {
	return r.queries.GetRoomByID(ctx, roomID)
}

func (r *RoomRepository) ListRoomsByFloorID(ctx context.Context, floorID pgtype.UUID) ([]db.Room, error) {
	return r.queries.ListRoomsByFloorID(ctx, floorID)
}

func (r *RoomRepository) CountRoomsByFloorID(ctx context.Context, floorID pgtype.UUID) (int64, error) {
	return r.queries.CountRoomsByFloorID(ctx, floorID)
}

func (r *RoomRepository) UpdateRoom(ctx context.Context, params db.UpdateRoomParams) (db.Room, error) {
	return r.queries.UpdateRoom(ctx, params)
}
