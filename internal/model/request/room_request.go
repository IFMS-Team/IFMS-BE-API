package request

type CreateRoomRequest struct {
	FloorID     string `json:"floorId" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Status      string `json:"status"`
}

type UpdateRoomRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Status      string `json:"status"`
}
