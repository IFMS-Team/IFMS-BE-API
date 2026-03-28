package model

import (
	"time"

	"github.com/google/uuid"
)

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=255"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone" binding:"required"`
	Address  string `json:"address" binding:"required"`
	CCCD     string `json:"cccd" binding:"required"`
	RoleID   string `json:"role_id" binding:"required,uuid"`
}

type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=255"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

type UserResponse struct {
	UserID    uuid.UUID  `json:"user_id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Status    int32      `json:"status"`
	Phone     string     `json:"phone"`
	Address   string     `json:"address"`
	CCCD      string     `json:"cccd"`
	RoleID    uuid.UUID  `json:"role_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
