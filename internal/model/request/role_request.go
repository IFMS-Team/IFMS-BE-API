package request

type CreateRoleRequest struct {
	RoleName    string `json:"role_name" binding:"required,min=2,max=50" example:"manager"`
	Description string `json:"description" binding:"omitempty,max=255" example:"Manager role with limited access"`
}

type UpdateRoleRequest struct {
	RoleName    string `json:"role_name" binding:"omitempty,min=2,max=50" example:"manager"`
	Description string `json:"description" binding:"omitempty,max=255" example:"Updated description"`
}

type UpdatePermissionRequest struct {
	PermissionName string `json:"permission_name" binding:"omitempty,min=2,max=100" example:"create_user"`
	Description    string `json:"description" binding:"omitempty,max=255" example:"Updated description"`
	Code           string `json:"code" binding:"omitempty,min=2,max=100" example:"create_user"`
}

type AssignPermissionRequest struct {
	RoleID       string `json:"role_id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	PermissionID string `json:"permission_id" binding:"required,uuid" example:"660e8400-e29b-41d4-a716-446655440000"`
}

type CreatePermissionRequest struct {
	PermissionName string `json:"permission_name" binding:"required,min=2,max=100" example:"create_user"`
	Description    string `json:"description" binding:"omitempty,max=255" example:"Allows creating new users"`
	Code           string `json:"code" binding:"required,min=2,max=100" example:"create_user"`
}
