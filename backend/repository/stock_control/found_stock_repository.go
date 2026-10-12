package stockcontrol

import (
	"context"
	model "wms-api/models/stock_control"
	inventoryrepo "wms-api/repository/inventory"
)

type FoundStockItemOption struct {
	ID, Code, Name, UOMCode         string
	LotControlled, SerialControlled bool
}
type FoundStockOption struct{ ID, Label string }
type FoundStockOptions struct {
	Items                                    []FoundStockItemOption
	Locations, Statuses, Lots, HandlingUnits []FoundStockOption
}
type FoundStockOptionsFilter struct {
	ItemID, LocationID, ItemSearch, LocationSearch, LotSearch, HUSearch string
}

// Scope comes from the authorized count document, never from lookup parameters.
func (r *CycleCountRepository) FoundStockOptions(ctx context.Context, header model.CycleCount, f FoundStockOptionsFilter) (FoundStockOptions, error) {
	result := FoundStockOptions{}
	db := r.db.WithContext(ctx)
	items := db.Table("item i").Select("i.item_id id,i.code,i.name,u.code uom_code,i.lot_controlled,i.serial_controlled").
		Joins("JOIN uom u ON u.uom_id=i.base_uom_id AND u.is_active").Where("i.owner_id=? AND i.is_active", header.OwnerID)
	if f.ItemSearch != "" {
		items = items.Where("(i.code ILIKE ? OR i.name ILIKE ? OR i.item_id::text=?)", "%"+f.ItemSearch+"%", "%"+f.ItemSearch+"%", f.ItemID)
	}
	if err := items.Order("i.code").Limit(100).Scan(&result.Items).Error; err != nil {
		return result, inventoryrepo.Error(err)
	}
	if f.ItemID != "" {
		present := false
		for _, row := range result.Items {
			present = present || row.ID == f.ItemID
		}
		if !present {
			var selected FoundStockItemOption
			if err := items.Where("i.item_id=?", f.ItemID).Limit(1).Scan(&selected).Error; err != nil {
				return result, inventoryrepo.Error(err)
			}
			if selected.ID != "" {
				result.Items = append(result.Items, selected)
			}
		}
	}
	types := []string{"STORAGE"}
	if header.CountTypeCode == "GRAND" {
		types = append(types, "PICK_FACE")
	}
	locations := db.Table("warehouse_location l").Select("l.location_id id,l.code label").
		Joins("JOIN warehouse_zone z ON z.zone_id=l.zone_id AND z.is_active").
		Joins("JOIN location_type t ON t.location_type_id=l.location_type_id AND t.is_active").
		Where("l.warehouse_id=? AND l.is_active AND NOT l.is_locked AND t.code IN ?", header.WarehouseID, types)
	if f.LocationSearch != "" {
		locations = locations.Where("(l.code ILIKE ? OR l.location_id::text=?)", "%"+f.LocationSearch+"%", f.LocationID)
	}
	if err := locations.Order("l.code").Limit(100).Scan(&result.Locations).Error; err != nil {
		return result, inventoryrepo.Error(err)
	}
	if f.LocationID != "" {
		present := false
		for _, row := range result.Locations {
			present = present || row.ID == f.LocationID
		}
		if !present {
			var selected FoundStockOption
			if err := locations.Where("l.location_id=?", f.LocationID).Limit(1).Scan(&selected).Error; err != nil {
				return result, inventoryrepo.Error(err)
			}
			if selected.ID != "" {
				result.Locations = append(result.Locations, selected)
			}
		}
	}
	if err := db.Table("inventory_status").Select("inventory_status_id id,code label").Where("is_active").Order("code").Scan(&result.Statuses).Error; err != nil {
		return result, inventoryrepo.Error(err)
	}
	if f.ItemID != "" {
		lots := db.Table("inventory_lot lot").Select("lot.lot_id id,lot.lot_number label").Joins("JOIN item ON item.item_id=lot.item_id AND item.owner_id=lot.owner_id AND item.is_active").Where("lot.owner_id=? AND lot.item_id=?", header.OwnerID, f.ItemID)
		if f.LotSearch != "" {
			lots = lots.Where("lot.lot_number ILIKE ?", "%"+f.LotSearch+"%")
		}
		if err := lots.Order("lot.lot_number").Limit(100).Scan(&result.Lots).Error; err != nil {
			return result, inventoryrepo.Error(err)
		}
	}
	if f.LocationID != "" {
		hus := db.Table("handling_unit").Select("handling_unit_id id,barcode label").Where("owner_id=? AND warehouse_id=? AND current_location_id=? AND NOT is_closed AND parent_handling_unit_id IS NULL", header.OwnerID, header.WarehouseID, f.LocationID)
		if f.HUSearch != "" {
			hus = hus.Where("barcode ILIKE ?", "%"+f.HUSearch+"%")
		}
		if err := hus.Order("barcode").Limit(100).Scan(&result.HandlingUnits).Error; err != nil {
			return result, inventoryrepo.Error(err)
		}
	}
	return result, nil
}

func (r *CycleCountRepository) HasLineIdentity(ctx context.Context, id string, k inventoryrepo.BalanceIdentity) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.CycleCountLine{}).Where("cycle_count_id=? AND item_id=? AND location_id=? AND inventory_status_id=? AND lot_id IS NOT DISTINCT FROM ? AND handling_unit_id IS NOT DISTINCT FROM ?", id, k.ItemID, k.LocationID, k.InventoryStatusID, k.LotID, k.HandlingUnitID).Count(&count).Error
	return count > 0, inventoryrepo.Error(err)
}
func (r *CycleCountRepository) AddLine(ctx context.Context, line *model.CycleCountLine) error {
	return inventoryrepo.Error(r.db.WithContext(ctx).Create(line).Error)
}
func (r *CycleCountRepository) NextLineNo(ctx context.Context, id string) (int, error) {
	var next int
	err := r.db.WithContext(ctx).Model(&model.CycleCountLine{}).Select("COALESCE(MAX(line_no),0)+1").Where("cycle_count_id=?", id).Scan(&next).Error
	return next, inventoryrepo.Error(err)
}
