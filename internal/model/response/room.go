package response

import (
	"time"

	"github.com/google/uuid"
	db "github.com/vippergod12/IFMS-BE/sql/generated"
)

type RoomResponse struct {
	ID          string    `json:"id"`
	FloorID     string    `json:"floorId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	Status      string    `json:"status"`
	QrCode      string    `json:"qrCode"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedBy   string    `json:"createdBy"`
	UpdatedBy   string    `json:"updatedBy"`
}

func ToRoomResponse(r db.Room) RoomResponse {
	qr := ""
	if r.QrCode.Valid {
		qr = r.QrCode.String
	}
	return RoomResponse{
		ID:          uuid.UUID(r.RoomID.Bytes).String(),
		FloorID:     uuid.UUID(r.FloorID.Bytes).String(),
		Name:        r.RoomName,
		Description: r.RoomDescription,
		Image:       r.RoomImage,
		Status:      r.RoomStatus,
		QrCode:      qr,
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
		CreatedBy:   uuid.UUID(r.CreatedBy.Bytes).String(),
		UpdatedBy:   uuid.UUID(r.UpdatedBy.Bytes).String(),
	}
}

func ToRoomListResponse(rooms []db.Room) []RoomResponse {
	list := make([]RoomResponse, 0, len(rooms))
	for _, r := range rooms {
		list = append(list, ToRoomResponse(r))
	}
	return list
}
