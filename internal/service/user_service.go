package service

import (
	"context"
	"errors"

	"IFMS-BE-API/internal/model"
	"IFMS-BE-API/internal/repo"
	db "github.com/vippergod12/IFMS-BE/sql/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repo.UserRepo
}

func NewUserService(repo *repo.UserRepo) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetByID(ctx context.Context, id string) (db.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return db.User{}, errors.New("invalid user id")
	}
	return s.repo.GetByID(ctx, pgtype.UUID{Bytes: uid, Valid: true})
}

func (s *UserService) List(ctx context.Context, page, limit int32) ([]db.User, error) {
	offset := (page - 1) * limit
	return s.repo.List(ctx, limit, offset)
}

func (s *UserService) Create(ctx context.Context, req model.CreateUserRequest) (db.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, errors.New("failed to hash password")
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return db.User{}, errors.New("invalid role id")
	}

	return s.repo.Create(ctx, db.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		Password:     req.Password,
		PasswordHash: string(hashedPassword),
		Phone:        req.Phone,
		Address:      req.Address,
		Cccd:         req.CCCD,
		RoleID:       pgtype.UUID{Bytes: roleID, Valid: true},
	})
}

func (s *UserService) Update(ctx context.Context, id string, req model.UpdateUserRequest) (db.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return db.User{}, errors.New("invalid user id")
	}

	return s.repo.Update(ctx, db.UpdateUserParams{
		UserID:   pgtype.UUID{Bytes: uid, Valid: true},
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Address:  req.Address,
	})
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid user id")
	}
	return s.repo.Delete(ctx, pgtype.UUID{Bytes: uid, Valid: true})
}
