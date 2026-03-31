package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type BuildingResponse struct {
	BuildingID          string    `json:"building_id"`
	BuildingName        string    `json:"building_name"`
	BuildingAddress     string    `json:"building_address"`
	BuildingDescription string    `json:"building_description"`
	BuildingImage       string    `json:"building_image"`
	BuildingStatus      int32     `json:"building_status"`
	MaximumFloor        int32     `json:"maximum_floor"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	CreatedBy           string    `json:"created_by"`
	UpdatedBy           string    `json:"updated_by"`
}

type BuildingRow struct {
	BuildingID          pgtype.UUID
	BuildingName        string
	BuildingAddress     string
	BuildingDescription string
	BuildingImage       string
	BuildingStatus      int32
	MaximumFloor        int32
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CreatedBy           pgtype.UUID
	UpdatedBy           pgtype.UUID
}

func ToBuildingResponse(b BuildingRow) BuildingResponse {
	return BuildingResponse{
		BuildingID:          uuid.UUID(b.BuildingID.Bytes).String(),
		BuildingName:        b.BuildingName,
		BuildingAddress:     b.BuildingAddress,
		BuildingDescription: b.BuildingDescription,
		BuildingImage:       b.BuildingImage,
		BuildingStatus:      b.BuildingStatus,
		MaximumFloor:        b.MaximumFloor,
		CreatedAt:           b.CreatedAt,
		UpdatedAt:           b.UpdatedAt,
		CreatedBy:           uuid.UUID(b.CreatedBy.Bytes).String(),
		UpdatedBy:           uuid.UUID(b.UpdatedBy.Bytes).String(),
	}
}

