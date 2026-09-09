package outbound

import "time"

type OutboundValidationResultDetail struct {
	ID              string    `gorm:"column:validation_result_id;size:170;primaryKey"`
	ValidationRunID string    `gorm:"column:validation_run_id;size:140;not null"`
	RuleID          string    `gorm:"column:outbound_validation_rule_id;type:uuid;not null"`
	Passed          bool      `gorm:"column:passed;not null"`
	ResultMessage   *string   `gorm:"column:result_message;type:text"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
}

func (OutboundValidationResultDetail) TableName() string { return "outbound_validation_result_detail" }
