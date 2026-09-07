package inventory

import "time"

type HandlingUnit struct {
	ID                   string    `gorm:"column:handling_unit_id;size:120;primaryKey"`
	WarehouseID          string    `gorm:"column:warehouse_id;type:uuid;not null"`
	OwnerID              string    `gorm:"column:owner_id;type:uuid;not null"`
	HandlingUnitTypeID   string    `gorm:"column:handling_unit_type_id;type:uuid;not null"`
	ParentHandlingUnitID *string   `gorm:"column:parent_handling_unit_id;size:120"`
	CurrentLocationID    *string   `gorm:"column:current_location_id;type:uuid"`
	Barcode              string    `gorm:"column:barcode;size:120;not null"`
	IsClosed             bool      `gorm:"column:is_closed;not null;default:false"`
	CreatedAt            time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy            *string   `gorm:"column:created_by;type:uuid"`
}

func (HandlingUnit) TableName() string { return "handling_unit" }
