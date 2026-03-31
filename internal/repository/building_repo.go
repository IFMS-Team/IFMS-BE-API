package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/vippergod12/IFMS-BE/sql/generated"
)

type BuildingRepository struct {
	db      *pgxpool.Pool
	queries *db.Queries
}

func NewBuildingRepository(pool *pgxpool.Pool) *BuildingRepository {
	queries := db.New(pool)
	return &BuildingRepository{
		db:      pool,
		queries: queries,
	}
}

func (r *BuildingRepository) CreateBuilding(ctx context.Context, params db.CreateBuildingParams) (db.Building, error) {
	return r.queries.CreateBuilding(ctx, params)
}

func (r *BuildingRepository) GetBuildingByID(ctx context.Context, buildingID pgtype.UUID) (db.Building, error) {
	return r.queries.GetBuildingByID(ctx, buildingID)
}

func (r *BuildingRepository) ListBuildings(ctx context.Context, limit, offset int32) ([]db.Building, error) {
	return r.queries.ListBuildings(ctx, db.ListBuildingsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *BuildingRepository) CountBuildings(ctx context.Context) (int64, error) {
	return r.queries.CountBuildings(ctx)
}
