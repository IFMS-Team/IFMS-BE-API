package request

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateUserRequest struct {
	Username string      `json:"username" binding:"required,min=6,max=29"`
	FullName string      `json:"full_name" binding:"required"`
	Email    string      `json:"email" binding:"required,email"`
	Password string      `json:"password" binding:"required,min=6"`
	RoleID   pgtype.UUID `json:"role_id" binding:"required"`
	Phone    string      `json:"phone" binding:"required,min=8,max=12"`
	Address  string      `json:"address" binding:"required,max=255"`
	CCCD     string      `json:"cccd" binding:"required,len=12"`
	Status   *int32      `json:"status" binding:"omitempty,oneof=0 1"`
}

type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=255"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty"`
	Address  string `json:"address" binding:"omitempty"`
	CCCD     string `json:"cccd" binding:"omitempty"`
	Status   int32  `json:"status" binding:"omitempty"`
	FullName string `json:"full_name" binding:"omitempty"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=6,max=29"`
	Password string `json:"password" binding:"required,min=6"`
}

type CustomClaims struct {
	Username string      `json:"username"`
	FullName string      `json:"full_name"`
	UserID   pgtype.UUID `json:"user_id"`
	Nonce    int64       `json:"nonce"`
	jwt.RegisteredClaims
}
