package outbound

import (
	"context"
	"math/big"
	"strings"

	mastermodel "wms-api/models/master"
	model "wms-api/models/outbound"
)

// Document quantities use the chosen order UOM; all warehouse execution uses
// the frozen base quantity. Never round a conversion into a different quantity.
func orderQuantitySnapshot(value, conversion string, serialized bool) (string, string, string, error) {
	qty, number, err := quantity(value, "ordered_qty", false)
	if err != nil {
		return "", "", "", err
	}
	factor, conversionNumber, err := quantity(conversion, "uom_conversion_to_base", false)
	if err != nil {
		return "", "", "", err
	}
	product := new(big.Rat).Mul(number, conversionNumber)
	base, exact, err := quantity(product.FloatString(6), "ordered_base_qty", false)
	if err != nil || exact.Cmp(product) != 0 {
		return "", "", "", invalid("converted base quantity exceeds numeric(20,6) precision")
	}
	if serialized && !exact.IsInt() {
		return "", "", "", invalid("serialized order quantities must convert to whole base units")
	}
	return qty, factor, base, nil
}

func validateOrderSnapshot(line model.OutboundOrderLine, item mastermodel.Item) error {
	if !item.IsActive || line.BaseUOMID != item.BaseUOMID {
		return state("item base UOM no longer matches the outbound quantity snapshot")
	}
	_, _, base, err := orderQuantitySnapshot(line.OrderedQty, line.UOMConversionToBase, item.SerialControlled)
	if err != nil {
		return err
	}
	_, stored, err := quantity(line.OrderedBaseQty, "stored base quantity", false)
	if err != nil || stored.Cmp(mustRat(base)) != 0 {
		return state("stored outbound quantity snapshot is inconsistent")
	}
	if line.UOMID == line.BaseUOMID && mustRat(line.UOMConversionToBase).Cmp(big.NewRat(1, 1)) != 0 {
		return state("base UOM conversion must equal one")
	}
	return nil
}

func (s *Service) orderUnit(ctx context.Context, item mastermodel.Item, requested *string, existing *model.OutboundOrderLine) (string, string, error) {
	uomID := item.BaseUOMID
	if existing != nil {
		uomID = existing.UOMID
	}
	if requested != nil {
		uomID = strings.ToLower(strings.TrimSpace(*requested))
	}
	if !validUUID(uomID) {
		return "", "", invalid("uom_id must be a UUID")
	}
	if existing != nil {
		if existing.BaseUOMID != item.BaseUOMID {
			return "", "", state("item base UOM no longer matches the outbound quantity snapshot")
		}
		if existing.UOMID == uomID {
			return uomID, existing.UOMConversionToBase, nil
		}
	}
	// Base-unit orders remain compatible with existing API clients. Alternate
	// packaging must have an active item conversion, independently of which
	// units the warehouse uses for picking or receiving.
	unit, err := s.repositories.Master.Catalog.UOM.Get(ctx, uomID)
	if err != nil || !unit.IsActive {
		return "", "", invalid("order UOM must be active")
	}
	if uomID == item.BaseUOMID {
		return uomID, "1.000000", nil
	}
	mapping, err := s.repositories.Master.Catalog.ItemUOM.ForUnit(ctx, item.ID, uomID)
	if err != nil || !mapping.IsActive {
		return "", "", invalid("order UOM does not have an active conversion for the item")
	}
	return uomID, mapping.ConversionToBase, nil
}
