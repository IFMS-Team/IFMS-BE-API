package request

type CreateFloorRequest struct {
	BuildingID  string `json:"buildingId" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Image       string `json:"image"`
	MaximumRoom *int32 `json:"maximumRoom" binding:"omitempty,min=0"`
}

type UpdateFloorRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Status      *int32 `json:"status" binding:"omitempty,oneof=0 1"`
	MaximumRoom *int32 `json:"maximumRoom" binding:"omitempty,min=0"`
}
