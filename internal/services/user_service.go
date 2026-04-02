package services

import (
	"IFMS-BE-API/internal/model/request"
	"IFMS-BE-API/internal/repository"
	"IFMS-BE-API/internal/utils"
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/vippergod12/IFMS-BE/sql/generated"
	"go.uber.org/zap"
)

type UserService struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	logger  *zap.Logger
	user    *repository.UserRepository
}

func NewUserService(pool *pgxpool.Pool, queries *db.Queries, logger *zap.Logger, user *repository.UserRepository) *UserService {
	return &UserService{
		pool:    pool,
		queries: queries,
		logger:  logger,
		user:    repository.NewUserRepository(pool),
	}
}

func (s *UserService) Login(ctx context.Context, req request.LoginRequest, apiAccessKeySecret []byte) (string, error) {
	user, err := s.user.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("auth.login_failed")
		}
		return "", err
	}

	if !utils.IsPasswordMatch(req.Password, user.PasswordHash) {
		return "", errors.New("auth.login_failed")
	}

	token, _, err := s.GenerateToken(ctx, req.Username, user.Username, user.UserID, apiAccessKeySecret)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) GenerateToken(ctx context.Context, username, name string, userid pgtype.UUID, accessKeySecret []byte) (string, int64, error) {
	nonce := time.Now().UnixNano()

	claims := request.CustomClaims{
		Username: username,
		FullName: name,
		UserID:   userid,
		Nonce:    nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(accessKeySecret)
	if err != nil {
		return "", 0, err
	}

	if err := s.user.InsertUserTrackingHistory(ctx, db.InsertUserTrackingHistoryParams{
		UserID: userid,
		Nonce:  nonce,
	}); err != nil {
		return "", 0, err
	}

	if err := s.user.DeleteUserSessionsByUserId(ctx, userid); err != nil {
		return "", 0, err
	}

	expiredAt := time.Now().Add(time.Hour * 24 * 30).UnixMilli()
	sessionParams := db.InsertUserSessionParams{
		UserID:    userid,
		Token:     tokenString,
		IsDeleted: false,
		ExpiredAt: expiredAt,
	}
	if err := s.user.InsertUserSession(ctx, sessionParams); err != nil {
		return "", 0, err
	}

	return tokenString, nonce, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (db.User, error) {
	user, err := s.user.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, errors.New("user.not_found")
		}
		return db.User{}, err
	}
	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID pgtype.UUID) (db.User, error) {
	user, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, errors.New("user.not_found")
		}
		return db.User{}, err
	}
	return user, nil
}

func (s *UserService) ListUsersWithRole(ctx context.Context, limit, offset int32) ([]db.ListUsersWithRoleRow, int64, error) {
	users, err := s.user.ListUsersWithRole(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.user.CountUsers(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (s *UserService) ChangePasswordByAdmin(ctx context.Context, userID pgtype.UUID, req request.ChangePasswordByAdminRequest) error {
	user, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("user.not_found")
		}
		return err
	}

	if utils.IsPasswordMatch(req.NewPassword, user.PasswordHash) {
		return errors.New("auth.same_password")
	}

	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	return s.user.ChangePassword(ctx, userID, req.NewPassword, newHash)
}

func (s *UserService) ChangePasswordByUser(ctx context.Context, userID pgtype.UUID, req request.ChangePasswordByUserRequest) error {
	user, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("user.not_found")
		}
		return err
	}

	if !utils.IsPasswordMatch(req.OldPassword, user.PasswordHash) {
		return errors.New("auth.old_password_incorrect")
	}

	if req.OldPassword == req.NewPassword {
		return errors.New("auth.same_password")
	}

	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	return s.user.ChangePassword(ctx, userID, req.NewPassword, newHash)
}

func (s *UserService) UpdateUserInfo(ctx context.Context, userID pgtype.UUID, req request.UpdateUserRequest) (db.User, error) {
	existingUser, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, errors.New("user.not_found")
		}
		return db.User{}, err
	}

	if req.Email == "" {
		req.Email = existingUser.Email
	}
	if req.Phone == "" {
		req.Phone = existingUser.Phone
	}
	if req.Address == "" {
		req.Address = existingUser.Address
	}

	user, err := s.user.UpdateUserInfo(ctx, userID, req, existingUser.Username)
	if err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return db.User{}, err
	}

	return user, nil
}

func (s *UserService) InsertUserInfo(ctx context.Context, req request.CreateUserRequest) (db.User, error) {
	_, err := s.user.GetUserByUsername(ctx, req.Username)
	if err == nil {
		return db.User{}, errors.New("user.username_already_exists")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.User{}, err
	}

	user, err := s.user.InsertUserInfo(ctx, req)
	if err != nil {
		s.logger.Error("Failed to insert user", zap.Error(err))
		return db.User{}, err
	}

	return user, nil
}
