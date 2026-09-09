package outbound

import "time"

type OutboundWaveOrder struct {
	WaveID     string    `gorm:"column:wave_id;size:120;primaryKey"`
	OutboundID string    `gorm:"column:outbound_id;size:120;primaryKey"`
	AddedAt    time.Time `gorm:"column:added_at;not null;default:clock_timestamp()"`
	AddedBy    string    `gorm:"column:added_by;type:uuid;not null"`
}

func (OutboundWaveOrder) TableName() string { return "outbound_wave_order" }
