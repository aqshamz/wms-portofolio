package system

import "time"

type SchemaMigration struct {
	Version   int64     `gorm:"column:version;primaryKey"`
	Name      string    `gorm:"column:name;size:150;not null"`
	AppliedAt time.Time `gorm:"column:applied_at;not null;default:clock_timestamp()"`
}

func (SchemaMigration) TableName() string { return "schema_migration" }
