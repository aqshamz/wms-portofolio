package outbound

import "time"

type Shipment struct {
	ID               string     `gorm:"column:shipment_id;size:120;primaryKey"`
	DocumentTypeID   string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID         string     `gorm:"column:status_id;type:uuid;not null"`
	OwnerID          string     `gorm:"column:owner_id;type:uuid;not null"`
	CarrierServiceID *string    `gorm:"column:carrier_service_id;type:uuid"`
	WarehouseID      string     `gorm:"column:warehouse_id;type:uuid;not null"`
	BusinessDate     time.Time  `gorm:"column:business_date;type:date;not null"`
	RouteReference   *string    `gorm:"column:route_reference;size:100"`
	TrackingNumber   *string    `gorm:"column:tracking_number;size:150"`
	VehicleNumber    *string    `gorm:"column:vehicle_number;size:60"`
	SealNumber       *string    `gorm:"column:seal_number;size:60"`
	ShippedAt        *time.Time `gorm:"column:shipped_at"`
	DispatchedBy     *string    `gorm:"column:dispatched_by;type:uuid"`
	Notes            *string    `gorm:"column:notes;type:text"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy        string     `gorm:"column:created_by;type:uuid;not null"`
}

func (Shipment) TableName() string { return "shipment" }
