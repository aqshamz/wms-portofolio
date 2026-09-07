package inventory

import "time"

type InventoryLot struct {
	ID              string     `gorm:"column:lot_id;size:120;primaryKey"`
	OwnerID         string     `gorm:"column:owner_id;type:uuid;not null"`
	ItemID          string     `gorm:"column:item_id;type:uuid;not null"`
	LotNumber       string     `gorm:"column:lot_number;size:100;not null"`
	ManufactureDate *time.Time `gorm:"column:manufacture_date;type:date"`
	ExpiryDate      *time.Time `gorm:"column:expiry_date;type:date"`
	QualityStatusID *string    `gorm:"column:quality_status_id;type:uuid"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy       *string    `gorm:"column:created_by;type:uuid"`
}

func (InventoryLot) TableName() string { return "inventory_lot" }
