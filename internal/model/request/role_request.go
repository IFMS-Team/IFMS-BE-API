package request

type CreateRoleRequest struct {
	RoleName    string `json:"role_name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

type UpdateRoleRequest struct {
	RoleName    string `json:"role_name" binding:"omitempty,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

type AssignPermissionRequest struct {
	RoleID       string `json:"role_id" binding:"required,uuid"`
	PermissionID string `json:"permission_id" binding:"required,uuid"`
}

type CreatePermissionRequest struct {
	PermissionName string `json:"permission_name" binding:"required,min=2,max=100"`
	Description    string `json:"description" binding:"omitempty,max=255"`
}
