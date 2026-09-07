package inventory

import (
	"context"
	"gorm.io/gorm"
	"time"
	model "wms-api/models/inventory"
)

type MovementRow struct {
	model.InventoryMovement
	MovementTypeCode, ItemCode, UOMCode                                       string
	LotNumber, FromLocationCode, ToLocationCode, FromStatusCode, ToStatusCode *string
}
type MovementFilter struct {
	OwnerID, WarehouseID, ItemID, MovementTypeID, SourceDocumentID, OperationKey, Search string
	OccurredFrom, OccurredUntil                                                          *time.Time
	Page, PageSize                                                                       int
}
type InventoryMovementRepository struct{ db *gorm.DB }

func NewInventoryMovementRepository(db *gorm.DB) *InventoryMovementRepository {
	return &InventoryMovementRepository{db: db}
}
func (r *InventoryMovementRepository) Create(ctx context.Context, v *model.InventoryMovement) error {
	return Error(r.db.WithContext(ctx).Create(v).Error)
}

const movementSelect = `m.*,mt.code movement_type_code,i.code item_code,lot.lot_number,u.code uom_code,
 fl.code from_location_code,tl.code to_location_code,fs.code from_status_code,ts.code to_status_code`

func movementQuery(db *gorm.DB) *gorm.DB {
	return db.Table("inventory_movement m").Select(movementSelect).
		Joins("JOIN movement_type mt ON mt.movement_type_id=m.movement_type_id").Joins("JOIN item i ON i.item_id=m.item_id").
		Joins("LEFT JOIN inventory_lot lot ON lot.lot_id=m.lot_id").Joins("JOIN uom u ON u.uom_id=m.uom_id").
		Joins("LEFT JOIN warehouse_location fl ON fl.location_id=m.from_location_id").Joins("LEFT JOIN warehouse_location tl ON tl.location_id=m.to_location_id").
		Joins("LEFT JOIN inventory_status fs ON fs.inventory_status_id=m.from_status_id").Joins("LEFT JOIN inventory_status ts ON ts.inventory_status_id=m.to_status_id")
}
func (r *InventoryMovementRepository) Get(ctx context.Context, id string) (MovementRow, error) {
	var v MovementRow
	err := movementQuery(r.db.WithContext(ctx)).Where("m.movement_id=?", id).Take(&v).Error
	return v, Error(err)
}
func (r *InventoryMovementRepository) GetByOperationKey(ctx context.Context, key string) (MovementRow, error) {
	var v MovementRow
	err := movementQuery(r.db.WithContext(ctx)).Where("m.operation_key=?", key).Take(&v).Error
	return v, Error(err)
}
func (r *InventoryMovementRepository) List(ctx context.Context, f MovementFilter) ([]MovementRow, int64, error) {
	q := movementQuery(r.db.WithContext(ctx))
	for col, val := range map[string]string{"m.owner_id": f.OwnerID, "m.warehouse_id": f.WarehouseID, "m.item_id": f.ItemID, "m.movement_type_id": f.MovementTypeID, "m.source_document_id": f.SourceDocumentID, "m.operation_key": f.OperationKey} {
		if val != "" {
			q = q.Where(col+"=?", val)
		}
	}
	if f.OccurredFrom != nil {
		q = q.Where("m.occurred_at>=?", *f.OccurredFrom)
	}
	if f.OccurredUntil != nil {
		q = q.Where("m.occurred_at<?", *f.OccurredUntil)
	}
	if f.Search != "" {
		wild := "%" + f.Search + "%"
		q = q.Where("(i.code ILIKE ? OR m.source_document_id ILIKE ? OR lot.lot_number ILIKE ?)", wild, wild, wild)
	}
	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, Error(err)
	}
	rows := make([]MovementRow, 0)
	err := q.Order("m.occurred_at DESC,m.movement_id DESC").Limit(f.PageSize).Offset((f.Page - 1) * f.PageSize).Find(&rows).Error
	return rows, total, Error(err)
}
