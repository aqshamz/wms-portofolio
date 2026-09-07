package inventory

import (
	"context"
	"strings"
	dto "wms-api/dto/inventory"
	model "wms-api/models/inventory"
	repository "wms-api/repository/inventory"
)

func sameLocation(a, b *string) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
}
func (s *Service) CreateHandlingUnit(ctx context.Context, q dto.CreateHandlingUnitRequest, actor string) (dto.HandlingUnitResponse, error) {
	if !uuid(q.OwnerID) || !uuid(q.WarehouseID) || !uuid(q.HandlingUnitTypeID) || !uuid(actor) {
		return dto.HandlingUnitResponse{}, invalid("owner, warehouse, handling unit type and authenticated actor must be UUIDs")
	}
	q.OwnerID = strings.ToLower(q.OwnerID)
	q.WarehouseID = strings.ToLower(q.WarehouseID)
	q.HandlingUnitTypeID = strings.ToLower(q.HandlingUnitTypeID)
	if q.CurrentLocationID != nil {
		if !uuid(*q.CurrentLocationID) {
			return dto.HandlingUnitResponse{}, invalid("current_location_id must be a UUID")
		}
		id := strings.ToLower(*q.CurrentLocationID)
		q.CurrentLocationID = &id
	}
	if q.ParentHandlingUnitID != nil && !identityID(*q.ParentHandlingUnitID, 120) {
		return dto.HandlingUnitResponse{}, invalid("invalid parent_handling_unit_id")
	}
	barcode, err := normalizeNumber(q.Barcode, 120)
	if err != nil {
		return dto.HandlingUnitResponse{}, err
	}
	id, err := newID("HU")
	if err != nil {
		return dto.HandlingUnitResponse{}, err
	}
	v := model.HandlingUnit{ID: id, OwnerID: q.OwnerID, WarehouseID: q.WarehouseID, HandlingUnitTypeID: q.HandlingUnitTypeID, CurrentLocationID: q.CurrentLocationID, ParentHandlingUnitID: q.ParentHandlingUnitID, Barcode: barcode, CreatedBy: &actor}
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories) error {
		owner, e := r.Owner.GetShared(ctx, q.OwnerID)
		if e != nil {
			return reference(e, "owner")
		}
		if !owner.IsActive {
			return invalid("owner is inactive")
		}
		warehouse, e := r.Warehouse.GetShared(ctx, q.WarehouseID)
		if e != nil {
			return reference(e, "warehouse")
		}
		if !warehouse.IsActive {
			return invalid("warehouse is inactive")
		}
		operator, e := r.Owner.GetShared(ctx, warehouse.OperatorID)
		if e != nil {
			return reference(e, "operator")
		}
		if !operator.IsActive {
			return invalid("warehouse operator is inactive")
		}
		assignment, e := r.WarehouseOwner.GetShared(ctx, q.WarehouseID, q.OwnerID)
		if e != nil {
			return reference(e, "warehouse-owner assignment")
		}
		if !assignment.IsActive {
			return invalid("warehouse-owner assignment is inactive")
		}
		kind, e := r.Catalog.HandlingUnitType.GetShared(ctx, q.HandlingUnitTypeID)
		if e != nil {
			return reference(e, "handling unit type")
		}
		if !kind.IsActive {
			return invalid("handling unit type is inactive")
		}
		// Creation only: an existing parent cannot point at this new identity.
		// Reject corrupt legacy cycles/deep trees and closed ancestors as well.
		next := v.ParentHandlingUnitID
		visited := map[string]bool{v.ID: true}
		for depth := 0; next != nil; depth++ {
			if depth >= 64 || visited[*next] {
				return invalid("parent hierarchy contains a cycle or exceeds 64 ancestors")
			}
			visited[*next] = true
			parent, e := r.HandlingUnit.GetShared(ctx, *next)
			if e != nil {
				return reference(e, "parent handling unit")
			}
			if parent.OwnerID != v.OwnerID || parent.WarehouseID != v.WarehouseID {
				return invalid("parent must have the same owner and warehouse")
			}
			if parent.IsClosed {
				return invalid("parent or ancestor handling unit is closed")
			}
			if depth == 0 && v.CurrentLocationID == nil {
				v.CurrentLocationID = parent.CurrentLocationID
			}
			if !sameLocation(v.CurrentLocationID, parent.CurrentLocationID) {
				return invalid("child and all ancestors must share the same current location")
			}
			next = parent.ParentHandlingUnitID
		}
		if v.CurrentLocationID != nil {
			location, e := r.Location.GetShared(ctx, *v.CurrentLocationID)
			if e != nil {
				return reference(e, "location")
			}
			if location.WarehouseID != v.WarehouseID {
				return invalid("location does not belong to warehouse")
			}
			if !location.IsActive || location.IsLocked {
				return invalid("location is inactive or locked")
			}
			zone, e := r.Zone.GetShared(ctx, location.ZoneID)
			if e != nil {
				return reference(e, "location zone")
			}
			if !zone.IsActive {
				return invalid("location zone is inactive")
			}
			locationType, e := r.LocationType.GetShared(ctx, location.LocationTypeID)
			if e != nil {
				return reference(e, "location type")
			}
			if !locationType.IsActive {
				return invalid("location type is inactive")
			}
		}
		return r.HandlingUnit.Create(ctx, &v)
	})
	return mapHandlingUnit(v), err
}
