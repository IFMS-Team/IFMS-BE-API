package response

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type UserInfoResponse struct {
	UserID    pgtype.UUID        `json:"user_id"`
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	FullName  string             `json:"full_name"`
	Status    string             `json:"status"`
	Phone     string             `json:"phone"`
	Address   string             `json:"address"`
	CCCD      string             `json:"cccd"`
	RoleID    pgtype.UUID        `json:"role_id"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

type UserListResponse struct {
	UserID    pgtype.UUID        `json:"user_id"`
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	FullName  string             `json:"full_name"`
	Status    string             `json:"status"`
	Phone     string             `json:"phone"`
	Address   string             `json:"address"`
	CCCD      string             `json:"cccd"`
	RoleID    pgtype.UUID        `json:"role_id"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

type LoginResponse struct {
	Token string           `json:"token"`
	User  UserInfoResponse `json:"user"`
}
