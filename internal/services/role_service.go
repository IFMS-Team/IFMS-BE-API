package services

import (
	"IFMS-BE-API/internal/model/request"
	"IFMS-BE-API/internal/model/response"
	"IFMS-BE-API/internal/repository"
	"context"
	"errors"

	db "github.com/IFMS-Team/IFMS-BE/sql/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type RoleService struct {
	logger *zap.Logger
	role   *repository.RoleRepository
}

func NewRoleService(role *repository.RoleRepository, logger *zap.Logger) *RoleService {
	return &RoleService{
		logger: logger,
		role:   role,
	}
}

func (s *RoleService) ListRoles(ctx context.Context) ([]db.Role, error) {
	return s.role.List(ctx)
}

func (s *RoleService) GetRoleWithPermissions(ctx context.Context, roleID pgtype.UUID) (response.RoleWithPermissionsResponse, error) {
	role, err := s.role.GetByID(ctx, roleID)
	if err != nil {
		return response.RoleWithPermissionsResponse{}, err
	}

	perms, err := s.role.GetPermissionsByRoleID(ctx, roleID)
	if err != nil {
		return response.RoleWithPermissionsResponse{}, err
	}

	return response.ToRoleWithPermissions(role, perms), nil
}

func (s *RoleService) CreateRole(ctx context.Context, req request.CreateRoleRequest) (db.Role, error) {
	_, err := s.role.GetByName(ctx, req.RoleName)
	if err == nil {
		return db.Role{}, errors.New("role.already_exists")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.Role{}, err
	}

	role, err := s.role.Create(ctx, db.CreateRoleParams{
		RoleName:    req.RoleName,
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		s.logger.Error("Failed to create role", zap.Error(err))
		return db.Role{}, err
	}

	return role, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, roleID pgtype.UUID, req request.UpdateRoleRequest) (db.Role, error) {
	existing, err := s.role.GetByID(ctx, roleID)
	if err != nil {
		return db.Role{}, errors.New("role.not_found")
	}

	roleName := existing.RoleName
	if req.RoleName != "" {
		roleName = req.RoleName
	}
	desc := existing.Description.String
	if req.Description != "" {
		desc = req.Description
	}

	role, err := s.role.Update(ctx, db.UpdateRoleParams{
		RoleID:      roleID,
		RoleName:    roleName,
		Description: pgtype.Text{String: desc, Valid: desc != ""},
	})
	if err != nil {
		s.logger.Error("Failed to update role", zap.Error(err))
		return db.Role{}, err
	}
	return role, nil
}

func (s *RoleService) DeleteRole(ctx context.Context, roleID pgtype.UUID) error {
	_, err := s.role.GetByID(ctx, roleID)
	if err != nil {
		return errors.New("role.not_found")
	}
	return s.role.Delete(ctx, roleID)
}

func (s *RoleService) AssignPermission(ctx context.Context, roleID, permID pgtype.UUID) error {
	if _, err := s.role.GetByID(ctx, roleID); err != nil {
		return errors.New("role.not_found")
	}
	if _, err := s.role.GetPermissionByID(ctx, permID); err != nil {
		return errors.New("permission.not_found")
	}

	return s.role.AddPermissionToRole(ctx, db.AddPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: permID,
	})
}

func (s *RoleService) RemovePermission(ctx context.Context, roleID, permID pgtype.UUID) error {
	return s.role.RemovePermissionFromRole(ctx, db.RemovePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permID,
	})
}

func (s *RoleService) ListPermissions(ctx context.Context) ([]db.Permission, error) {
	return s.role.ListPermissions(ctx)
}

func (s *RoleService) CreatePermission(ctx context.Context, req request.CreatePermissionRequest) (db.Permission, error) {
	perm, err := s.role.CreatePermission(ctx, db.CreatePermissionParams{
		PermissionName: req.PermissionName,
		Description:    pgtype.Text{String: req.Description, Valid: req.Description != ""},
		Code:           req.Code,
	})
	if err != nil {
		s.logger.Error("Failed to create permission", zap.Error(err))
		return db.Permission{}, err
	}
	return perm, nil
}

func (s *RoleService) UpdatePermission(ctx context.Context, permID pgtype.UUID, req request.UpdatePermissionRequest) (db.Permission, error) {
	existing, err := s.role.GetPermissionByID(ctx, permID)
	if err != nil {
		return db.Permission{}, errors.New("permission.not_found")
	}

	permName := existing.PermissionName
	if req.PermissionName != "" {
		permName = req.PermissionName
	}
	desc := existing.Description.String
	if req.Description != "" {
		desc = req.Description
	}
	code := existing.Code
	if req.Code != "" {
		code = req.Code
	}

	perm, err := s.role.UpdatePermission(ctx, db.UpdatePermissionParams{
		PermissionID:   permID,
		PermissionName: permName,
		Description:    pgtype.Text{String: desc, Valid: desc != ""},
		Code:           code,
	})
	if err != nil {
		s.logger.Error("Failed to update permission", zap.Error(err))
		return db.Permission{}, err
	}
	return perm, nil
}

func (s *RoleService) DeletePermission(ctx context.Context, permID pgtype.UUID) error {
	return s.role.DeletePermission(ctx, permID)
}
