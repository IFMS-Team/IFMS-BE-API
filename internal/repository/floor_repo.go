package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/vippergod12/IFMS-BE/sql/generated"
)

type FloorRepository struct {
	db      *pgxpool.Pool
	queries *db.Queries
}

func NewFloorRepository(pool *pgxpool.Pool, queries *db.Queries) *FloorRepository {
	return &FloorRepository{
		db:      pool,
		queries: queries,
	}
}

func (r *FloorRepository) CreateFloor(ctx context.Context, params db.CreateFloorParams) (db.Floor, error) {
	return r.queries.CreateFloor(ctx, params)
}

func (r *FloorRepository) CountFloorsByBuildingID(ctx context.Context, buildingID pgtype.UUID) (int64, error) {
	return r.queries.CountFloorsByBuildingID(ctx, buildingID)
}

func (r *FloorRepository) GetFloorByID(ctx context.Context, floorID pgtype.UUID) (db.Floor, error) {
	return r.queries.GetFloorByID(ctx, floorID)
}

func (r *FloorRepository) UpdateFloor(ctx context.Context, params db.UpdateFloorParams) (db.Floor, error) {
	return r.queries.UpdateFloor(ctx, params)
}

func (r *FloorRepository) CreateAuditLog(ctx context.Context, params db.CreateAuditLogParams) error {
	_, err := r.queries.CreateAuditLog(ctx, params)
	return err
}
