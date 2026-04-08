package services

import (
	"IFMS-BE-API/internal/model/request"
	"IFMS-BE-API/internal/repository"
	"IFMS-BE-API/internal/utils"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	db "github.com/IFMS-Team/IFMS-BE/sql/generated"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type UserService struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	logger  *zap.Logger
	user    *repository.UserRepository
	redis   *redis.Client
}

func NewUserService(pool *pgxpool.Pool, queries *db.Queries, logger *zap.Logger, user *repository.UserRepository, rds *redis.Client) *UserService {
	return &UserService{
		pool:    pool,
		queries: queries,
		logger:  logger,
		user:    repository.NewUserRepository(pool),
		redis:   rds,
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
	if req.FullName == "" {
		req.FullName = existingUser.FullName
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

func (s *UserService) ForgotPassword(ctx context.Context, req request.ForgotPasswordRequest) error {
	if s.redis == nil {
		return errors.New("service.redis_unavailable")
	}

	user, err := s.user.GetUserByUsername(ctx, strings.ToLower(strings.TrimSpace(req.Username)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("user.identity_mismatch")
		}
		return err
	}

	if !strings.EqualFold(user.Email, req.Email) ||
		user.Phone != req.Phone ||
		user.Cccd != req.CCCD {
		return errors.New("user.identity_mismatch")
	}

	attemptsKey := fmt.Sprintf("otp_attempts:%s", user.Email)
	attempts, _ := s.redis.Get(ctx, attemptsKey).Int()
	if attempts >= 3 {
		return errors.New("otp.max_attempts_reached")
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		return err
	}

	otpKey := fmt.Sprintf("otp:%s", user.Email)
	if err := s.redis.Set(ctx, otpKey, otp, 5*time.Minute).Err(); err != nil {
		return err
	}

	if err := s.redis.Set(ctx, attemptsKey, 0, 5*time.Minute).Err(); err != nil {
		return err
	}

	if err := utils.SendOTPEmail(user.Email, otp); err != nil {
		s.logger.Error("Failed to send OTP email", zap.Error(err))
		s.redis.Del(ctx, otpKey)
		return errors.New("otp.send_failed")
	}

	return nil
}

func (s *UserService) VerifyOTP(ctx context.Context, req request.VerifyOTPRequest) (string, error) {
	if s.redis == nil {
		return "", errors.New("service.redis_unavailable")
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	attemptsKey := fmt.Sprintf("otp_attempts:%s", email)
	attempts, _ := s.redis.Get(ctx, attemptsKey).Int()
	if attempts >= 3 {
		return "", errors.New("otp.max_attempts_reached")
	}

	otpKey := fmt.Sprintf("otp:%s", email)
	storedOTP, err := s.redis.Get(ctx, otpKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", errors.New("otp.expired")
		}
		return "", err
	}

	if storedOTP != req.OTP {
		s.redis.Incr(ctx, attemptsKey)
		remaining := 3 - (attempts + 1)
		return "", fmt.Errorf("otp.invalid:%d", remaining)
	}

	s.redis.Del(ctx, otpKey, attemptsKey)

	resetToken := uuid.New().String()
	resetKey := fmt.Sprintf("reset_token:%s", resetToken)
	if err := s.redis.Set(ctx, resetKey, email, 10*time.Minute).Err(); err != nil {
		return "", err
	}

	return resetToken, nil
}

func (s *UserService) ResetPassword(ctx context.Context, req request.ResetPasswordRequest) error {
	if s.redis == nil {
		return errors.New("service.redis_unavailable")
	}

	if valid, msg := utils.ValidatePasswordStrength(req.NewPassword); !valid {
		return fmt.Errorf("password.weak:%s", msg)
	}

	resetKey := fmt.Sprintf("reset_token:%s", req.ResetToken)
	email, err := s.redis.Get(ctx, resetKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errors.New("reset.token_expired")
		}
		return err
	}

	user, err := s.queries.GetUserByEmail(ctx, email)
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

	if err := s.user.ChangePassword(ctx, user.UserID, req.NewPassword, newHash); err != nil {
		return err
	}

	s.redis.Del(ctx, resetKey)

	return nil
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
