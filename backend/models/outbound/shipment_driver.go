package outbound

import "time"

type ShipmentDriver struct {
	ShipmentID string    `gorm:"column:shipment_id;size:120;primaryKey"`
	DriverID   string    `gorm:"column:driver_id;type:uuid;primaryKey"`
	IsPrimary  bool      `gorm:"column:is_primary;not null;default:false"`
	AssignedAt time.Time `gorm:"column:assigned_at;not null;default:clock_timestamp()"`
	AssignedBy string    `gorm:"column:assigned_by;type:uuid;not null"`
}

func (ShipmentDriver) TableName() string { return "shipment_driver" }
