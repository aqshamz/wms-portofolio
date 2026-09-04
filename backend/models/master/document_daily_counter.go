package master

import "time"

type DocumentDailyCounter struct {
	DocumentTypeID string    `gorm:"column:document_type_id;type:uuid;primaryKey"`
	BusinessDate   time.Time `gorm:"column:business_date;type:date;primaryKey"`
	LastNumber     int64     `gorm:"column:last_number;not null"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:clock_timestamp()"`
}

func (DocumentDailyCounter) TableName() string { return "document_daily_counter" }
