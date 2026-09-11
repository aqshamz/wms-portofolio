package authentication

type CreateAccountRequest struct {
	Username               string  `json:"username" binding:"required,max=100"`
	Email                  *string `json:"email" binding:"omitempty,email,max=254"`
	DisplayName            string  `json:"display_name" binding:"required,max=150"`
	Password               string  `json:"password" binding:"required,min=12,max=72"`
	AccountStatusID        *string `json:"account_status_id" binding:"omitempty,uuid"`
	AuthenticationPolicyID *string `json:"authentication_policy_id" binding:"omitempty,uuid"`
	PreferredTimezone      *string `json:"preferred_timezone" binding:"omitempty,max=50"`
}

type UpdateAccountRequest struct {
	Username               string  `json:"username" binding:"required,max=100"`
	Email                  *string `json:"email" binding:"omitempty,email,max=254"`
	DisplayName            string  `json:"display_name" binding:"required,max=150"`
	AuthenticationPolicyID *string `json:"authentication_policy_id" binding:"omitempty,uuid"`
	PreferredTimezone      *string `json:"preferred_timezone" binding:"omitempty,max=50"`
	ExpectedVersion        int     `json:"expected_version" binding:"required,min=1"`
}

type ChangeAccountStatusRequest struct {
	AccountStatusID string `json:"account_status_id" binding:"required,uuid"`
	ExpectedVersion int    `json:"expected_version" binding:"required,min=1"`
}

type ResetAccountPasswordRequest struct {
	Password        string `json:"password" binding:"required,min=12,max=72"`
	ExpectedVersion int    `json:"expected_version" binding:"required,min=1"`
}

type AccountVersionRequest struct {
	ExpectedVersion int `json:"expected_version" binding:"required,min=1"`
}

type AccountOwnerRequest struct {
	OwnerID string `json:"owner_id" binding:"required,uuid"`
}

type AccountWarehouseRequest struct {
	WarehouseID string `json:"warehouse_id" binding:"required,uuid"`
}

type CreateRoleRequest struct {
	Code          string   `json:"code" binding:"required,max=60"`
	Name          string   `json:"name" binding:"required,max=120"`
	Description   *string  `json:"description"`
	PermissionIDs []string `json:"permission_ids" binding:"omitempty,dive,uuid"`
}

type UpdateRoleRequest struct {
	Name            string  `json:"name" binding:"required,max=120"`
	Description     *string `json:"description"`
	ExpectedVersion int     `json:"expected_version" binding:"required,min=1"`
}

type ReplaceRolePermissionsRequest struct {
	PermissionIDs   *[]string `json:"permission_ids" binding:"required,dive,uuid"`
	ExpectedVersion int       `json:"expected_version" binding:"required,min=1"`
}

type ChangeRoleStatusRequest struct {
	IsActive        bool `json:"is_active"`
	ExpectedVersion int  `json:"expected_version" binding:"required,min=1"`
}

type AssignAccountRoleRequest struct {
	RoleID string `json:"role_id" binding:"required,uuid"`
}
