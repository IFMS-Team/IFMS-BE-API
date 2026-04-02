package request

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateUserRequest struct {
	Username string      `json:"username" binding:"required,min=6,max=29" example:"john_doe"`
	FullName string      `json:"full_name" binding:"required" example:"John Doe"`
	Email    string      `json:"email" binding:"required,email" example:"john@example.com"`
	Password string      `json:"password" binding:"required,min=6" example:"secret123"`
	RoleID   pgtype.UUID `json:"role_id" binding:"required" swaggertype:"string" example:"550e8400-e29b-41d4-a716-446655440000"`
	Phone    string      `json:"phone" binding:"required,min=8,max=12" example:"0912345678"`
	Address  string      `json:"address" binding:"required,max=255" example:"123 Main Street"`
	CCCD     string      `json:"cccd" binding:"required,len=12" example:"012345678901"`
	Status   *int32      `json:"status" binding:"omitempty,oneof=0 1" example:"1"`
}

type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=255" example:"john_doe"`
	Email    string `json:"email" binding:"omitempty,email" example:"john@example.com"`
	Phone    string `json:"phone" binding:"omitempty" example:"0912345678"`
	Address  string `json:"address" binding:"omitempty" example:"123 Main Street"`
	CCCD     string `json:"cccd" binding:"omitempty" example:"012345678901"`
	Status   int32  `json:"status" binding:"omitempty" example:"1"`
	FullName string `json:"full_name" binding:"omitempty" example:"John Doe"`
}

type ChangePasswordByUserRequest struct {
	OldPassword     string `json:"old_password" binding:"required,min=6" example:"oldpass123"`
	NewPassword     string `json:"new_password" binding:"required,min=6" example:"newpass456"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword" example:"newpass456"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=6,max=29" example:"admin01"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

type CustomClaims struct {
	Username string      `json:"username"`
	FullName string      `json:"full_name"`
	UserID   pgtype.UUID `json:"user_id"`
	Nonce    int64       `json:"nonce"`
	jwt.RegisteredClaims
}

type ChangePasswordByAdminRequest struct {
	NewPassword     string `json:"new_password" binding:"required,min=6" example:"newpass456"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword" example:"newpass456"`
}
