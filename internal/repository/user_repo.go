package repository

import (
	"IFMS-BE-API/internal/model/request"
	"IFMS-BE-API/internal/utils"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/IFMS-Team/IFMS-BE/sql/generated"
)

type UserRepository struct {
	db      *pgxpool.Pool
	queries *db.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	queries := db.New(pool)
	return &UserRepository{
		db:      pool,
		queries: queries,
	}
}

func (r *UserRepository) DeleteUserSessionsByUserId(ctx context.Context, userid pgtype.UUID) error {
	return r.queries.DeleteUserSessionsByUserId(ctx, userid)
}

func (r *UserRepository) InsertUserSession(ctx context.Context, params db.InsertUserSessionParams) error {
	return r.queries.InsertUserSession(ctx, params)
}

func (r *UserRepository) InsertUserTrackingHistory(ctx context.Context, params db.InsertUserTrackingHistoryParams) error {
	return r.queries.InsertUserTrackingHistory(ctx, params)
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (db.User, error) {
	return r.queries.GetUserByUsername(ctx, username)
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID pgtype.UUID) (db.User, error) {
	return r.queries.GetUserByID(ctx, userID)
}

func (r *UserRepository) ListUsersWithRole(ctx context.Context, limit, offset int32) ([]db.ListUsersWithRoleRow, error) {
	return r.queries.ListUsersWithRole(ctx, db.ListUsersWithRoleParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *UserRepository) CountUsers(ctx context.Context) (int64, error) {
	return r.queries.CountUsers(ctx)
}

func (r *UserRepository) ChangePassword(ctx context.Context, userID pgtype.UUID, newPassword, newHash string) error {
	return r.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		UserID:       userID,
		Password:     newPassword,
		PasswordHash: newHash,
	})
}

func (r *UserRepository) UpdateUserInfo(ctx context.Context, userID pgtype.UUID, req request.UpdateUserRequest, currentUsername string) (db.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return db.User{}, err
	}
	defer func(tx pgx.Tx) {
		_ = tx.Rollback(ctx)
	}(tx)

	qtx := r.queries.WithTx(tx)

	user, err := qtx.UpdateUser(ctx, db.UpdateUserParams{
		UserID:   userID,
		Username: currentUsername,
		Email:    req.Email,
		FullName: req.FullName,
		Phone:    req.Phone,
		Address:  req.Address,
	})
	if err != nil {
		return db.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return db.User{}, err
	}

	return user, nil
}

func (r *UserRepository) InsertUserInfo(ctx context.Context, req request.CreateUserRequest) (db.User, error) {
	hashedPass, err := utils.HashPassword(req.Password)
	if err != nil {
		return db.User{}, err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return db.User{}, err
	}
	defer func(tx pgx.Tx) {
		_ = tx.Rollback(ctx)
	}(tx)

	qtx := r.queries.WithTx(tx)

	user, err := qtx.CreateUser(ctx, db.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		Password:     req.Password,
		PasswordHash: hashedPass,
		FullName:     req.FullName,
		Phone:        req.Phone,
		Address:      req.Address,
		Cccd:         req.CCCD,
		RoleID:       req.RoleID,
	})
	if err != nil {
		return db.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return db.User{}, err
	}

	return user, nil
}
