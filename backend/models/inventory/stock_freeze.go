package inventory

import "time"

// StockFreeze exists only while its grand stock opname is unfinished.
type StockFreeze struct {
	OwnerID      string    `gorm:"column:owner_id;type:uuid;primaryKey"`
	WarehouseID  string    `gorm:"column:warehouse_id;type:uuid;primaryKey"`
	CycleCountID string    `gorm:"column:cycle_count_id;size:140;not null;uniqueIndex"`
	FrozenAt     time.Time `gorm:"column:frozen_at;not null;default:clock_timestamp()"`
}

func (StockFreeze) TableName() string { return "inventory_stock_freeze" }
