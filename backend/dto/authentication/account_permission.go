package authentication

import "time"

type GrantAccountPermissionRequest struct {
	PermissionID string `json:"permission_id" binding:"required,uuid"`
}

type PermissionResponse struct {
	PermissionID string `json:"permission_id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	ModuleCode   string `json:"module_code"`
}

type AccountPermissionResponse struct {
	AccountID    string    `json:"account_id"`
	PermissionID string    `json:"permission_id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	ModuleCode   string    `json:"module_code"`
	GrantedAt    time.Time `json:"granted_at"`
	GrantedBy    *string   `json:"granted_by,omitempty"`
}
