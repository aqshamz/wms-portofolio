package inventory

import (
	"context"
	"strings"
	dto "wms-api/dto/inventory"
	model "wms-api/models/inventory"
	repository "wms-api/repository/inventory"
)

func (s *Service) CreateLot(ctx context.Context, q dto.CreateLotRequest, actor string) (dto.LotResponse, error) {
	if !uuid(q.OwnerID) || !uuid(q.ItemID) || !uuid(actor) {
		return dto.LotResponse{}, invalid("owner_id, item_id and authenticated actor must be UUIDs")
	}
	q.OwnerID = strings.ToLower(q.OwnerID)
	q.ItemID = strings.ToLower(q.ItemID)
	number, err := normalizeNumber(q.LotNumber, 100)
	if err != nil {
		return dto.LotResponse{}, err
	}
	manufactured, err := date(q.ManufactureDate)
	if err != nil {
		return dto.LotResponse{}, err
	}
	expires, err := date(q.ExpiryDate)
	if err != nil {
		return dto.LotResponse{}, err
	}
	if manufactured != nil && expires != nil && expires.Before(*manufactured) {
		return dto.LotResponse{}, invalid("expiry_date cannot precede manufacture_date")
	}
	if q.QualityStatusID != nil {
		if !uuid(*q.QualityStatusID) {
			return dto.LotResponse{}, invalid("quality_status_id must be a UUID")
		}
		id := strings.ToLower(*q.QualityStatusID)
		q.QualityStatusID = &id
	}
	id, err := newID("LOT")
	if err != nil {
		return dto.LotResponse{}, err
	}
	v := model.InventoryLot{ID: id, OwnerID: q.OwnerID, ItemID: q.ItemID, LotNumber: number, ManufactureDate: manufactured, ExpiryDate: expires, QualityStatusID: q.QualityStatusID, CreatedBy: &actor}
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories) error {
		item, e := requireItem(ctx, r, q.OwnerID, q.ItemID)
		if e != nil {
			return e
		}
		if !item.LotControlled {
			return invalid("item is not lot-controlled")
		}
		if q.QualityStatusID != nil {
			quality, e := r.Catalog.QualityStatus.GetShared(ctx, *q.QualityStatusID)
			if e != nil {
				return reference(e, "quality status")
			}
			if !quality.IsActive {
				return invalid("quality status is inactive")
			}
		}
		return r.Lot.Create(ctx, &v)
	})
	return mapLot(v), err
}
