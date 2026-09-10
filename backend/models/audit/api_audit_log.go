package audit

import "time"

type APIAuditLog struct {
	ID         string    `gorm:"column:audit_log_id;type:uuid;default:gen_random_uuid();primaryKey"`
	RequestID  string    `gorm:"column:request_id;size:64;not null"`
	AccountID  *string   `gorm:"column:account_id;type:uuid"`
	Method     string    `gorm:"column:method;size:10;not null"`
	Route      string    `gorm:"column:route;size:255;not null"`
	StatusCode int       `gorm:"column:status_code;not null"`
	IPAddress  *string   `gorm:"column:ip_address;size:64"`
	UserAgent  *string   `gorm:"column:user_agent;size:500"`
	DurationMS int64     `gorm:"column:duration_ms;not null"`
	OccurredAt time.Time `gorm:"column:occurred_at;not null;default:clock_timestamp()"`
}

func (APIAuditLog) TableName() string { return "api_audit_log" }
