package response

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type UserInfoResponse struct {
	UserID    pgtype.UUID      `json:"user_id" swaggertype:"string" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username  string           `json:"username" example:"john_doe"`
	Email     string           `json:"email" example:"john@example.com"`
	Status    int32            `json:"status" example:"1"`
	Phone     string           `json:"phone" example:"0912345678"`
	Address   string           `json:"address" example:"123 Main Street"`
	CCCD      string           `json:"cccd" example:"012345678901"`
	RoleID    pgtype.UUID      `json:"role_id" swaggertype:"string" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreatedAt pgtype.Timestamp `json:"created_at" swaggertype:"string" example:"2024-01-01T00:00:00Z"`
	UpdatedAt pgtype.Timestamp `json:"updated_at" swaggertype:"string" example:"2024-01-01T00:00:00Z"`
}

type UserListResponse struct {
	UserID    pgtype.UUID      `json:"user_id" swaggertype:"string" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username  string           `json:"username" example:"john_doe"`
	Email     string           `json:"email" example:"john@example.com"`
	Status    int32            `json:"status" example:"1"`
	Phone     string           `json:"phone" example:"0912345678"`
	Address   string           `json:"address" example:"123 Main Street"`
	CCCD      string           `json:"cccd" example:"012345678901"`
	RoleID    pgtype.UUID      `json:"role_id" swaggertype:"string" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreatedAt pgtype.Timestamp `json:"created_at" swaggertype:"string" example:"2024-01-01T00:00:00Z"`
	UpdatedAt pgtype.Timestamp `json:"updated_at" swaggertype:"string" example:"2024-01-01T00:00:00Z"`
}

type LoginResponse struct {
	Token string           `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  UserInfoResponse `json:"user"`
}
