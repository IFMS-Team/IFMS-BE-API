package response

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type UserInfoResponse struct {
	UserID    pgtype.UUID      `json:"user_id"`
	Username  string           `json:"username"`
	Email     string           `json:"email"`
	Status    int32            `json:"status"`
	Phone     string           `json:"phone"`
	Address   string           `json:"address"`
	CCCD      string           `json:"cccd"`
	RoleID    pgtype.UUID      `json:"role_id"`
	CreatedAt pgtype.Timestamp `json:"created_at"`
	UpdatedAt pgtype.Timestamp `json:"updated_at"`
}

type UserListResponse struct {
	UserID    pgtype.UUID      `json:"user_id"`
	Username  string           `json:"username"`
	Email     string           `json:"email"`
	Status    int32            `json:"status"`
	Phone     string           `json:"phone"`
	Address   string           `json:"address"`
	CCCD      string           `json:"cccd"`
	RoleID    pgtype.UUID      `json:"role_id"`
	CreatedAt pgtype.Timestamp `json:"created_at"`
	UpdatedAt pgtype.Timestamp `json:"updated_at"`
}

type LoginResponse struct {
	Token string           `json:"token"`
	User  UserInfoResponse `json:"user"`
}
