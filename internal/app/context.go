package app

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/IFMS-Team/IFMS-BE/database"
	db "github.com/IFMS-Team/IFMS-BE/sql/generated"
)

type AppContext struct {
	Engine     *gin.Engine
	DbPostgres *database.DbPostgres
	Pool       *pgxpool.Pool
	Queries    *db.Queries
	Logger     *zap.Logger
	Validator  *validator.Validate
	Redis      *redis.Client
	KeySecret  []byte
}
