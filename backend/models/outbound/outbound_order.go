package outbound

import "time"

type OutboundOrder struct {
	ID                    string     `gorm:"column:outbound_id;size:120;primaryKey"`
	DocumentTypeID        string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID              string     `gorm:"column:status_id;type:uuid;not null"`
	OwnerID               string     `gorm:"column:owner_id;type:uuid;not null"`
	CustomerID            string     `gorm:"column:customer_id;type:uuid;not null"`
	WarehouseID           string     `gorm:"column:warehouse_id;type:uuid;not null"`
	BusinessDate          time.Time  `gorm:"column:business_date;type:date;not null"`
	RequestedShipAt       *time.Time `gorm:"column:requested_ship_at"`
	ExternalReference     *string    `gorm:"column:external_reference;size:120"`
	CustomerOrderNo       *string    `gorm:"column:customer_order_no;size:120"`
	ClientDeliveryOrderNo string     `gorm:"column:client_delivery_order_no;size:120;not null"`
	ShipToPartnerID       *string    `gorm:"column:ship_to_partner_id;type:uuid"`
	ShipToName            string     `gorm:"column:ship_to_name;size:150;not null"`
	ShipToAddress1        string     `gorm:"column:ship_to_address_1;size:255;not null"`
	ShipToAddress2        *string    `gorm:"column:ship_to_address_2;size:255"`
	ShipToCity            *string    `gorm:"column:ship_to_city;size:100"`
	ShipToProvince        *string    `gorm:"column:ship_to_province;size:100"`
	ShipToPostalCode      *string    `gorm:"column:ship_to_postal_code;size:20"`
	ShipToCountryCode     *string    `gorm:"column:ship_to_country_code;size:2"`
	Notes                 *string    `gorm:"column:notes;type:text"`
	ValidatedAt           *time.Time `gorm:"column:validated_at"`
	ValidatedBy           *string    `gorm:"column:validated_by;type:uuid"`
	AllocatedAt           *time.Time `gorm:"column:allocated_at"`
	CompletedAt           *time.Time `gorm:"column:completed_at"`
	CreatedAt             time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy             string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt             time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy             *string    `gorm:"column:updated_by;type:uuid"`
	VersionNo             int64      `gorm:"column:version_no;not null;default:1"`
}

func (OutboundOrder) TableName() string { return "outbound_order" }
