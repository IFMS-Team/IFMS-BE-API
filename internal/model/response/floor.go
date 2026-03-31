package response

import (
	"time"

	"github.com/google/uuid"

	db "github.com/vippergod12/IFMS-BE/sql/generated"
)

type FloorResponse struct {
	ID         string    `json:"id"`
	BuildingID string    `json:"buildingId"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	CreatedBy  string    `json:"createdBy"`
	UpdatedBy  string    `json:"updatedBy"`
}

func ToFloorResponseFromDB(b db.Floor) FloorResponse {
	return FloorResponse{
		ID:         uuid.UUID(b.FloorID.Bytes).String(),
		BuildingID: uuid.UUID(b.BuildingID.Bytes).String(),
		Name:       b.FloorName,
		CreatedAt:  b.CreatedAt.Time,
		UpdatedAt:  b.UpdatedAt.Time,
		CreatedBy:  uuid.UUID(b.CreatedBy.Bytes).String(),
		UpdatedBy:  uuid.UUID(b.UpdatedBy.Bytes).String(),
	}
}

func ToFloorListResponse(floors []db.Floor) []FloorResponse {
	var list []FloorResponse
	for _, f := range floors {
		list = append(list, ToFloorResponseFromDB(f))
	}
	return list
}
