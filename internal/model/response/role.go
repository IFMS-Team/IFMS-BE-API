package response

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vippergod12/IFMS-BE/sql/generated"
)

type RoleResponse struct {
	RoleID      string `json:"role_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoleName    string `json:"role_name" example:"admin"`
	Description string `json:"description" example:"Administrator role"`
}

type RoleWithPermissionsResponse struct {
	RoleID      string               `json:"role_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoleName    string               `json:"role_name" example:"admin"`
	Description string               `json:"description" example:"Administrator role"`
	Permissions []PermissionResponse `json:"permissions"`
}

type PermissionResponse struct {
	PermissionID   string `json:"permission_id" example:"660e8400-e29b-41d4-a716-446655440000"`
	PermissionName string `json:"permission_name" example:"create_user"`
	Description    string `json:"description" example:"Allows creating new users"`
}

func ToRoleResponse(r db.Role) RoleResponse {
	return RoleResponse{
		RoleID:      uuid.UUID(r.RoleID.Bytes).String(),
		RoleName:    r.RoleName,
		Description: r.Description.String,
	}
}

func ToRoleListResponse(roles []db.Role) []RoleResponse {
	result := make([]RoleResponse, len(roles))
	for i, r := range roles {
		result[i] = ToRoleResponse(r)
	}
	return result
}

func ToPermissionResponse(p db.Permission) PermissionResponse {
	return PermissionResponse{
		PermissionID:   uuid.UUID(p.PermissionID.Bytes).String(),
		PermissionName: p.PermissionName,
		Description:    p.Description.String,
	}
}

func ToPermissionListResponse(perms []db.Permission) []PermissionResponse {
	result := make([]PermissionResponse, len(perms))
	for i, p := range perms {
		result[i] = ToPermissionResponse(p)
	}
	return result
}

func ToRoleWithPermissions(r db.Role, perms []db.Permission) RoleWithPermissionsResponse {
	return RoleWithPermissionsResponse{
		RoleID:      uuid.UUID(r.RoleID.Bytes).String(),
		RoleName:    r.RoleName,
		Description: r.Description.String,
		Permissions: ToPermissionListResponse(perms),
	}
}

func StringToUUID(s string) (pgtype.UUID, error) {
	uid, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: uid, Valid: true}, nil
}
