package request

type CreateBuildingRequest struct {
	BuildingName        string `json:"building_name" binding:"required,max=255"`
	BuildingAddress     string `json:"building_address" binding:"required,max=255"`
	BuildingDescription string `json:"building_description" binding:"required,max=255"`
	BuildingImage       string `json:"building_image" binding:"required,max=255"`
	BuildingStatus      *int32 `json:"building_status" binding:"omitempty,oneof=0 1"`
	MaximumFloor        *int32 `json:"maximum_floor" binding:"omitempty,min=0"`
}

