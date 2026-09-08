package inbound

import "time"

type QualityInspection struct {
	ID                 string     `gorm:"column:inspection_id;size:120;primaryKey"`
	ReceiptInventoryID string     `gorm:"column:receipt_inventory_id;size:160;not null"`
	ParentInspectionID *string    `gorm:"column:parent_inspection_id;size:120"`
	SourceBalanceID    *string    `gorm:"column:source_balance_id;size:160"`
	QualityStatusID    string     `gorm:"column:quality_status_id;type:uuid;not null"`
	InspectionResultID *string    `gorm:"column:inspection_result_id;type:uuid"`
	InspectedQty       string     `gorm:"column:inspected_qty;type:numeric(20,6);not null"`
	PassedQty          string     `gorm:"column:passed_qty;type:numeric(20,6);not null;default:0"`
	FailedQty          string     `gorm:"column:failed_qty;type:numeric(20,6);not null;default:0"`
	InspectedAt        *time.Time `gorm:"column:inspected_at"`
	InspectedBy        *string    `gorm:"column:inspected_by;type:uuid"`
	CancelledAt        *time.Time `gorm:"column:cancelled_at"`
	CancelledBy        *string    `gorm:"column:cancelled_by;type:uuid"`
	CancellationReason *string    `gorm:"column:cancellation_reason;type:text"`
	Notes              *string    `gorm:"column:notes;type:text"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy          string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy          *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo          int64      `gorm:"column:version_no;not null;default:1"`
}

func (QualityInspection) TableName() string { return "quality_inspection" }
