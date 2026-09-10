package authentication

import "time"

type UserResponse struct {
	AccountID         string     `json:"account_id"`
	Username          string     `json:"username"`
	Email             *string    `json:"email,omitempty"`
	DisplayName       string     `json:"display_name"`
	PreferredTimezone *string    `json:"preferred_timezone,omitempty"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	Permissions       []string   `json:"permissions"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	TokenType string       `json:"token_type"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}
