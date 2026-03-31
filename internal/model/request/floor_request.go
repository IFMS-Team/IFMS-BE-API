package request

type CreateFloorRequest struct {
	BuildingID string `json:"buildingId" binding:"required"`
	Name       string `json:"name" binding:"required"`
}
