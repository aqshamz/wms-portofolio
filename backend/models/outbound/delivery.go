package outbound

import "time"

type Delivery struct {
	ID                 string     `gorm:"column:delivery_id;size:120;primaryKey"`
	DocumentTypeID     string     `gorm:"column:document_type_id;type:uuid;not null"`
	StatusID           string     `gorm:"column:status_id;type:uuid;not null"`
	ShipmentID         string     `gorm:"column:shipment_id;size:120;not null;uniqueIndex:uq_delivery_shipment_outbound,priority:1"`
	OutboundID         string     `gorm:"column:outbound_id;size:120;not null;uniqueIndex:uq_delivery_shipment_outbound,priority:2"`
	BusinessDate       time.Time  `gorm:"column:business_date;type:date;not null"`
	PlannedDeliveryAt  *time.Time `gorm:"column:planned_delivery_at"`
	ArrivedAt          *time.Time `gorm:"column:arrived_at"`
	DeliveredAt        *time.Time `gorm:"column:delivered_at"`
	RecipientName      *string    `gorm:"column:recipient_name;size:150"`
	RecipientReference *string    `gorm:"column:recipient_reference;size:100"`
	ProofReference     *string    `gorm:"column:proof_reference;size:200"`
	ProofURI           *string    `gorm:"column:proof_uri;type:text"`
	Latitude           *string    `gorm:"column:latitude;type:numeric(10,7)"`
	Longitude          *string    `gorm:"column:longitude;type:numeric(10,7)"`
	Notes              *string    `gorm:"column:notes;type:text"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy          string     `gorm:"column:created_by;type:uuid;not null"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;not null;default:clock_timestamp()"`
	UpdatedBy          string     `gorm:"column:updated_by;type:uuid;not null"`
}

func (Delivery) TableName() string { return "delivery" }
