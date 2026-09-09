package outbound

import "time"

type CarrierDriver struct {
	ID            string    `gorm:"column:driver_id;type:uuid;primaryKey;default:gen_random_uuid()"`
	CarrierID     string    `gorm:"column:carrier_id;type:uuid;not null;uniqueIndex:uq_carrier_driver_code,priority:1"`
	AccountID     *string   `gorm:"column:account_id;type:uuid;uniqueIndex"`
	Code          string    `gorm:"column:code;size:40;not null;uniqueIndex:uq_carrier_driver_code,priority:2"`
	Name          string    `gorm:"column:name;size:150;not null"`
	PhoneNumber   *string   `gorm:"column:phone_number;size:50"`
	LicenseNumber *string   `gorm:"column:license_number;size:80"`
	IsActive      bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:clock_timestamp()"`
	CreatedBy     *string   `gorm:"column:created_by;type:uuid"`
}

func (CarrierDriver) TableName() string { return "carrier_driver" }
