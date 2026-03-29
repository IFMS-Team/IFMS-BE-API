package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/vippergod12/IFMS-BE/sql/generated"
)

type RoleRepository struct {
	db      *pgxpool.Pool
	queries *db.Queries
}

func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{
		db:      pool,
		queries: db.New(pool),
	}
}

func (r *RoleRepository) GetByID(ctx context.Context, roleID pgtype.UUID) (db.Role, error) {
	return r.queries.GetRoleByID(ctx, roleID)
}

func (r *RoleRepository) GetByName(ctx context.Context, name string) (db.Role, error) {
	return r.queries.GetRoleByName(ctx, name)
}

func (r *RoleRepository) List(ctx context.Context) ([]db.Role, error) {
	return r.queries.ListRoles(ctx)
}

func (r *RoleRepository) Create(ctx context.Context, params db.CreateRoleParams) (db.Role, error) {
	return r.queries.CreateRole(ctx, params)
}

func (r *RoleRepository) Delete(ctx context.Context, roleID pgtype.UUID) error {
	return r.queries.DeleteRole(ctx, roleID)
}

func (r *RoleRepository) GetPermissionsByRoleID(ctx context.Context, roleID pgtype.UUID) ([]db.Permission, error) {
	return r.queries.GetPermissionsByRoleID(ctx, roleID)
}

func (r *RoleRepository) AddPermissionToRole(ctx context.Context, params db.AddPermissionToRoleParams) error {
	return r.queries.AddPermissionToRole(ctx, params)
}

func (r *RoleRepository) RemovePermissionFromRole(ctx context.Context, params db.RemovePermissionFromRoleParams) error {
	return r.queries.RemovePermissionFromRole(ctx, params)
}

func (r *RoleRepository) ListPermissions(ctx context.Context) ([]db.Permission, error) {
	return r.queries.ListPermissions(ctx)
}

func (r *RoleRepository) GetPermissionByID(ctx context.Context, permID pgtype.UUID) (db.Permission, error) {
	return r.queries.GetPermissionByID(ctx, permID)
}

func (r *RoleRepository) CreatePermission(ctx context.Context, params db.CreatePermissionParams) (db.Permission, error) {
	return r.queries.CreatePermission(ctx, params)
}

func (r *RoleRepository) DeletePermission(ctx context.Context, permID pgtype.UUID) error {
	return r.queries.DeletePermission(ctx, permID)
}
