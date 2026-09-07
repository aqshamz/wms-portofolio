package inventory

import (
	"context"
	"strings"
	dto "wms-api/dto/inventory"
	model "wms-api/models/inventory"
	repository "wms-api/repository/inventory"
)

func (s *Service) CreateSerial(ctx context.Context, q dto.CreateSerialRequest, actor string) (dto.SerialResponse, error) {
	if !uuid(q.OwnerID) || !uuid(q.ItemID) || !uuid(actor) {
		return dto.SerialResponse{}, invalid("owner_id, item_id and authenticated actor must be UUIDs")
	}
	q.OwnerID = strings.ToLower(q.OwnerID)
	q.ItemID = strings.ToLower(q.ItemID)
	number, err := normalizeNumber(q.SerialNo, 120)
	if err != nil {
		return dto.SerialResponse{}, err
	}
	id, err := newID("SER")
	if err != nil {
		return dto.SerialResponse{}, err
	}
	v := model.SerialNumber{ID: id, OwnerID: q.OwnerID, ItemID: q.ItemID, SerialNo: number, CreatedBy: &actor}
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories) error {
		item, e := requireItem(ctx, r, q.OwnerID, q.ItemID)
		if e != nil {
			return e
		}
		if !item.SerialControlled {
			return invalid("item is not serial-controlled")
		}
		return r.Serial.Create(ctx, &v)
	})
	return mapSerial(v), err
}
