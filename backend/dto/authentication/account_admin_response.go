package authentication

import "time"

type AccountStatusResponse struct {
	AccountStatusID string `json:"account_status_id"`
	Code            string `json:"code"`
	Name            string `json:"name"`
	AllowsLogin     bool   `json:"allows_login"`
}

type AuthenticationPolicyResponse struct {
	AuthenticationPolicyID string `json:"authentication_policy_id"`
	Code                   string `json:"code"`
	Name                   string `json:"name"`
	MaxFailedAttempts      int    `json:"max_failed_attempts"`
	LockoutSeconds         int    `json:"lockout_seconds"`
	SessionTTLSeconds      int    `json:"session_ttl_seconds"`
	IsDefault              bool   `json:"is_default"`
}

type AccountSummaryResponse struct {
	AccountID                string                `json:"account_id"`
	Username                 string                `json:"username"`
	Email                    *string               `json:"email,omitempty"`
	DisplayName              string                `json:"display_name"`
	Status                   AccountStatusResponse `json:"status"`
	AuthenticationPolicyID   *string               `json:"authentication_policy_id,omitempty"`
	AuthenticationPolicyCode *string               `json:"authentication_policy_code,omitempty"`
	PreferredTimezone        *string               `json:"preferred_timezone,omitempty"`
	FailedLoginCount         int                   `json:"failed_login_count"`
	LockedUntil              *time.Time            `json:"locked_until,omitempty"`
	LastLoginAt              *time.Time            `json:"last_login_at,omitempty"`
	CreatedAt                time.Time             `json:"created_at"`
	UpdatedAt                time.Time             `json:"updated_at"`
	VersionNo                int                   `json:"version_no"`
}

type AccountRoleResponse struct {
	AccountID  string    `json:"account_id"`
	RoleID     string    `json:"role_id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	IsActive   bool      `json:"is_active"`
	AssignedAt time.Time `json:"assigned_at"`
	AssignedBy *string   `json:"assigned_by,omitempty"`
}

type AccountOwnerAccessResponse struct {
	OwnerID   string    `json:"owner_id"`
	OwnerCode string    `json:"owner_code"`
	OwnerName string    `json:"owner_name"`
	GrantedAt time.Time `json:"granted_at"`
	GrantedBy *string   `json:"granted_by,omitempty"`
}

type AccountWarehouseAccessResponse struct {
	WarehouseID   string    `json:"warehouse_id"`
	WarehouseCode string    `json:"warehouse_code"`
	WarehouseName string    `json:"warehouse_name"`
	GrantedAt     time.Time `json:"granted_at"`
	GrantedBy     *string   `json:"granted_by,omitempty"`
}

type AccountDetailResponse struct {
	AccountSummaryResponse
	ActiveSessionCount   int64                            `json:"active_session_count"`
	Roles                []AccountRoleResponse            `json:"roles"`
	DirectPermissions    []AccountPermissionResponse      `json:"direct_permissions"`
	EffectivePermissions []string                         `json:"effective_permissions"`
	OwnerAccess          []AccountOwnerAccessResponse     `json:"owner_access"`
	WarehouseAccess      []AccountWarehouseAccessResponse `json:"warehouse_access"`
}

type AccountPageResponse struct {
	Items      []AccountSummaryResponse `json:"items"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalItems int64                    `json:"total_items"`
	TotalPages int                      `json:"total_pages"`
}

type RoleResponse struct {
	RoleID      string               `json:"role_id"`
	Code        string               `json:"code"`
	Name        string               `json:"name"`
	Description *string              `json:"description,omitempty"`
	IsActive    bool                 `json:"is_active"`
	Permissions []PermissionResponse `json:"permissions"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	VersionNo   int                  `json:"version_no"`
}

type RolePageResponse struct {
	Items      []RoleResponse `json:"items"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalItems int64          `json:"total_items"`
	TotalPages int            `json:"total_pages"`
}
