package outbound

import "time"

type DeliveryEvent struct {
	ID                      string    `gorm:"column:delivery_event_id;size:160;primaryKey"`
	DeliveryID              string    `gorm:"column:delivery_id;size:120;not null"`
	DeliveryEventTypeID     string    `gorm:"column:delivery_event_type_id;type:uuid;not null"`
	EventAt                 time.Time `gorm:"column:event_at;not null"`
	DeliveryFailureReasonID *string   `gorm:"column:delivery_failure_reason_id;type:uuid"`
	RecipientName           *string   `gorm:"column:recipient_name;size:150"`
	RecipientReference      *string   `gorm:"column:recipient_reference;size:100"`
	ProofReference          *string   `gorm:"column:proof_reference;size:200"`
	ProofURI                *string   `gorm:"column:proof_uri;type:text"`
	Notes                   *string   `gorm:"column:notes;type:text"`
	Latitude                *string   `gorm:"column:latitude;type:numeric(10,7)"`
	Longitude               *string   `gorm:"column:longitude;type:numeric(10,7)"`
	RecordedBy              string    `gorm:"column:recorded_by;type:uuid;not null"`
	CreatedAt               time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
}

func (DeliveryEvent) TableName() string { return "delivery_event" }
